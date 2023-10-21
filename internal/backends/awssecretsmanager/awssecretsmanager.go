package awssecretsmanager

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/shogo82148/op-sync/internal/backends"
	"github.com/shogo82148/op-sync/internal/maputils"
	"github.com/shogo82148/op-sync/internal/services"
)

func isNotFoundError(err error) bool {
	var awsErr *types.ResourceNotFoundException
	return errors.As(err, &awsErr)
}

type Backend struct {
	opts *Options
}

type Options struct {
	services.OnePasswordReader
	services.STSCallerIdentityGetter
	services.SecretsManagerSecretCreator
	services.SecretsManagerSecretGetter
	services.SecretsManagerSecretPutter
}

func New(opts *Options) *Backend {
	return &Backend{opts: opts}
}

func (b *Backend) Plan(ctx context.Context, params map[string]any) ([]backends.Plan, error) {
	c := new(maputils.Context)
	account := maputils.Must[string](c, params, "account")
	region := maputils.Must[string](c, params, "region")
	name := maputils.Must[string](c, params, "name")
	if err := c.Err(); err != nil {
		return nil, fmt.Errorf("awsssm: validation failed: %w", err)
	}

	id, err := b.opts.STSGetCallerIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get caller identity: %w", err)
	}
	if aws.ToString(id.Account) != account {
		return []backends.Plan{}, nil
	}

	// check the secret exists
	value, err := b.opts.SecretsManagerGetSecretValue(ctx, region, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(name),
	})
	if isNotFoundError(err) {
		return []backends.Plan{
			&PlanCreate{
				backend: b,
				account: account,
				region:  region,
				name:    name,
				secret:  `{"foo":"bar"}`,
			},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get secret value: %w", err)
	}

	// TODO: check the secret value is up-to-date
	_ = value

	return []backends.Plan{
		&PlanUpdate{
			backend: b,
			region:  region,
			arn:     aws.ToString(value.ARN),
			secret:  `{"foo":"bar"}`,
		},
	}, nil
}

var _ backends.Plan = (*PlanCreate)(nil)

type PlanCreate struct {
	backend *Backend
	account string
	region  string
	name    string
	secret  string
}

func (p *PlanCreate) Preview() string {
	return fmt.Sprintf("create AWS Secrets Manager secret %s on account %s", p.name, p.account)
}

func (p *PlanCreate) Apply(ctx context.Context) error {
	_, err := p.backend.opts.SecretsManagerCreateSecret(ctx, p.region, &secretsmanager.CreateSecretInput{
		Name:         aws.String(p.name),
		SecretString: aws.String(p.secret),
	})
	return err
}

var _ backends.Plan = (*PlanUpdate)(nil)

type PlanUpdate struct {
	backend *Backend
	region  string
	arn     string
	secret  string
}

func (p *PlanUpdate) Preview() string {
	return fmt.Sprintf("update AWS Secrets Manager secret %s", p.arn)
}

func (p *PlanUpdate) Apply(ctx context.Context) error {
	_, err := p.backend.opts.SecretsManagerPutSecretValue(ctx, p.region, &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(p.arn),
		SecretString: aws.String(p.secret),
	})
	return err
}
