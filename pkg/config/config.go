package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/lichuan0620/taliban/pkg/model"
	"github.com/pkg/errors"
	prommodel "github.com/prometheus/common/model"
	"gopkg.in/yaml.v2"
)

// Load parses the YAML input s into a Config.
func Load(in []byte) (*Config, error) {
	cfg := &Config{}
	*cfg = DefaultConfig
	err := yaml.UnmarshalStrict(in, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// LoadFile parses the given YAML file into a Config.
func LoadFile(filename string) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return &DefaultConfig, nil
		}
		return nil, err
	}
	cfg, err := Load(content)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing YAML file %s", filename)
	}
	return cfg, nil
}

var (
	DefaultConfig = Config{
		Factories: []FactoryConfig{
			DefaultFactoryConfig,
		},
	}
	DefaultFactoryConfig = FactoryConfig{
		ExpositionPath:   "/metrics",
		ExpositionFormat: model.ExpositionFormatPrometheus,
		Vectors: []VectorConfig{
			DefaultVectorConfig,
		},
	}
	DefaultSampleGeneratorConfig = SampleGeneratorConfig{
		Distribution: model.DistributionRandom,
		Interval:     prommodel.Duration(time.Second),
		Precision:    3,
	}
	DefaultVectorConfig = VectorConfig{
		Type:                  model.MetricTypeGauge,
		LabelCount:            3,
		LabelCardinality:      10,
		SampleGeneratorConfig: DefaultSampleGeneratorConfig,
	}
)

type Config struct {
	Factories []FactoryConfig `yaml:"factories"`
}

// String implements fmt.Stringer interface.
func (c *Config) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("<error creating config string: %s>", err)
	}
	return string(b)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultConfig
	type plain Config
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	return nil
}

type FactoryConfig struct {
	ExpositionPath    string                 `yaml:"exposition_path"`
	ExpositionFormat  model.ExpositionFormat `yaml:"exposition_format,omitempty"`
	InstrumentHandler bool                   `yaml:"instrument_handler,omitempty"`
	Vectors           []VectorConfig         `yaml:"vectors,omitempty"`
}

// String implements fmt.Stringer interface.
func (c *FactoryConfig) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("<error creating metrics string: %s>", err)
	}
	return string(b)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *FactoryConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultFactoryConfig
	type plain FactoryConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	return nil
}

type SampleGeneratorConfig struct {
	Distribution model.SampleDistribution `yaml:"distribution"`
	Interval     prommodel.Duration       `yaml:"sample_interval"`
	Precision    int                      `yaml:"sample_precision,omitempty"`
	Max          string                   `yaml:"sample_max,omitempty"`
	Min          string                   `yaml:"sample_min,omitempty"`

	// StdDev for normal distribution
	StdDev string `yaml:"sample_std_dev,omitempty"`
	// Mean for normal distribution
	Mean string `yaml:"sample_mean,omitempty"`

	// RateParameter for exponential distribution
	RateParameter string `yaml:"rate_parameter,omitempty"`
}

// String implements fmt.Stringer interface.
func (c *SampleGeneratorConfig) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("<error creating metrics string: %s>", err)
	}
	return string(b)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *SampleGeneratorConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultSampleGeneratorConfig
	type plain SampleGeneratorConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	return nil
}

type VectorConfig struct {
	Type                  model.MetricType      `yaml:"type"`
	NamePrefix            string                `yaml:"name_prefix,omitempty"`
	Buckets               []string              `yaml:"buckets,omitempty"`
	LabelCount            int                   `yaml:"label_count"`
	LabelCardinality      int                   `yaml:"label_cardinality"`
	SampleGeneratorConfig SampleGeneratorConfig `yaml:",inline"`
}

// String implements fmt.Stringer interface.
func (c *VectorConfig) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Sprintf("<error creating metrics string: %s>", err)
	}
	return string(b)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
func (c *VectorConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	*c = DefaultVectorConfig
	type plain VectorConfig
	if err := unmarshal((*plain)(c)); err != nil {
		return err
	}
	return nil
}
