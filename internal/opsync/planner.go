package opsync

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/shogo82148/op-sync/internal/backends"
	"github.com/shogo82148/op-sync/internal/backends/template"
	"github.com/shogo82148/op-sync/internal/maputils"
	"github.com/shogo82148/op-sync/internal/services"
)

type Planner struct {
	cfg      *PlannerOptions
	backends map[string]backends.Backend
}

type PlannerOptions struct {
	Config *Config
	services.WhoAmIer
	services.Injector
}

func NewPlanner(cfg *PlannerOptions) *Planner {
	return &Planner{
		cfg: cfg,
		backends: map[string]backends.Backend{
			"template": template.New(&template.Options{
				Injector: cfg.Injector,
			}),
		},
	}
}

func (p *Planner) Plan(ctx context.Context) ([]backends.Plan, error) {
	// check 1password cli is available.
	userInfo, err := p.cfg.WhoAmI(ctx)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "1password user information", slog.String("url", userInfo.URL), slog.String("email", userInfo.Email))

	secrets := p.cfg.Config.Secrets
	keys := make([]string, 0, len(secrets))
	for key := range secrets {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	// do planning
	plans := make([]backends.Plan, 0, len(keys))
	for _, key := range keys {
		slog.DebugContext(ctx, "planning", slog.String("key", key))
		cfg := secrets[key]
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
