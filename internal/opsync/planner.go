package opsync

import (
	"context"
	"errors"
	"log/slog"
	"slices"

	"github.com/shogo82148/op-sync/internal/op"
)

// errNotSupported is an error sentinel that the backend not supported the type.
var errNotSupported = errors.New("opsync: not supported")

// errSkip is an error sentinel that the backend skip the type.
var errSkip = errors.New("opsync: skip")

type Planner struct {
	cgf      *Config
	svc      *op.Service
	backends []Backend
}

func NewPlanner(cfg *Config, svc *op.Service) *Planner {
	return &Planner{
		cgf: cfg,
		svc: svc,
		backends: []Backend{
			NewTemplateBackend(svc),
			NewGitHubBackend(svc),
		},
	}
}

func (p *Planner) Plan(ctx context.Context) ([]Plan, error) {
	// check 1password cli is available.
	userInfo, err := p.svc.Whoami(ctx)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "1password user information", slog.String("url", userInfo.URL), slog.String("email", userInfo.Email))

	keys := make([]string, 0, len(p.cgf.Secrets))
	for key := range p.cgf.Secrets {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	// do planning
	plans := make([]Plan, 0, len(keys))
LOOP:
	for _, key := range keys {
		slog.InfoContext(ctx, "planning", slog.String("key", key))
		cfg := p.cgf.Secrets[key]
		for _, backend := range p.backends {
			plan, err := backend.Plan(ctx, cfg)
			if errors.Is(err, errNotSupported) {
				continue
			}
			if errors.Is(err, errSkip) {
				continue LOOP
			}
			if err != nil {
				return nil, err
			}
			plans = append(plans, plan)
			continue LOOP
		}
		slog.WarnContext(ctx, "no backend found", slog.String("key", key))
	}
	return plans, nil
}

type Backend interface {
	Plan(ctx context.Context, cfg *SyncConfig) (Plan, error)
}

type Plan interface {
	Preview() string
	Apply(ctx context.Context) error
}
