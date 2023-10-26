package opsync

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/shogo82148/op-sync/internal/backends"
	"github.com/shogo82148/op-sync/internal/backends/awssecretsmanager"
	"github.com/shogo82148/op-sync/internal/backends/awsssm"
	"github.com/shogo82148/op-sync/internal/backends/github"
	"github.com/shogo82148/op-sync/internal/backends/template"
	"github.com/shogo82148/op-sync/internal/maputils"
	svcsecretsmanager "github.com/shogo82148/op-sync/internal/services/awssecretsmanager"
	svcssm "github.com/shogo82148/op-sync/internal/services/awsssm"
	"github.com/shogo82148/op-sync/internal/services/awssts"
	"github.com/shogo82148/op-sync/internal/services/gh"
	"github.com/shogo82148/op-sync/internal/services/op"
)

type Planner struct {
	cfg      *PlannerOptions
	backends map[string]backends.Backend
}

type PlannerOptions struct {
	Config            *Config
	OnePassword       *op.Service
	GitHub            *gh.Service
	AWSSTS            *awssts.Service
	AWSSSM            *svcssm.Service
	AWSSecretsManager *svcsecretsmanager.Service
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

				SSMParameterGetter: cfg.AWSSSM,
				SSMParameterPutter: cfg.AWSSSM,
			}),
			"aws-secrets-manager": awssecretsmanager.New(&awssecretsmanager.Options{
				OnePasswordReader: cfg.OnePassword,

				STSCallerIdentityGetter: cfg.AWSSTS,

				SecretsManagerSecretCreator: cfg.AWSSecretsManager,
				SecretsManagerSecretGetter:  cfg.AWSSecretsManager,
				SecretsManagerSecretUpdater: cfg.AWSSecretsManager,
			}),
		},
	}
}

// check 1password cli is available.
func (p *Planner) checkIsOPAvailable(ctx context.Context) error {
	userInfo, err := p.cfg.OnePassword.WhoAmI(ctx)
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "1password user information", slog.String("url", userInfo.URL), slog.String("email", userInfo.Email))
	return nil
}

// Plan plans all secrets.
func (p *Planner) Plan(ctx context.Context) ([]backends.Plan, error) {
	s := p.cfg.Config.Secrets
	keys := make([]string, 0, len(s))
	for key := range s {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	return p.plan(ctx, keys)
}

// PlanWithSecrets plans the specified secrets.
func (p *Planner) PlanWithSecrets(ctx context.Context, secrets []string) ([]backends.Plan, error) {
	return p.plan(ctx, secrets)
}

// PlanWithType plans the specified type secrets.
func (p Planner) PlanWithType(ctx context.Context, type_ string) ([]backends.Plan, error) {
	if _, ok := p.backends[type_]; !ok {
		return nil, fmt.Errorf("opsync: backend for type %q not found", type_)
	}

	s := p.cfg.Config.Secrets
	keys := make([]string, 0, len(s))
	for key, cfg := range s {
		c := new(maputils.Context)
		typ := maputils.Must[string](c, cfg, "type")
		if err := c.Err(); err != nil {
			return nil, err
		}
		if typ == type_ {
			keys = append(keys, key)
		}
	}
	slices.Sort(keys)

	return p.plan(ctx, keys)
}

func (p *Planner) plan(ctx context.Context, secrets []string) ([]backends.Plan, error) {
	// check 1password cli is available.
	if err := p.checkIsOPAvailable(ctx); err != nil {
		return nil, err
	}

	// list the secrets
	s := p.cfg.Config.Secrets
	keys := make([]string, 0, len(secrets))
	unknown := make([]string, 0, len(secrets))
	for _, key := range secrets {
		if _, ok := s[key]; ok {
			keys = append(keys, key)
		} else {
			unknown = append(unknown, key)
		}
	}
	slices.Sort(keys)
	slices.Sort(unknown)
	if len(unknown) > 0 {
		return nil, fmt.Errorf("opsync: unknown secrets %q", unknown)
	}

	// do planning
	errs := []error{}
	plans := make([]backends.Plan, 0, len(keys))
	for _, key := range keys {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		slog.InfoContext(ctx, "planning", slog.String("key", key))
		cfg := s[key]
		c := new(maputils.Context)
		typ := maputils.Must[string](c, cfg, "type")
		if err := c.Err(); err != nil {
			return nil, err
		}

		backend, ok := p.backends[typ]
		if !ok {
			errs = append(errs, fmt.Errorf("opsync: backend for type %q not found", typ))
			continue
		}
		plan, err := backend.Plan(ctx, cfg)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		plans = append(plans, plan...)
	}
	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}
	return plans, nil
}
