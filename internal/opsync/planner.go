package opsync

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/shogo82148/op-sync/internal/backends"
	"github.com/shogo82148/op-sync/internal/backends/awsssm"
	"github.com/shogo82148/op-sync/internal/backends/github"
	"github.com/shogo82148/op-sync/internal/backends/template"
	"github.com/shogo82148/op-sync/internal/maputils"
	"github.com/shogo82148/op-sync/internal/services/awssts"
	"github.com/shogo82148/op-sync/internal/services/gh"
	"github.com/shogo82148/op-sync/internal/services/op"
)

type Planner struct {
	cfg      *PlannerOptions
	backends map[string]backends.Backend
}

type PlannerOptions struct {
	Config      *Config
	OnePassword *op.Service
	GitHub      *gh.Service
	AWSSTS      *awssts.Service
}

func NewPlanner(cfg *PlannerOptions) *Planner {
	return &Planner{
		cfg: cfg,
		backends: map[string]backends.Backend{
			"template": template.New(&template.Options{
				Injector: cfg.OnePassword,
			}),
			"github": github.New(&github.Options{
				OnePasswordItemGetter: cfg.OnePassword,
				OnePasswordReader:     cfg.OnePassword,

				GitHubRepoGetter:                cfg.GitHub,
				GitHubRepoSecretGetter:          cfg.GitHub,
				GitHubRepoSecretCreator:         cfg.GitHub,
				GitHubRepoPublicKeyGetter:       cfg.GitHub,
				GitHubEnvSecretGetter:           cfg.GitHub,
				GitHubEnvSecretCreator:          cfg.GitHub,
				GitHubEnvPublicKeyGetter:        cfg.GitHub,
				GitHubOrgSecretGetter:           cfg.GitHub,
				GitHubOrgSecretCreator:          cfg.GitHub,
				GitHubOrgPublicKeyGetter:        cfg.GitHub,
				GitHubReposIDForOrgSecretLister: cfg.GitHub,
			}),
			"aws-ssm": awsssm.New(&awsssm.Options{
				OnePasswordReader: cfg.OnePassword,

				STSCallerIdentityGetter: cfg.AWSSTS,
			}),
		},
	}
}

func (p *Planner) Plan(ctx context.Context) ([]backends.Plan, error) {
	// check 1password cli is available.
	userInfo, err := p.cfg.OnePassword.WhoAmI(ctx)
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
