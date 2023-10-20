package github

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v56/github"
	"github.com/shogo82148/op-sync/internal/backends"
	"github.com/shogo82148/op-sync/internal/maputils"
	"github.com/shogo82148/op-sync/internal/services"
	"github.com/shogo82148/op-sync/internal/services/op"
	"golang.org/x/crypto/nacl/box"
)

var _ backends.Backend = (*Backend)(nil)

type Backend struct {
	opts *Options
}

type Options struct {
	services.OnePasswordItemGetter
	services.OnePasswordReader
	services.GitHubRepoGetter
	services.GitHubRepoSecretGetter
	services.GitHubRepoSecretCreator
	services.GitHubRepoPublicKeyGetter
	services.GitHubEnvSecretGetter
	services.GitHubEnvSecretCreator
	services.GitHubEnvPublicKeyGetter
	services.GitHubOrgSecretGetter
	services.GitHubOrgSecretCreator
	services.GitHubOrgPublicKeyGetter
}

func New(opts *Options) *Backend {
	return &Backend{opts: opts}
}

func (b *Backend) Plan(ctx context.Context, params map[string]any) ([]backends.Plan, error) {
	c := new(maputils.Context)
	organization, hasOrganization := maputils.Get[string](c, params, "organization")
	repository, hasRepository := maputils.Get[string](c, params, "repository")
	environment, hasEnvironment := maputils.Get[string](c, params, "environment")
	application, hasApplication := maputils.Get[string](c, params, "application")
	name := maputils.Must[string](c, params, "name")
	source := maputils.Must[string](c, params, "source")
	if err := c.Err(); err != nil {
		return nil, fmt.Errorf("github: validation failed: %w", err)
	}

	if hasOrganization && hasRepository {
		return nil, errors.New("github: both organization and repository are specified")
	}
	if hasEnvironment && hasApplication {
		return nil, errors.New("github: both environment and application are specified")
	}

	var app services.GitHubApplication
	switch application {
	case "actions":
		app = services.GitHubApplicationActions
	case "codespaces":
		app = services.GitHubApplicationCodespaces
	case "dependabot":
		app = services.GitHubApplicationDependabot
	default:
		if hasApplication {
			return nil, fmt.Errorf("github: unknown application %q", application)
		}
		app = services.GitHubApplicationActions
	}

	if hasOrganization {
		return b.planOrgSecret(ctx, organization, name, source)
	}

	if hasRepository {
		owner, repo, ok := strings.Cut(repository, "/")
		if !ok {
			return nil, fmt.Errorf("github: invalid repository name %q", repository)
		}
		if hasEnvironment {
			return b.planEnvSecret(ctx, owner, repo, environment, name, source)
		} else {
			return b.planRepoSecret(ctx, app, owner, repo, name, source)
		}
	}

	return []backends.Plan{}, nil
}

func isNotFound(err error) bool {
	var ghErr *github.ErrorResponse
	return errors.As(err, &ghErr) && ghErr.Response.StatusCode == http.StatusNotFound
}

func (b *Backend) planRepoSecret(ctx context.Context, app services.GitHubApplication, owner, repo, name, source string) ([]backends.Plan, error) {
	secret, err := b.opts.GetGitHubRepoSecret(ctx, app, owner, repo, name)
	if isNotFound(err) {
		// the secret is not found.
		// we should create it.
		return b.newPlanRepoSecret(ctx, owner, repo, name, source)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub repo secret: %w", err)
	}

	uri, err := op.ParseURI(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source: %w", err)
	}

	// check the secret is up-to-date
	opItem, err := b.opts.GetOnePasswordItem(ctx, uri.Vault, uri.Item)
	if err != nil {
		return nil, err
	}

	if secret.UpdatedAt.After(opItem.UpdatedAt) {
		// the secret is up-to-date.
		return []backends.Plan{}, nil
	}

	return b.newPlanRepoSecret(ctx, owner, repo, name, source)
}

func (b *Backend) newPlanRepoSecret(ctx context.Context, owner, repo, name, source string) ([]backends.Plan, error) {
	// get the public key
	key, err := b.opts.GetGitHubRepoPublicKey(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub repo public key: %w", err)
	}

	// get the secret from 1password
	secret, err := b.opts.ReadOnePassword(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret from 1password: %w", err)
	}

	// encrypt the secret
	encryptedSecret, err := encryptSecret(key.GetKey(), secret)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}

	return []backends.Plan{
		&PlanRepoSecret{
			backend:         b,
			owner:           owner,
			repo:            repo,
			name:            name,
			keyID:           key.GetKeyID(),
			encryptedSecret: encryptedSecret,
			overwrite:       false,
		},
	}, nil
}

func (b *Backend) planEnvSecret(ctx context.Context, owner, repo, env, name, source string) ([]backends.Plan, error) {
	ghRepo, err := b.opts.GetGitHubRepo(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub repo: %w", err)
	}
	secret, err := b.opts.GetGitHubEnvSecret(ctx, int(ghRepo.GetID()), env, name)
	if isNotFound(err) {
		// the secret is not found.
		// we should create it.
		return b.newPlanEnvSecret(ctx, ghRepo, env, name, source)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub repo secret: %w", err)
	}

	uri, err := op.ParseURI(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source: %w", err)
	}

	// check the secret is up-to-date
	opItem, err := b.opts.GetOnePasswordItem(ctx, uri.Vault, uri.Item)
	if err != nil {
		return nil, err
	}

	if secret.UpdatedAt.After(opItem.UpdatedAt) {
		// the secret is up-to-date.
		return []backends.Plan{}, nil
	}

	return b.newPlanEnvSecret(ctx, ghRepo, env, name, source)
}

func (b *Backend) newPlanEnvSecret(ctx context.Context, ghRepo *github.Repository, env, name, source string) ([]backends.Plan, error) {
	// get the public key
	key, err := b.opts.GetGitHubEnvPublicKey(ctx, int(ghRepo.GetID()), env)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub environment public key: %w", err)
	}

	// get the secret from 1password
	secret, err := b.opts.ReadOnePassword(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret from 1password: %w", err)
	}

	// encrypt the secret
	encryptedSecret, err := encryptSecret(key.GetKey(), secret)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}

	return []backends.Plan{
		&PlanEnvSecret{
			backend:         b,
			owner:           ghRepo.GetOwner().GetLogin(),
			repo:            ghRepo.GetName(),
			repoID:          ghRepo.GetID(),
			env:             env,
			name:            name,
			keyID:           key.GetKeyID(),
			encryptedSecret: encryptedSecret,
			overwrite:       false,
		},
	}, nil
}

func (b *Backend) planOrgSecret(ctx context.Context, organization, name, source string) ([]backends.Plan, error) {
	secret, err := b.opts.GetGitHubOrgSecret(ctx, organization, name)
	if isNotFound(err) {
		// the secret is not found.
		// we should create it.
		return b.newPlanOrgSecret(ctx, organization, name, source, false)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub org secret: %w", err)
	}

	uri, err := op.ParseURI(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source: %w", err)
	}

	// check the secret is up-to-date
	opItem, err := b.opts.GetOnePasswordItem(ctx, uri.Vault, uri.Item)
	if err != nil {
		return nil, err
	}

	if secret.UpdatedAt.After(opItem.UpdatedAt) {
		// the secret is up-to-date.
		return []backends.Plan{}, nil
	}

	return b.newPlanOrgSecret(ctx, organization, name, source, true)
}

func (b *Backend) newPlanOrgSecret(ctx context.Context, organization, name, source string, overwrite bool) ([]backends.Plan, error) {
	// get the public key
	key, err := b.opts.GetGitHubOrgPublicKey(ctx, organization)
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub org public key: %w", err)
	}

	// get the secret from 1password
	secret, err := b.opts.ReadOnePassword(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret from 1password: %w", err)
	}

	// encrypt the secret
	encryptedSecret, err := encryptSecret(key.GetKey(), secret)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}

	return []backends.Plan{
		&PlanOrgSecret{
			backend:         b,
			organization:    organization,
			name:            name,
			keyID:           key.GetKeyID(),
			encryptedSecret: encryptedSecret,
			overwrite:       overwrite,
		},
	}, nil
}

func encryptSecret(pubKey string, secret []byte) (string, error) {
	decodedPubKey, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		return "", err
	}
	if len(decodedPubKey) != 32 {
		return "", fmt.Errorf("github: invalid public key")
	}

	var peersPubKey [32]byte
	copy(peersPubKey[:], decodedPubKey)

	encrypted, err := box.SealAnonymous(nil, secret, &peersPubKey, nil)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

var _ backends.Plan = (*PlanRepoSecret)(nil)

type PlanRepoSecret struct {
	backend         *Backend
	app             services.GitHubApplication
	owner           string
	repo            string
	name            string
	keyID           string
	encryptedSecret string
	overwrite       bool
}

func (p *PlanRepoSecret) Preview() string {
	if p.overwrite {
		return fmt.Sprintf("secret %q in %s/%s will be updated", p.name, p.owner, p.repo)
	}
	return fmt.Sprintf("secret %q in %s/%s will be created", p.name, p.owner, p.repo)
}

func (p *PlanRepoSecret) Apply(ctx context.Context) error {
	eSecret := &github.EncryptedSecret{
		Name:           p.name,
		KeyID:          p.keyID,
		EncryptedValue: p.encryptedSecret,
	}
	return p.backend.opts.CreateGitHubRepoSecret(ctx, p.app, p.owner, p.repo, eSecret)
}

var _ backends.Plan = (*PlanEnvSecret)(nil)

type PlanEnvSecret struct {
	backend         *Backend
	owner           string
	repo            string
	repoID          int64
	env             string
	name            string
	keyID           string
	encryptedSecret string
	overwrite       bool
}

func (p *PlanEnvSecret) Preview() string {
	if p.overwrite {
		return fmt.Sprintf("secret %q in %s/%s environment %s will be updated", p.name, p.owner, p.repo, p.env)
	}
	return fmt.Sprintf("secret %q in %s/%s environment %s will be created", p.name, p.owner, p.repo, p.env)
}

func (p *PlanEnvSecret) Apply(ctx context.Context) error {
	eSecret := &github.EncryptedSecret{
		Name:           p.name,
		KeyID:          p.keyID,
		EncryptedValue: p.encryptedSecret,
	}
	return p.backend.opts.CreateGitHubEnvSecret(ctx, int(p.repoID), p.env, eSecret)
}

var _ backends.Plan = (*PlanOrgSecret)(nil)

type PlanOrgSecret struct {
	backend         *Backend
	organization    string
	name            string
	keyID           string
	encryptedSecret string
	overwrite       bool
}

func (p *PlanOrgSecret) Preview() string {
	if p.overwrite {
		return fmt.Sprintf("secret %q in organization %s will be updated", p.name, p.organization)
	}
	return fmt.Sprintf("secret %q in organization %s will be created", p.name, p.organization)
}

func (p *PlanOrgSecret) Apply(ctx context.Context) error {
	eSecret := &github.EncryptedSecret{
		Name:           p.name,
		KeyID:          p.keyID,
		EncryptedValue: p.encryptedSecret,
	}
	return p.backend.opts.CreateGitHubOrgSecret(ctx, p.organization, eSecret)
}
