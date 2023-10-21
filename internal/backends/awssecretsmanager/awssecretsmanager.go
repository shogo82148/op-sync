package awssecretsmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

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
	template := maputils.Must[map[string]any](c, params, "template")
	description, hasDescription := maputils.Get[string](c, params, "description")
	if err := c.Err(); err != nil {
		return nil, fmt.Errorf("awsssm: validation failed: %w", err)
	}
	if !hasDescription {
		data, err := json.MarshalIndent(template, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal the template: %w", err)
		}
		description = fmt.Sprintf("managed by op-sync:\n%s", string(data))
	}

	id, err := b.opts.STSGetCallerIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get caller identity: %w", err)
	}
	if aws.ToString(id.Account) != account {
		return []backends.Plan{}, nil
	}

	// inject the template
	injected, err := b.inject(ctx, template)
	if err != nil {
		return nil, err
	}

	// check the secret exists
	value, err := b.opts.SecretsManagerGetSecretValue(ctx, region, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(name),
	})
	if isNotFoundError(err) {
		// the secret doesn't exist. create it.
		data, err := json.Marshal(injected)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal the secret value: %w", err)
		}
		return []backends.Plan{
			&PlanCreate{
				backend:     b,
				account:     account,
				region:      region,
				name:        name,
				description: description,
				secret:      string(data),
			},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get secret value: %w", err)
	}

	// check the secret value is up-to-date
	var current any
	err = json.Unmarshal([]byte(aws.ToString(value.SecretString)), &current)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret value: %w", err)
	}
	if reflect.DeepEqual(injected, current) {
		return []backends.Plan{}, nil
	}

	// update the secret
	data, err := json.Marshal(injected)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the secret value: %w", err)
	}
	return []backends.Plan{
		&PlanUpdate{
			backend: b,
			region:  region,
			arn:     aws.ToString(value.ARN),
			secret:  string(data),
		},
	}, nil
}

func (b *Backend) inject(ctx context.Context, template any) (any, error) {
	switch tmpl := template.(type) {
	case string:
		if strings.HasPrefix(tmpl, "{{") && strings.HasSuffix(tmpl, "}}") {
			uri := strings.TrimSpace(tmpl[2 : len(tmpl)-2])
			secret, err := b.opts.ReadOnePassword(ctx, uri)
			if err != nil {
				return nil, err
			}
			return string(secret), nil
		}
		return tmpl, nil
	case []any:
		ret := make([]any, 0, len(tmpl))
		for _, v := range tmpl {
			injected, err := b.inject(ctx, v)
			if err != nil {
				return nil, err
			}
			ret = append(ret, injected)
		}
		return ret, nil
	case map[string]any:
		ret := make(map[string]any, len(tmpl))
		for k, v := range tmpl {
			injected, err := b.inject(ctx, v)
			if err != nil {
				return nil, err
			}
			ret[k] = injected
		}
		return ret, nil
	default:
		return tmpl, nil
	}
}

var _ backends.Plan = (*PlanCreate)(nil)

type PlanCreate struct {
	backend     *Backend
	account     string
	region      string
	name        string
	description string
	secret      string
}

func (p *PlanCreate) Preview() string {
	return fmt.Sprintf("create AWS Secrets Manager secret %s on account %s", p.name, p.account)
}

func (p *PlanCreate) Apply(ctx context.Context) error {
	_, err := p.backend.opts.SecretsManagerCreateSecret(ctx, p.region, &secretsmanager.CreateSecretInput{
		Name:         aws.String(p.name),
		Description:  aws.String(p.description),
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
