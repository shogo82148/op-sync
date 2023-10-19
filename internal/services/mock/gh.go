package mock

import (
	"context"

	"github.com/google/go-github/v56/github"
	"github.com/shogo82148/op-sync/internal/services"
)

var _ services.GitHubUserGetter = GitHubUserGetter(nil)

// GitHubUserGetter fetches the authenticated GitHub user.
type GitHubUserGetter func(ctx context.Context) (*github.User, error)

func (f GitHubUserGetter) GetGitHubUser(ctx context.Context) (*github.User, error) {
	return f(ctx)
}

var _ services.GitHubRepoGetter = GitHubRepoGetter(nil)

// GitHubRepoGetter fetches the GitHub repository.
type GitHubRepoGetter func(ctx context.Context, owner, repo string) (*github.Repository, error)

func (f GitHubRepoGetter) GetGitHubRepo(ctx context.Context, owner, repo string) (*github.Repository, error) {
	return f(ctx, owner, repo)
}

var _ services.GitHubRepoSecretGetter = GitHubRepoSecretGetter(nil)

// GitHubRepoSecretGetter gets a single repository secret without revealing its encrypted value.
type GitHubRepoSecretGetter func(ctx context.Context, owner, repo, name string) (*github.Secret, error)

func (f GitHubRepoSecretGetter) GetGitHubRepoSecret(ctx context.Context, owner, repo, name string) (*github.Secret, error) {
	return f(ctx, owner, repo, name)
}

var _ services.GitHubRepoSecretCreator = GitHubRepoSecretCreator(nil)

// GitHubRepoSecretCreator creates or updates a repository secret with an encrypted value.
type GitHubRepoSecretCreator func(ctx context.Context, owner, repo string, secret *github.EncryptedSecret) error

func (f GitHubRepoSecretCreator) CreateGitHubRepoSecret(ctx context.Context, owner, repo string, secret *github.EncryptedSecret) error {
	return f(ctx, owner, repo, secret)
}

var _ services.GitHubRepoPublicKeyGetter = GitHubRepoPublicKeyGetter(nil)

// GitHubRepoPublicKeyGetter gets a public key that should be used for secret encryption.
type GitHubRepoPublicKeyGetter func(ctx context.Context, owner, repo string) (*github.PublicKey, error)

func (f GitHubRepoPublicKeyGetter) GetGitHubRepoPublicKey(ctx context.Context, owner, repo string) (*github.PublicKey, error) {
	return f(ctx, owner, repo)
}

var _ services.GitHubEnvSecretGetter = GitHubEnvSecretGetter(nil)

// GitHubEnvSecretGetter gets a single environment secret without revealing its encrypted value.
type GitHubEnvSecretGetter func(ctx context.Context, repoID int, env, name string) (*github.Secret, error)

func (f GitHubEnvSecretGetter) GetGitHubEnvSecret(ctx context.Context, repoID int, env, name string) (*github.Secret, error) {
	return f(ctx, repoID, env, name)
}

var _ services.GitHubEnvSecretCreator = GitHubEnvSecretCreator(nil)

// GitHubEnvSecretCreator creates or updates a single environment secret with an encrypted value.
type GitHubEnvSecretCreator func(ctx context.Context, repoID int, env string, secret *github.EncryptedSecret) error

func (f GitHubEnvSecretCreator) CreateGitHubEnvSecret(ctx context.Context, repoID int, env string, secret *github.EncryptedSecret) error {
	return f(ctx, repoID, env, secret)
}

var _ services.GitHubEnvPublicKeyGetter = GitHubEnvPublicKeyGetter(nil)

// GitHubEnvPublicKeyGetter gets a public key that should be used for secret encryption.
type GitHubEnvPublicKeyGetter func(ctx context.Context, repoID int, env string) (*github.PublicKey, error)

func (f GitHubEnvPublicKeyGetter) GetGitHubEnvPublicKey(ctx context.Context, repoID int, env string) (*github.PublicKey, error) {
	return f(ctx, repoID, env)
}
