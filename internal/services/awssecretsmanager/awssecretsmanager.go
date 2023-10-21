package awssecretsmanager

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/shogo82148/op-sync/internal/services"
)

type Service struct {
	mu  sync.Mutex
	svc map[string]*secretsmanager.Client
}

func New() *Service {
	return &Service{}
}

func (s *Service) getClient(ctx context.Context, region string) (*secretsmanager.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.svc == nil {
		s.svc = make(map[string]*secretsmanager.Client)
	}
	if svc, ok := s.svc[region]; ok {
		return svc, nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}
	svc := secretsmanager.NewFromConfig(cfg)
	s.svc[region] = svc
	return svc, nil
}

var _ services.SecretsManagerSecretCreator = (*Service)(nil)

// SecretsManagerCreateSecret creates a new secret.
func (s *Service) SecretsManagerCreateSecret(ctx context.Context, region string, in *secretsmanager.CreateSecretInput) (*secretsmanager.CreateSecretOutput, error) {
	svc, err := s.getClient(ctx, region)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "create secrets manager secret", slog.String("name", aws.ToString(in.Name)))
	return svc.CreateSecret(ctx, in)
}

var _ services.SecretsManagerSecretGetter = (*Service)(nil)

// SecretsManagerGetSecretValue retrieves the contents of the encrypted fields SecretString or SecretBinary from the specified version of a secret.
func (s *Service) SecretsManagerGetSecretValue(ctx context.Context, region string, in *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	svc, err := s.getClient(ctx, region)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "get secrets manager secret", slog.String("name", aws.ToString(in.SecretId)))
	return svc.GetSecretValue(ctx, in)
}

var _ services.SecretsManagerSecretPutter = (*Service)(nil)

// SecretsManagerPutSecretValue creates a new version with a new encrypted secret value and attaches it to the secret.
func (s *Service) SecretsManagerPutSecretValue(ctx context.Context, region string, in *secretsmanager.PutSecretValueInput) (*secretsmanager.PutSecretValueOutput, error) {
	svc, err := s.getClient(ctx, region)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "put secrets manager secret", slog.String("name", aws.ToString(in.SecretId)))
	return svc.PutSecretValue(ctx, in)
}
