package opsync

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Secrets map[string]*SyncConfig `yaml:"secrets"`
}

type SyncConfig struct {
	Type SyncType `yaml:"type"`

	// for SyncTypeTemplate
	Output   string `yaml:"output"`   // output file path
	Template string `yaml:"template"` // the template
}

type SyncType string

const (
	// SyncTypeTemplate is the type for injecting secrets into a file.
	SyncTypeTemplate SyncType = "template"
)

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
