package metrics

import (
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/lichuan0620/taliban/pkg/config"
	"github.com/lichuan0620/taliban/pkg/model"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type SampleGenerator interface {
	Get() float64
}

type sampleGenerator struct {
	lock    sync.Mutex
	cur     int
	samples []float64
}

func NewSampleGenerator(cfg *config.SampleGeneratorConfig) (SampleGenerator, error) {
	const samplePoolSize = 4096
	generateSampleFunc, err := buildGenerateSampleFunc(cfg)
	if err != nil {
		return nil, err
	}
	samples := make([]float64, samplePoolSize)
	for i := range samples {
		samples[i], _ = generateSampleFunc().Float64()
	}
	return &sampleGenerator{samples: samples}, nil
}

func (sg *sampleGenerator) Get() float64 {
	sg.lock.Lock()
	defer sg.lock.Unlock()
	sg.cur++
	if sg.cur >= len(sg.samples) {
		sg.cur = 0
	}
	return sg.samples[sg.cur]
}

func buildGenerateSampleFunc(cfg *config.SampleGeneratorConfig) (func() decimal.Decimal, error) {
	var err error
	min, max := decimal.NewFromFloat(math.MaxFloat64), decimal.NewFromFloat(-math.MaxFloat64)
	if len(cfg.Max) > 0 {
		if max, err = decimal.NewFromString(cfg.Max); err != nil {
			return nil, errors.Wrap(err, "invalid max")
		}
	}
	if len(cfg.Min) > 0 {
		if min, err = decimal.NewFromString(cfg.Min); err != nil {
			return nil, errors.Wrap(err, "invalid min")
		}
	}
	switch cfg.Distribution {
	case model.DistributionRandom:
		return func() decimal.Decimal {
			return decimal.NewFromFloat(rand.Float64()).Mul(max.Sub(min)).Add(min).Round(int32(cfg.Precision))
		}, nil
	case model.DistributionNormal:
		stdDev, err := decimal.NewFromString(cfg.StdDev)
		if err != nil {
			return nil, errors.Wrap(err, "invalid standard deviation")
		}
		mean, err := decimal.NewFromString(cfg.Mean)
		if err != nil {
			return nil, errors.Wrap(err, "invalid mean")
		}
		return func() decimal.Decimal {
			ret := decimal.NewFromFloat(rand.NormFloat64()).Mul(stdDev).Add(mean)
			if ret.LessThan(min) {
				ret = min
			} else if ret.GreaterThan(max) {
				ret = max
			}
			return ret.Round(int32(cfg.Precision))
		}, nil
	case model.DistributionExponential:
		rateParam, err := decimal.NewFromString(cfg.RateParameter)
		if err != nil {
			return nil, errors.Wrap(err, "invalid rate parameter")
		}
		return func() decimal.Decimal {
			ret := decimal.NewFromFloat(rand.ExpFloat64()).Div(rateParam)
			if ret.LessThan(min) {
				ret = min
			} else if ret.GreaterThan(max) {
				ret = max
			}
			return ret.Round(int32(cfg.Precision))
		}, nil
	default:
		return nil, errors.Errorf("unknown distribution \"%s\"", cfg.Distribution)
	}
}

func GeneratorRandomNames(count int) ([]string, bool) {
	const maxGenRetry = 100
	labels := make(map[string]struct{}, count)
	for i := 0; i < count; i++ {
		if ok := func() bool {
			for retry := 0; retry < maxGenRetry; retry++ {
				name := namesgenerator.GetRandomName(1)
				if _, exists := labels[name]; !exists {
					labels[name] = struct{}{}
					return true
				}
			}
			return false
		}(); !ok {
			return nil, false
		}
	}
	ret := make([]string, 0, count)
	for label := range labels {
		ret = append(ret, label)
	}
	sort.Strings(ret)
	return ret, true
}
