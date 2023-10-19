package services

import (
	"context"

	"github.com/google/go-github/v56/github"
)

// GitHubUserGetter fetches the authenticated GitHub user.
type GitHubUserGetter interface {
	GetGitHubUser(ctx context.Context) (*github.User, error)
}

// GitHubRepoGetter fetches the GitHub repository.
type GitHubRepoGetter interface {
	GetGitHubRepo(ctx context.Context, owner, repo string) (*github.Repository, error)
}

// GitHubRepoSecretGetter gets a single repository secret without revealing its encrypted value.
type GitHubRepoSecretGetter interface {
	GetGitHubRepoSecret(ctx context.Context, owner, repo, name string) (*github.Secret, error)
}

// GitHubRepoSecretCreator creates or updates a repository secret with an encrypted value.
type GitHubRepoSecretCreator interface {
	CreateGitHubRepoSecret(ctx context.Context, owner, repo string, secret *github.EncryptedSecret) error
}

// GitHubRepoPublicKeyGetter gets a public key that should be used for secret encryption.
type GitHubRepoPublicKeyGetter interface {
	GetGitHubRepoPublicKey(ctx context.Context, owner, repo string) (*github.PublicKey, error)
}

// GitHubEnvSecretGetter gets a single environment secret without revealing its encrypted value.
type GitHubEnvSecretGetter interface {
	GetGitHubEnvSecret(ctx context.Context, repoID int, env, name string) (*github.Secret, error)
}

// GitHubEnvSecretCreator creates or updates a single environment secret with an encrypted value.
type GitHubEnvSecretCreator interface {
	CreateGitHubEnvSecret(ctx context.Context, repoID int, env string, secret *github.EncryptedSecret) error
}

// GitHubEnvPublicKeyGetter gets a public key that should be used for secret encryption.
type GitHubEnvPublicKeyGetter interface {
	GetGitHubEnvPublicKey(ctx context.Context, repoID int, env string) (*github.PublicKey, error)
}
