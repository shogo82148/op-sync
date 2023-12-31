package awsssm

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/shogo82148/op-sync/internal/backends"
	"github.com/shogo82148/op-sync/internal/maputils"
	"github.com/shogo82148/op-sync/internal/services"
)

func isNotFoundError(err error) bool {
	var awsErr *types.ParameterNotFound
	return errors.As(err, &awsErr)
}

var _ backends.Backend = (*Backend)(nil)

type Backend struct {
	opts *Options
}

type Options struct {
	services.OnePasswordReader
	services.STSCallerIdentityGetter
	services.SSMParameterGetter
	services.SSMParameterPutter
}

func New(opts *Options) *Backend {
	return &Backend{opts: opts}
}

func (b *Backend) Plan(ctx context.Context, params map[string]any) ([]backends.Plan, error) {
	c := new(maputils.Context)
	account := maputils.Must[string](c, params, "account")
	region := maputils.Must[string](c, params, "region")
	name := maputils.Must[string](c, params, "name")
	source := maputils.Must[string](c, params, "source")
	description, hasDescription := maputils.Get[string](c, params, "description")
	if err := c.Err(); err != nil {
		return nil, fmt.Errorf("awsssm: validation failed: %w", err)
	}
	if !hasDescription {
		description = fmt.Sprintf("managed by op-sync: %s", source)
	}

	id, err := b.opts.STSGetCallerIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get caller identity: %w", err)
	}
	if aws.ToString(id.Account) != account {
		return []backends.Plan{}, nil
	}

	secret, err := b.opts.ReadOnePassword(ctx, source)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret: %w", err)
	}

	param, err := b.opts.SSMGetParameter(ctx, region, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(true),
	})
	if isNotFoundError(err) {
		return []backends.Plan{
			&Plan{
				backend:     b,
				account:     account,
				region:      region,
				name:        name,
				description: description,
				secret:      secret,
				overwrite:   false,
			},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get parameter from parameter store: %w", err)
	}

	if aws.ToString(param.Parameter.Value) == string(secret) {
		return []backends.Plan{}, nil
	}

	return []backends.Plan{
		&Plan{
			backend:     b,
			account:     account,
			region:      region,
			name:        name,
			description: description,
			secret:      secret,
			overwrite:   true,
		},
	}, nil
}

var _ backends.Plan = (*Plan)(nil)

type Plan struct {
	backend     *Backend
	account     string
	region      string
	name        string
	description string
	secret      []byte
	overwrite   bool
}

func (p *Plan) Preview() string {
	if p.overwrite {
		return fmt.Sprintf("aws ssm parameter store %s on account %s will be updated", p.name, p.account)
	}
	return fmt.Sprintf("aws ssm parameter store %s on account %s will be created", p.name, p.account)
}

func (p *Plan) Apply(ctx context.Context) error {
	_, err := p.backend.opts.SSMPutParameter(ctx, p.region, &ssm.PutParameterInput{
		Name:        aws.String(p.name),
		Type:        types.ParameterTypeSecureString,
		Value:       aws.String(string(p.secret)),
		Description: aws.String(p.description),
		Overwrite:   aws.Bool(p.overwrite),
	})
	if err != nil {
		return fmt.Errorf("failed to put parameter to parameter store: %w", err)
	}
	return nil
}
