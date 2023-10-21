package services

import (
	"context"

	"github.com/google/go-github/v56/github"
)

// GitHubApplication represents a GitHub application for secret management.
type GitHubApplication string

const (
	GitHubApplicationActions    GitHubApplication = "actions"
	GitHubApplicationCodespaces GitHubApplication = "codespaces"
	GitHubApplicationDependabot GitHubApplication = "dependabot"
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
	GetGitHubRepoSecret(ctx context.Context, app GitHubApplication, owner, repo, name string) (*github.Secret, error)
}

// GitHubRepoSecretCreator creates or updates a repository secret with an encrypted value.
type GitHubRepoSecretCreator interface {
	CreateGitHubRepoSecret(ctx context.Context, app GitHubApplication, owner, repo string, secret *github.EncryptedSecret) error
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

// GitHubOrgSecretGetter gets a single organization secret without revealing its encrypted value.
type GitHubOrgSecretGetter interface {
	GetGitHubOrgSecret(ctx context.Context, app GitHubApplication, org, name string) (*github.Secret, error)
}

// GitHubOrgSecretCreator creates or updates a single organization secret with an encrypted value.
type GitHubOrgSecretCreator interface {
	CreateGitHubOrgSecret(ctx context.Context, app GitHubApplication, org string, secret *github.EncryptedSecret) error
}

// GitHubOrgPublicKeyGetter gets a public key that should be used for secret encryption.
type GitHubOrgPublicKeyGetter interface {
	GetGitHubOrgPublicKey(ctx context.Context, org string) (*github.PublicKey, error)
}

// GitHubReposIDForOrgSecretLister lists all repositories that have access to a secret.
type GitHubReposIDForOrgSecretLister interface {
	ListGitHubReposIDForOrgSecret(ctx context.Context, app GitHubApplication, org, name string) ([]int64, error)
}
