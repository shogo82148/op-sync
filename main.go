package main

import (
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/goccy/go-yaml"
)

func main() {
	data, err := os.ReadFile(".op-sync.yml")
	if err != nil {
		panic(err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		panic(err)
	}

	if err := run(&config); err != nil {
		panic(err)
	}
}

type Config struct {
	Secrets []*SyncConfig `yaml:"secrets"`
}

type SyncConfig struct {
	Type SyncType `yaml:"type"`

	// for SyncTypeInject
	Output   string `yaml:"output"`   // output file path
	Template string `yaml:"template"` // the template
}

type SyncType string

const (
	// SyncTypeInject is the type for injecting secrets into a file.
	SyncTypeInject SyncType = "inject"
)

func run(cfg *Config) error {
	for _, secret := range cfg.Secrets {
		switch secret.Type {
		case SyncTypeInject:
			if err := runInject(context.TODO(), secret); err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

func runInject(ctx context.Context, cfg *SyncConfig) error {
	cmd := exec.CommandContext(ctx, "op", "inject", "--force", "--out-file", cfg.Output)
	cmd.Stdin = strings.NewReader(cfg.Template)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
