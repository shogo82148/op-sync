package opsync

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v56/github"
	"github.com/shogo82148/op-sync/internal/gh"
	"github.com/shogo82148/op-sync/internal/op"
	"golang.org/x/crypto/nacl/box"
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

	owner, repo, ok := strings.Cut(cfg.Repository, "/")
	if !ok {
		return nil, fmt.Errorf("opsync: invalid repository %q", cfg.Repository)
	}
	secret, err := b.gh.GetRepoSecret(ctx, owner, repo, cfg.Name)

	var ghErr *github.ErrorResponse
	if errors.As(err, &ghErr) {
		if ghErr.Response.StatusCode == http.StatusNotFound {
			return b.newPlan(ctx, cfg, false)
		}
	}
	if err != nil {
		return nil, err
	}

	// TODO: check the secret is up-to-date
	_ = secret

	return b.newPlan(ctx, cfg, true)
}

func (b *GitHubBackend) newPlan(ctx context.Context, cfg *SyncConfig, overwrite bool) (Plan, error) {
	owner, repo, ok := strings.Cut(cfg.Repository, "/")
	if !ok {
		return nil, fmt.Errorf("opsync: invalid repository %q", cfg.Repository)
	}

	pubKey, err := b.gh.GetRepoPublicKey(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	decodedPubKey, err := base64.StdEncoding.DecodeString(pubKey.GetKey())
	if err != nil {
		return nil, err
	}
	if len(decodedPubKey) < 32 {
		return nil, fmt.Errorf("opsync: invalid public key")
	}

	var peersPubKey [32]byte
	copy(peersPubKey[:], decodedPubKey)

	secret, err := b.op.Read(ctx, cfg.Source)
	if err != nil {
		return nil, err
	}

	encryptedSecret, err := box.SealAnonymous(nil, secret[:], &peersPubKey, nil)
	if err != nil {
		return nil, err
	}

	return &GitHubPlan{
		backend:         b,
		owner:           owner,
		repo:            repo,
		name:            cfg.Name,
		keyID:           pubKey.GetKeyID(),
		encryptedSecret: encryptedSecret,
		overwrite:       overwrite,
	}, nil
}

type GitHubPlan struct {
	backend         *GitHubBackend
	owner           string
	repo            string
	name            string
	keyID           string
	encryptedSecret []byte
	overwrite       bool
}

func (p *GitHubPlan) Preview() string {
	if p.overwrite {
		return fmt.Sprintf("secret %q in %s/%s will be updated", p.name, p.owner, p.repo)
	}
	return fmt.Sprintf("secret %q in %s/%s will be created", p.name, p.owner, p.repo)
}

func (p *GitHubPlan) Apply(ctx context.Context) error {
	secret := base64.StdEncoding.EncodeToString(p.encryptedSecret)
	eSecret := &github.EncryptedSecret{
		Name:           p.name,
		KeyID:          p.keyID,
		EncryptedValue: secret,
	}
	return p.backend.gh.CreateOrUpdateRepoSecret(ctx, p.owner, p.repo, eSecret)
}
