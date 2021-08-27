package metrics

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/lichuan0620/tailiban/pkg/config"
	"github.com/lichuan0620/tailiban/pkg/model"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

type Factory interface {
	Handler() http.Handler
	Run(<-chan struct{})
}

type runner func(<-chan struct{})

type factory struct {
	runners []runner
	handler http.Handler
}

func NewFactory(cfg *config.FactoryConfig) (Factory, error) {
	if cfg == nil {
		cfg = &config.DefaultFactoryConfig
	}
	registry := prometheus.NewRegistry()
	var enableOpenMetrics bool
	switch cfg.ExpositionFormat {
	case model.ExpositionFormatOpenMetrics:
		enableOpenMetrics = true
	case model.ExpositionFormatPrometheus:
	default:
		return nil, errors.Errorf("unknown exposition format \"%s\"", cfg.ExpositionFormat)
	}
	runners := make([]runner, len(cfg.Vectors))
	for i := range cfg.Vectors {
		log.Infof("constructing vector (%d/%d)", i, len(cfg.Vectors))
		collector, r, err := buildVector(&cfg.Vectors[i])
		if err != nil {
			return nil, errors.Wrap(err, "invalid vector")
		}
		if err = registry.Register(collector); err != nil {
			return nil, errors.Wrap(err, "register collector")
		}
		runners[i] = r
	}
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics: enableOpenMetrics,
	})
	if cfg.InstrumentHandler {
		promhttp.InstrumentMetricHandler(registry, handler)
	}
	return &factory{
		handler: handler,
		runners: runners,
	}, nil
}

func (f *factory) Handler() http.Handler {
	return f.handler
}

func (f *factory) Run(stopCh <-chan struct{}) {
	for i := range f.runners {
		r := f.runners[i]
		go r(stopCh)
	}
}

func buildVector(cfg *config.VectorConfig) (prometheus.Collector, runner, error) {
	labelNames, ok := GeneratorRandomNames(cfg.LabelCount)
	if !ok {
		return nil, nil, errors.New("failed to generate label names (try to reduce label count)")
	}
	labelValues := make([][]string, cfg.LabelCount)
	for i := range labelValues {
		if labelValues[i], ok = GeneratorRandomNames(cfg.LabelCardinality); !ok {
			return nil, nil, errors.Errorf("failed to generate label values (try to reduce label cardinality)")
		}
	}
	generator, err := NewSampleGenerator(&cfg.SampleGeneratorConfig)
	if err != nil {
		return nil, nil, err
	}
	var handle func(labels prometheus.Labels)
	name := namesgenerator.GetRandomName(0)
	if len(cfg.NamePrefix) > 0 {
		name = strings.TrimSuffix(cfg.NamePrefix, "_") + "_" + name
	}
	constLabels := map[string]string{
		"type":              string(cfg.Type),
		"precision":         strconv.Itoa(cfg.SampleGeneratorConfig.Precision),
		"label_count":       strconv.Itoa(cfg.LabelCount),
		"label_cardinality": strconv.Itoa(cfg.LabelCardinality),
	}
	var collector prometheus.Collector
	switch cfg.Type {
	case model.MetricTypeGauge:
		vec := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name:        name,
				Help:        "Arbitrarily-generated gauge metrics",
				ConstLabels: constLabels,
			}, labelNames,
		)
		handle = func(labels prometheus.Labels) {
			vec.With(labels).Set(generator.Get())
		}
		collector = vec
	case model.MetricTypeCounter:
		name += "_total"
		vec := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:        name,
				Help:        "Arbitrarily-generated counter metrics",
				ConstLabels: constLabels,
			}, labelNames,
		)
		handle = func(labels prometheus.Labels) {
			vec.With(labels).Add(generator.Get())
		}
		collector = vec
	case model.MetricTypeSummary:
		vec := prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:        name,
				Help:        "Arbitrarily-generated summary metrics",
				ConstLabels: constLabels,
			}, labelNames,
		)
		handle = func(labels prometheus.Labels) {
			vec.With(labels).Observe(generator.Get())
		}
		collector = vec
	case model.MetricTypeHistogram:
		buckets := prometheus.DefBuckets
		if len(cfg.Buckets) > 0 {
			buckets = make([]float64, len(cfg.Buckets))
			for i := range cfg.Buckets {
				le, err := strconv.ParseFloat(cfg.Buckets[i], 64)
				if err != nil {
					return nil, nil, errors.Wrap(err, "invalid bucket limit")
				}
				buckets[i] = le
			}
			sort.Float64s(buckets)
		}
		vec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:        name,
			Help:        "Arbitrarily-generated histogram metrics",
			Buckets:     buckets,
			ConstLabels: constLabels,
		}, labelNames)
		handle = func(labels prometheus.Labels) {
			vec.With(labels).Observe(generator.Get())
		}
		collector = vec
	default:
		return nil, nil, errors.Errorf("unknown metric type \"%s\"", cfg.Type)
	}
	interval := time.Duration(cfg.SampleGeneratorConfig.Interval)
	log.WithFields(log.Fields{
		"type":        cfg.Type,
		"name":        name,
		"labels":      labelNames,
		"cardinality": cfg.LabelCardinality,
		"precision":   cfg.SampleGeneratorConfig.Precision,
	}).Info("vector constructed")
	return collector, func(stopCh <-chan struct{}) {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				generateSamples(handle, make(prometheus.Labels), labelNames, labelValues)
			}
		}
	}, nil
}

func generateSamples(
	handler func(prometheus.Labels),
	labels prometheus.Labels,
	labelNames []string,
	labelValues [][]string,
) {
	if len(labelNames) == 0 {
		handler(labels)
		return
	}
	labelName := labelNames[0]
	for _, value := range labelValues[0] {
		labels[labelName] = value
		generateSamples(handler, labels, labelNames[1:], labelValues[1:])
		delete(labels, labelName)
	}
}
