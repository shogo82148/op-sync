package opsync

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Secrets map[string]map[string]any `yaml:"secrets"`
}

func ParseConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("opsync: failed to read %q: %w", filename, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("opsync: failed to parse %q: %w", filename, err)
	}
	return &config, nil
}
