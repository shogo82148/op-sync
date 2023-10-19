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
