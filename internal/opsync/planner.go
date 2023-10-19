package opsync

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/shogo82148/op-sync/internal/backends"
	"github.com/shogo82148/op-sync/internal/backends/template"
	"github.com/shogo82148/op-sync/internal/maputils"
	"github.com/shogo82148/op-sync/internal/op"
)

type Planner struct {
	cgf      *Config
	svc      *op.Service
	backends map[string]backends.Backend
}

func NewPlanner(cfg *Config, svc *op.Service) *Planner {
	return &Planner{
		cgf: cfg,
		svc: svc,
		backends: map[string]backends.Backend{
			"template": template.New(&template.Options{
				Injector: svc,
			}),
		},
	}
}

func (p *Planner) Plan(ctx context.Context) ([]backends.Plan, error) {
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
	plans := make([]backends.Plan, 0, len(keys))
	for _, key := range keys {
		slog.InfoContext(ctx, "planning", slog.String("key", key))
		cfg := p.cgf.Secrets[key]
		c := new(maputils.Context)
		typ := maputils.Must[string](c, cfg, "type")
		if err := c.Err(); err != nil {
			return nil, err
		}

		backend, ok := p.backends[typ]
		if !ok {
			return nil, fmt.Errorf("opsync: backend for type %q not found", typ)
		}
		plan, err := backend.Plan(ctx, cfg)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan...)
	}
	return plans, nil
}
