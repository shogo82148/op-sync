package github

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	"github.com/google/go-github/v56/github"
	"github.com/shogo82148/op-sync/internal/services"
	"github.com/shogo82148/op-sync/internal/services/mock"
	"golang.org/x/crypto/nacl/box"
)

func TestPlan_RepoSecret(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Key is generated by gen_key.go
	pubKey := "X4Dk1lVSh8C21cDezdGivFERcid2L7UAfFCWeiRn1mU="
	privKey := "BnLGcl9+miWXxrHDmWOROLRqBU6mM/biZTYO16LcRag="

	var repoSecret *github.EncryptedSecret
	b := New(&Options{
		OnePasswordItemGetter: mock.OnePasswordItemGetter(func(ctx context.Context, repository, name string) (*services.OnePasswordItem, error) {
			return &services.OnePasswordItem{}, nil
		}),
		OnePasswordReader: mock.OnePasswordReader(func(ctx context.Context, uri string) ([]byte, error) {
			return []byte("secret"), nil
		}),
		GitHubRepoSecretGetter: mock.GitHubRepoSecretGetter(func(ctx context.Context, app services.GitHubApplication, owner, repo, name string) (*github.Secret, error) {
			return nil, &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: http.StatusNotFound,
				},
			}
		}),
		GitHubRepoSecretCreator: mock.GitHubRepoSecretCreator(func(ctx context.Context, app services.GitHubApplication, owner, repo string, secret *github.EncryptedSecret) error {
			if app != services.GitHubApplicationActions {
				t.Errorf("unexpected application: want actions, got %s", app)
			}
			if owner != "shogo82148" || repo != "op-sync" {
				t.Errorf("unexpected repository: want shogo82148/op-sync, got %s/%s", owner, repo)
			}
			if secret.Name != "VERY_SECRET_TOKEN" {
				t.Errorf("unexpected name: want VERY_SECRET_TOKEN, got %s", secret.Name)
			}
			repoSecret = secret
			return nil
		}),
		GitHubRepoPublicKeyGetter: mock.GitHubRepoPublicKeyGetter(func(ctx context.Context, owner, repo string) (*github.PublicKey, error) {
			return &github.PublicKey{
				KeyID: github.String("key_id"),
				Key:   github.String(pubKey),
			}, nil
		}),
	})

	// do planning
	plans, err := b.Plan(ctx, map[string]any{
		"repository": "shogo82148/op-sync",
		"name":       "VERY_SECRET_TOKEN",
		"source":     "op://vault/item/VERY_SECRET_TOKEN",
	})
	if err != nil {
		t.Fatal(err)
	}

	// verify the plan
	if len(plans) != 1 {
		t.Fatalf("unexpected length: want 1, got %d", len(plans))
	}

	// apply the plan
	if err := plans[0].Apply(ctx); err != nil {
		t.Fatal(err)
	}

	// verify the secret
	if repoSecret == nil {
		t.Fatal("repoSecret is not set")
	}
	encrypted, err := base64.StdEncoding.DecodeString(repoSecret.EncryptedValue)
	if err != nil {
		t.Fatal(err)
	}
	decodedPubKey, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		t.Fatal(err)
	}
	decodedPrivKey, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		t.Fatal(err)
	}

	var peersPubKey [32]byte
	copy(peersPubKey[:], decodedPubKey)
	var peersPrivKey [32]byte
	copy(peersPrivKey[:], decodedPrivKey)
	message, ok := box.OpenAnonymous(nil, encrypted, &peersPubKey, &peersPrivKey)
	if !ok {
		t.Fatal("failed to decrypt")
	}
	if string(message) != "secret" {
		t.Errorf("unexpected message: want secret, got %s", string(message))
	}
}

func TestPlan_RepoSecret_UpToDate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b := New(&Options{
		OnePasswordItemGetter: mock.OnePasswordItemGetter(func(ctx context.Context, repository, name string) (*services.OnePasswordItem, error) {
			return &services.OnePasswordItem{
				UpdatedAt: time.Unix(1234567890, 0),
			}, nil
		}),
		OnePasswordReader: mock.OnePasswordReader(func(ctx context.Context, uri string) ([]byte, error) {
			return []byte("secret"), nil
		}),
		GitHubRepoSecretGetter: mock.GitHubRepoSecretGetter(func(ctx context.Context, app services.GitHubApplication, owner, repo, name string) (*github.Secret, error) {
			return &github.Secret{
				UpdatedAt: github.Timestamp{Time: time.Unix(1234567890+1, 0)},
			}, nil
		}),
	})

	// do planning
	plans, err := b.Plan(ctx, map[string]any{
		"repository": "shogo82148/op-sync",
		"name":       "VERY_SECRET_TOKEN",
		"source":     "op://vault/item/VERY_SECRET_TOKEN",
	})
	if err != nil {
		t.Fatal(err)
	}

	// verify the plan
	// the secret is up-to-date, so the plan should be empty.
	if len(plans) != 0 {
		t.Fatalf("unexpected length: want 0, got %d", len(plans))
	}
}

func TestPlan_RepoSecret_for_Dependabot(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Key is generated by gen_key.go
	pubKey := "X4Dk1lVSh8C21cDezdGivFERcid2L7UAfFCWeiRn1mU="
	privKey := "BnLGcl9+miWXxrHDmWOROLRqBU6mM/biZTYO16LcRag="

	var repoSecret *github.EncryptedSecret
	b := New(&Options{
		OnePasswordItemGetter: mock.OnePasswordItemGetter(func(ctx context.Context, repository, name string) (*services.OnePasswordItem, error) {
			return &services.OnePasswordItem{}, nil
		}),
		OnePasswordReader: mock.OnePasswordReader(func(ctx context.Context, uri string) ([]byte, error) {
			return []byte("secret"), nil
		}),
		GitHubRepoSecretGetter: mock.GitHubRepoSecretGetter(func(ctx context.Context, app services.GitHubApplication, owner, repo, name string) (*github.Secret, error) {
			return nil, &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: http.StatusNotFound,
				},
			}
		}),
		GitHubRepoSecretCreator: mock.GitHubRepoSecretCreator(func(ctx context.Context, app services.GitHubApplication, owner, repo string, secret *github.EncryptedSecret) error {
			if app != services.GitHubApplicationDependabot {
				t.Errorf("unexpected application: want actions, got %s", app)
			}
			if owner != "shogo82148" || repo != "op-sync" {
				t.Errorf("unexpected repository: want shogo82148/op-sync, got %s/%s", owner, repo)
			}
			if secret.Name != "VERY_SECRET_TOKEN" {
				t.Errorf("unexpected name: want VERY_SECRET_TOKEN, got %s", secret.Name)
			}
			repoSecret = secret
			return nil
		}),
		GitHubRepoPublicKeyGetter: mock.GitHubRepoPublicKeyGetter(func(ctx context.Context, owner, repo string) (*github.PublicKey, error) {
			return &github.PublicKey{
				KeyID: github.String("key_id"),
				Key:   github.String(pubKey),
			}, nil
		}),
	})

	// do planning
	plans, err := b.Plan(ctx, map[string]any{
		"repository":  "shogo82148/op-sync",
		"name":        "VERY_SECRET_TOKEN",
		"application": "dependabot",
		"source":      "op://vault/item/VERY_SECRET_TOKEN",
	})
	if err != nil {
		t.Fatal(err)
	}

	// verify the plan
	if len(plans) != 1 {
		t.Fatalf("unexpected length: want 1, got %d", len(plans))
	}

	// apply the plan
	if err := plans[0].Apply(ctx); err != nil {
		t.Fatal(err)
	}

	// verify the secret
	if repoSecret == nil {
		t.Fatal("repoSecret is not set")
	}
	encrypted, err := base64.StdEncoding.DecodeString(repoSecret.EncryptedValue)
	if err != nil {
		t.Fatal(err)
	}
	decodedPubKey, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		t.Fatal(err)
	}
	decodedPrivKey, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		t.Fatal(err)
	}

	var peersPubKey [32]byte
	copy(peersPubKey[:], decodedPubKey)
	var peersPrivKey [32]byte
	copy(peersPrivKey[:], decodedPrivKey)
	message, ok := box.OpenAnonymous(nil, encrypted, &peersPubKey, &peersPrivKey)
	if !ok {
		t.Fatal("failed to decrypt")
	}
	if string(message) != "secret" {
		t.Errorf("unexpected message: want secret, got %s", string(message))
	}
}

func TestPlan_EnvSecret(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Key is generated by gen_key.go
	pubKey := "X4Dk1lVSh8C21cDezdGivFERcid2L7UAfFCWeiRn1mU="
	privKey := "BnLGcl9+miWXxrHDmWOROLRqBU6mM/biZTYO16LcRag="

	var envSecret *github.EncryptedSecret
	b := New(&Options{
		OnePasswordItemGetter: mock.OnePasswordItemGetter(func(ctx context.Context, repository, name string) (*services.OnePasswordItem, error) {
			return &services.OnePasswordItem{}, nil
		}),
		OnePasswordReader: mock.OnePasswordReader(func(ctx context.Context, uri string) ([]byte, error) {
			return []byte("secret"), nil
		}),
		GitHubRepoGetter: mock.GitHubRepoGetter(func(ctx context.Context, owner, repo string) (*github.Repository, error) {
			return &github.Repository{
				ID: github.Int64(1234567890),
			}, nil
		}),
		GitHubEnvSecretGetter: mock.GitHubEnvSecretGetter(func(ctx context.Context, repoID int, env, name string) (*github.Secret, error) {
			return nil, &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: http.StatusNotFound,
				},
			}
		}),
		GitHubEnvSecretCreator: mock.GitHubEnvSecretCreator(func(ctx context.Context, repoID int, env string, secret *github.EncryptedSecret) error {
			if repoID != 1234567890 {
				t.Errorf("unexpected repository ID: want 1234567890, got %d", repoID)
			}
			if env != "test" {
				t.Errorf("unexpected environment: want test, got %s", env)
			}
			if secret.Name != "VERY_SECRET_TOKEN" {
				t.Errorf("unexpected name: want VERY_SECRET_TOKEN, got %s", secret.Name)
			}
			envSecret = secret
			return nil
		}),
		GitHubEnvPublicKeyGetter: mock.GitHubEnvPublicKeyGetter(func(ctx context.Context, repoID int, env string) (*github.PublicKey, error) {
			return &github.PublicKey{
				KeyID: github.String("key_id"),
				Key:   github.String(pubKey),
			}, nil
		}),
	})

	// do planning
	plans, err := b.Plan(ctx, map[string]any{
		"repository":  "shogo82148/op-sync",
		"environment": "test",
		"name":        "VERY_SECRET_TOKEN",
		"source":      "op://vault/item/VERY_SECRET_TOKEN",
	})
	if err != nil {
		t.Fatal(err)
	}

	// verify the plan
	if len(plans) != 1 {
		t.Fatalf("unexpected length: want 1, got %d", len(plans))
	}

	// apply the plan
	if err := plans[0].Apply(ctx); err != nil {
		t.Fatal(err)
	}

	// verify the secret
	if envSecret == nil {
		t.Fatal("envSecret is not set")
	}
	encrypted, err := base64.StdEncoding.DecodeString(envSecret.EncryptedValue)
	if err != nil {
		t.Fatal(err)
	}
	decodedPubKey, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		t.Fatal(err)
	}
	decodedPrivKey, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		t.Fatal(err)
	}

	var peersPubKey [32]byte
	copy(peersPubKey[:], decodedPubKey)
	var peersPrivKey [32]byte
	copy(peersPrivKey[:], decodedPrivKey)
	message, ok := box.OpenAnonymous(nil, encrypted, &peersPubKey, &peersPrivKey)
	if !ok {
		t.Fatal("failed to decrypt")
	}
	if string(message) != "secret" {
		t.Errorf("unexpected message: want secret, got %s", string(message))
	}
}

func TestPlan_EnvSecret_UpToDate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b := New(&Options{
		OnePasswordItemGetter: mock.OnePasswordItemGetter(func(ctx context.Context, repository, name string) (*services.OnePasswordItem, error) {
			return &services.OnePasswordItem{
				UpdatedAt: time.Unix(1234567890, 0),
			}, nil
		}),
		OnePasswordReader: mock.OnePasswordReader(func(ctx context.Context, uri string) ([]byte, error) {
			return []byte("secret"), nil
		}),
		GitHubRepoGetter: mock.GitHubRepoGetter(func(ctx context.Context, owner, repo string) (*github.Repository, error) {
			return &github.Repository{
				ID: github.Int64(1234567890),
			}, nil
		}),
		GitHubEnvSecretGetter: mock.GitHubEnvSecretGetter(func(ctx context.Context, repoID int, env, name string) (*github.Secret, error) {
			return &github.Secret{
				UpdatedAt: github.Timestamp{Time: time.Unix(1234567890+1, 0)},
			}, nil
		}),
	})

	// do planning
	plans, err := b.Plan(ctx, map[string]any{
		"repository":  "shogo82148/op-sync",
		"environment": "test",
		"name":        "VERY_SECRET_TOKEN",
		"source":      "op://vault/item/VERY_SECRET_TOKEN",
	})
	if err != nil {
		t.Fatal(err)
	}

	// verify the plan
	if len(plans) != 0 {
		t.Fatalf("unexpected length: want 0, got %d", len(plans))
	}
}

func TestPlan_OrgSecret(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Key is generated by gen_key.go
	pubKey := "X4Dk1lVSh8C21cDezdGivFERcid2L7UAfFCWeiRn1mU="
	privKey := "BnLGcl9+miWXxrHDmWOROLRqBU6mM/biZTYO16LcRag="

	var orgSecret *github.EncryptedSecret
	b := New(&Options{
		OnePasswordItemGetter: mock.OnePasswordItemGetter(func(ctx context.Context, repository, name string) (*services.OnePasswordItem, error) {
			return &services.OnePasswordItem{}, nil
		}),
		OnePasswordReader: mock.OnePasswordReader(func(ctx context.Context, uri string) ([]byte, error) {
			return []byte("secret"), nil
		}),
		GitHubOrgSecretGetter: mock.GitHubOrgSecretGetter(func(ctx context.Context, app services.GitHubApplication, org, name string) (*github.Secret, error) {
			return nil, &github.ErrorResponse{
				Response: &http.Response{
					StatusCode: http.StatusNotFound,
				},
			}
		}),
		GitHubOrgSecretCreator: mock.GitHubOrgSecretCreator(func(ctx context.Context, app services.GitHubApplication, org string, secret *github.EncryptedSecret) error {
			if app != services.GitHubApplicationActions {
				t.Errorf("unexpected application: want actions, got %s", app)
			}
			if secret.Name != "VERY_SECRET_TOKEN" {
				t.Errorf("unexpected name: want VERY_SECRET_TOKEN, got %s", secret.Name)
			}
			orgSecret = secret
			return nil
		}),
		GitHubOrgPublicKeyGetter: mock.GitHubOrgPublicKeyGetter(func(ctx context.Context, org string) (*github.PublicKey, error) {
			return &github.PublicKey{
				KeyID: github.String("key_id"),
				Key:   github.String(pubKey),
			}, nil
		}),
	})

	// do planning
	plans, err := b.Plan(ctx, map[string]any{
		"organization": "shogo82148",
		"name":         "VERY_SECRET_TOKEN",
		"source":       "op://vault/item/VERY_SECRET_TOKEN",
	})
	if err != nil {
		t.Fatal(err)
	}

	// verify the plan
	if len(plans) != 1 {
		t.Fatalf("unexpected length: want 1, got %d", len(plans))
	}

	// apply the plan
	if err := plans[0].Apply(ctx); err != nil {
		t.Fatal(err)
	}

	// verify the secret
	if orgSecret == nil {
		t.Fatal("envSecret is not set")
	}
	encrypted, err := base64.StdEncoding.DecodeString(orgSecret.EncryptedValue)
	if err != nil {
		t.Fatal(err)
	}
	decodedPubKey, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		t.Fatal(err)
	}
	decodedPrivKey, err := base64.StdEncoding.DecodeString(privKey)
	if err != nil {
		t.Fatal(err)
	}

	var peersPubKey [32]byte
	copy(peersPubKey[:], decodedPubKey)
	var peersPrivKey [32]byte
	copy(peersPrivKey[:], decodedPrivKey)
	message, ok := box.OpenAnonymous(nil, encrypted, &peersPubKey, &peersPrivKey)
	if !ok {
		t.Fatal("failed to decrypt")
	}
	if string(message) != "secret" {
		t.Errorf("unexpected message: want secret, got %s", string(message))
	}
}
