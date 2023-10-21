package mock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/shogo82148/op-sync/internal/services"
)

var _ services.SecretsManagerSecretCreator = SecretsManagerSecretCreator(nil)

type SecretsManagerSecretCreator func(ctx context.Context, region string, in *secretsmanager.CreateSecretInput) (*secretsmanager.CreateSecretOutput, error)

func (f SecretsManagerSecretCreator) SecretsManagerCreateSecret(ctx context.Context, region string, in *secretsmanager.CreateSecretInput) (*secretsmanager.CreateSecretOutput, error) {
	return f(ctx, region, in)
}

var _ services.SecretsManagerSecretGetter = SecretsManagerSecretGetter(nil)

type SecretsManagerSecretGetter func(ctx context.Context, region string, in *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error)

func (f SecretsManagerSecretGetter) SecretsManagerGetSecretValue(ctx context.Context, region string, in *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	return f(ctx, region, in)
}

var _ services.SecretsManagerSecretUpdater = SecretsManagerSecretUpdater(nil)

type SecretsManagerSecretUpdater func(ctx context.Context, region string, in *secretsmanager.UpdateSecretInput) (*secretsmanager.UpdateSecretOutput, error)

func (f SecretsManagerSecretUpdater) SecretsManagerUpdateSecret(ctx context.Context, region string, in *secretsmanager.UpdateSecretInput) (*secretsmanager.UpdateSecretOutput, error) {
	return f(ctx, region, in)
}
