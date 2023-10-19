package services

import (
	"context"

	"github.com/google/go-github/v56/github"
)

// GitHubUserGetter fetches the authenticated GitHub user.
type GitHubUserGetter interface {
	GetGitHubUser(ctx context.Context) (*github.User, error)
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
