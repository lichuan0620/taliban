package model

type ExpositionFormat string

const (
	ExpositionFormatOpenMetrics ExpositionFormat = "openmetrics"
	ExpositionFormatPrometheus  ExpositionFormat = "prometheus"
)

type MetricType string

const (
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeCounter   MetricType = "counter"
	MetricTypeSummary   MetricType = "summary"
	MetricTypeHistogram MetricType = "histogram"
)

type SampleDistribution string

const (
	DistributionRandom      SampleDistribution = "random"
	DistributionNormal      SampleDistribution = "normal"
	DistributionExponential SampleDistribution = "exponential"
)
