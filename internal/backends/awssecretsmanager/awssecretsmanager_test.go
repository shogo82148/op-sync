package awssecretsmanager

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/shogo82148/op-sync/internal/services/mock"
)

func TestPlan(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b := New(&Options{
		OnePasswordReader: mock.OnePasswordReader(func(ctx context.Context, uri string) ([]byte, error) {
			return []byte("secret"), nil
		}),
		STSCallerIdentityGetter: mock.STSCallerIdentityGetter(func(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
			return &sts.GetCallerIdentityOutput{
				Account: aws.String("123456789012"),
			}, nil
		}),
		SecretsManagerSecretGetter: mock.SecretsManagerSecretGetter(func(ctx context.Context, region string, in *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
			return nil, &types.ResourceNotFoundException{}
		}),
		SecretsManagerSecretCreator: mock.SecretsManagerSecretCreator(func(ctx context.Context, region string, in *secretsmanager.CreateSecretInput) (*secretsmanager.CreateSecretOutput, error) {
			return &secretsmanager.CreateSecretOutput{}, nil
		}),
	})

	// do planning
	plans, err := b.Plan(ctx, map[string]any{
		"account": "123456789012",
		"region":  "ap-northeast-1",
		"name":    "/path/to/secret",
		"source":  "op://vault/item/field",
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
}
