package gh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"sync"

	"github.com/google/go-github/v56/github"
	"github.com/shogo82148/op-sync/internal/services"
)

func command(ctx context.Context, name string, args ...string) *exec.Cmd {
	slog.DebugContext(ctx, "run GitHub cli", slog.String("name", name), slog.Any("args", args))
	return exec.CommandContext(ctx, name, args...)
}

func wrap(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return fmt.Errorf("failed to run gh command: %q: %w", string(exitErr.Stderr), err)
	}
	return fmt.Errorf("failed to run gh command: %w", err)
}

type Service struct {
	mu sync.Mutex
	c  *github.Client
}

func NewService() *Service {
	return &Service{}
}

// client returns an authorized GitHub client by GitHub CLI.
func (s *Service) client(ctx context.Context) (*github.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.c != nil {
		return s.c, nil
	}

	cmd := command(ctx, "gh", "auth", "token")
	out, err := cmd.Output()
	if err != nil {
		return nil, wrap(err)
	}
	token := string(bytes.TrimSpace(out))
	s.c = github.NewClient(nil).WithAuthToken(token)
	return s.c, nil
}

var _ services.GitHubUserGetter = (*Service)(nil)

// GetGitHubUser fetches the authenticated GitHub user.
func (s *Service) GetGitHubUser(ctx context.Context) (*github.User, error) {
	client, err := s.client(ctx)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "get the user")
	u, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}
	return u, nil
}

var _ services.GitHubRepoGetter = (*Service)(nil)

// GetGitHubRepo fetches the GitHub repository.
func (s *Service) GetGitHubRepo(ctx context.Context, owner, repo string) (*github.Repository, error) {
	client, err := s.client(ctx)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "get the repo", slog.String("owner", owner), slog.String("repo", repo))
	r, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return r, nil
}

var _ services.GitHubRepoSecretGetter = (*Service)(nil)

// GetGitHubRepoSecret gets a single repository secret without revealing its encrypted value.
func (s *Service) GetGitHubRepoSecret(ctx context.Context, owner, repo, name string) (*github.Secret, error) {
	client, err := s.client(ctx)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "get the repo secret", slog.String("owner", owner), slog.String("repo", repo), slog.String("name", name))
	secret, _, err := client.Actions.GetRepoSecret(ctx, owner, repo, name)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

var _ services.GitHubRepoSecretCreator = (*Service)(nil)

// CreateGitHubRepoSecret creates or updates a repository secret with an encrypted value.
func (s *Service) CreateGitHubRepoSecret(ctx context.Context, owner, repo string, secret *github.EncryptedSecret) error {
	client, err := s.client(ctx)
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, "create or update the repo secret", slog.String("owner", owner), slog.String("repo", repo), slog.String("name", secret.Name))
	_, err = client.Actions.CreateOrUpdateRepoSecret(ctx, owner, repo, secret)
	return err
}

var _ services.GitHubRepoPublicKeyGetter = (*Service)(nil)

// GetGitHubRepoPublicKey gets a public key that should be used for secret encryption.
func (s *Service) GetGitHubRepoPublicKey(ctx context.Context, owner, repo string) (*github.PublicKey, error) {
	client, err := s.client(ctx)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "get the repo public key", slog.String("owner", owner), slog.String("repo", repo))
	key, _, err := client.Actions.GetRepoPublicKey(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return key, nil
}

var _ services.GitHubEnvSecretGetter = (*Service)(nil)

// GetGitHubEvSecret gets a single environment secret without revealing its encrypted value.
func (s *Service) GetGitHubEnvSecret(ctx context.Context, repoID int, env, name string) (*github.Secret, error) {
	client, err := s.client(ctx)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "get the environment secret", slog.Int("repoID", repoID), slog.String("name", name))
	secret, _, err := client.Actions.GetEnvSecret(ctx, repoID, env, name)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

var _ services.GitHubEnvSecretCreator = (*Service)(nil)

// CreateGitHubEnvSecret creates or updates a single environment secret with an encrypted value.
func (s *Service) CreateGitHubEnvSecret(ctx context.Context, repoID int, env string, secret *github.EncryptedSecret) error {
	client, err := s.client(ctx)
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, "create or update the environment secret", slog.Int("repoID", repoID), slog.String("name", secret.Name))
	_, err = client.Actions.CreateOrUpdateEnvSecret(ctx, repoID, env, secret)
	return err
}

var _ services.GitHubEnvPublicKeyGetter = (*Service)(nil)

// GetGitHubEnvPublicKey gets a public key that should be used for secret encryption.
func (s *Service) GetGitHubEnvPublicKey(ctx context.Context, repoID int, env string) (*github.PublicKey, error) {
	client, err := s.client(ctx)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "get the environment public key", slog.Int("repoID", repoID), slog.String("env", env))
	key, _, err := client.Actions.GetEnvPublicKey(ctx, repoID, env)
	if err != nil {
		return nil, err
	}
	return key, nil
}

var _ services.GitHubOrgSecretGetter = (*Service)(nil)

// GetGitHubOrgSecret gets a single organization secret without revealing its encrypted value.
func (s *Service) GetGitHubOrgSecret(ctx context.Context, org, name string) (*github.Secret, error) {
	client, err := s.client(ctx)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "get the org secret", slog.String("org", org), slog.String("name", name))
	secret, _, err := client.Actions.GetOrgSecret(ctx, org, name)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

var _ services.GitHubOrgSecretCreator = (*Service)(nil)

// CreateGitHubOrgSecret creates or updates an organization secret with an encrypted value.
func (s *Service) CreateGitHubOrgSecret(ctx context.Context, org string, secret *github.EncryptedSecret) error {
	client, err := s.client(ctx)
	if err != nil {
		return err
	}

	slog.DebugContext(ctx, "create or update the org secret", slog.String("org", org), slog.String("name", secret.Name))
	_, err = client.Actions.CreateOrUpdateOrgSecret(ctx, org, secret)
	return err
}

var _ services.GitHubOrgPublicKeyGetter = (*Service)(nil)

// GetGitHubOrgPublicKey gets a public key that should be used for secret encryption.
func (s *Service) GetGitHubOrgPublicKey(ctx context.Context, org string) (*github.PublicKey, error) {
	client, err := s.client(ctx)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "get the org public key", slog.String("org", org))
	key, _, err := client.Actions.GetOrgPublicKey(ctx, org)
	if err != nil {
		return nil, err
	}
	return key, nil
}
