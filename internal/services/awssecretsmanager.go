package services

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type SecretsManagerSecretCreator interface {
	SecretsManagerCreateSecret(ctx context.Context, region string, in *secretsmanager.CreateSecretInput) (*secretsmanager.CreateSecretOutput, error)
}

type SecretsManagerSecretGetter interface {
	SecretsManagerGetSecretValue(ctx context.Context, region string, in *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error)
}

type SecretsManagerSecretPutter interface {
	SecretsManagerPutSecretValue(ctx context.Context, region string, in *secretsmanager.PutSecretValueInput) (*secretsmanager.PutSecretValueOutput, error)
}
