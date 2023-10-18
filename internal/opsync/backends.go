package opsync

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/shogo82148/op-sync/internal/gh"
	"github.com/shogo82148/op-sync/internal/op"
)

var _ Backend = (*TemplateBackend)(nil)

// TemplateBackend is a backend for SyncTypeTemplate.
type TemplateBackend struct {
	svc *op.Service
}

func NewTemplateBackend(svc *op.Service) *TemplateBackend {
	return &TemplateBackend{svc: svc}
}

func (b *TemplateBackend) Plan(ctx context.Context, cfg *SyncConfig) (Plan, error) {
	if cfg.Type != SyncTypeTemplate {
		return nil, errNotSupported
	}
	return &TemplatePlan{
		backend:  b,
		output:   cfg.Output,
		template: cfg.Template,
	}, nil
}

type TemplatePlan struct {
	backend  *TemplateBackend
	output   string
	template string
}

func (p *TemplatePlan) Preview() string {
	return fmt.Sprintf("file %q will be created", p.output)
}

func (p *TemplatePlan) Apply(ctx context.Context) error {
	log.Printf("injecting secrets into %q", p.output)
	data, err := p.backend.svc.Inject(ctx, p.template)
	if err != nil {
		return err
	}
	return os.WriteFile(p.output, data, 0600)
}

var _ Backend = (*GitHubBackend)(nil)

// GitHubBackend is a backend for SyncTypeGitHub.
type GitHubBackend struct {
	op *op.Service
	gh *gh.Service
}

func NewGitHubBackend(svc *op.Service) *GitHubBackend {
	return &GitHubBackend{op: svc, gh: gh.NewService()}
}

func (b *GitHubBackend) Plan(ctx context.Context, cfg *SyncConfig) (Plan, error) {
	if cfg.Type != SyncTypeGitHub {
		return nil, errNotSupported
	}

	log.Println(b.gh.UserInfo(ctx))

	return &GitHubPlan{
		backend:   b,
		repo:      cfg.Repository,
		secret:    cfg.Name,
		overwrite: true,
	}, nil
}

type GitHubPlan struct {
	backend   *GitHubBackend
	repo      string
	secret    string
	overwrite bool
}

func (p *GitHubPlan) Preview() string {
	return ""
}

func (p *GitHubPlan) Apply(ctx context.Context) error {
	return nil
}
