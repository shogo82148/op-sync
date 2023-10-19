package github

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/go-github/v56/github"
	"github.com/shogo82148/op-sync/internal/backends"
	"github.com/shogo82148/op-sync/internal/services"
	"golang.org/x/crypto/nacl/box"
)

var _ backends.Backend = (*Backend)(nil)

type Backend struct {
	opts *Options
}

type Options struct {
	services.OnePasswordItemGetter
	services.OnePasswordReader
	services.GitHubRepoSecretGetter
	services.GitHubRepoSecretCreator
	services.GitHubRepoPublicKeyGetter
}

func New(opts *Options) *Backend {
	return &Backend{opts: opts}
}

func (b *Backend) Plan(ctx context.Context, params map[string]any) ([]backends.Plan, error) {
	repository := params["repository"].(string) // TODO: validation
	name := params["name"].(string)             // TODO: validation
	source := params["source"].(string)         // TODO: validation

	uri, err := url.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source: %w", err)
	}
	if uri.Scheme != "op" {
		return nil, fmt.Errorf("github: invalid source: %q", source)
	}

	owner, repo, ok := strings.Cut(repository, "/")
	if !ok {
		return nil, fmt.Errorf("github: invalid repository name %q", repository)
	}
	secret, err := b.opts.GetGitHubRepoSecret(ctx, owner, repo, name)

	var ghErr *github.ErrorResponse
	if errors.As(err, &ghErr) {
		if ghErr.Response.StatusCode == http.StatusNotFound {
			// the secret is not found.
			// we should create it.
			return b.planRepoSecret(ctx, owner, repo, name, source)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub repo secret: %w", err)
	}

	// check the secret is up-to-date
	path := strings.TrimPrefix(uri.Path, "/")
	item, _, _ := strings.Cut(path, "/")
	opItem, err := b.opts.GetOnePasswordItem(ctx, uri.Host, item)
	if err != nil {
		return nil, err
	}

	if secret.UpdatedAt.After(opItem.UpdatedAt) {
		// the secret is up-to-date.
		return []backends.Plan{}, nil
	}

	return b.planRepoSecret(ctx, owner, repo, name, source)
}

func (b *Backend) planRepoSecret(ctx context.Context, owner, repo, name, source string) ([]backends.Plan, error) {
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
	return p.backend.opts.CreateGitHubRepoSecret(ctx, p.owner, p.repo, eSecret)
}
