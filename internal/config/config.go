package config

import (
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

type PreCrawlConfig struct {
	BaseTargetURL      *string   `yaml:"base_target_url,omitempty"`
	DefaultSelector    *string   `yaml:"default_selector,omitempty"`
	DefaultWaitTimeout *string   `yaml:"default_wait_timeout,omitempty"`
	Transformers       *[]string `yaml:"transformers,omitempty"`
	WorkerCount        *int      `yaml:"worker_count,omitempty"`
}

var posibleTransformerTypes = []string{
	"ImageURLPruner",
	"ClassPruner",
	"StylePruner",
}

func LoadConfig(
	source []byte,
) (*PreCrawlConfig, error) {
	var config PreCrawlConfig
	err := yaml.Unmarshal(source, &config)
	if err != nil {
		return nil, err
	}

	for _, t := range *config.Transformers {
		if !slices.Contains(posibleTransformerTypes, t) {
			return nil, fmt.Errorf("invalid transformer type %s", t)
		}
	}

	return &config, nil
}
