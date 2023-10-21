package awssecretsmanager

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/shogo82148/op-sync/internal/services/mock"
)

func TestPlan(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var got *secretsmanager.CreateSecretInput
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
			got = in
			return &secretsmanager.CreateSecretOutput{}, nil
		}),
	})

	// do planning
	plans, err := b.Plan(ctx, map[string]any{
		"account": "123456789012",
		"region":  "ap-northeast-1",
		"name":    "secret",
		"template": map[string]any{
			"password": "{{ op://vault/item/field }}",
		},
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

	// verify the result
	want := &secretsmanager.CreateSecretInput{
		Name: aws.String("secret"),
		Description: aws.String(`managed by op-sync:
{
  "password": "{{ op://vault/item/field }}"
}`),
		SecretString: aws.String(`{"password":"secret"}`),
	}
	opts := cmpopts.IgnoreUnexported(secretsmanager.CreateSecretInput{})
	if diff := cmp.Diff(want, got, opts); diff != "" {
		t.Errorf("unexpected result (-want +got):\n%s", diff)
	}
}

func TestPlan_Update(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var got *secretsmanager.PutSecretValueInput
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
			return &secretsmanager.GetSecretValueOutput{
				ARN:          aws.String("arn:aws:secretsmanager:ap-northeast-1:123456789012:secret:secret-abcd"),
				SecretString: aws.String(`{"password":"old-secret"}`),
			}, nil
		}),
		SecretsManagerSecretPutter: mock.SecretsManagerSecretPutter(func(ctx context.Context, region string, in *secretsmanager.PutSecretValueInput) (*secretsmanager.PutSecretValueOutput, error) {
			got = in
			return &secretsmanager.PutSecretValueOutput{}, nil
		}),
	})

	// do planning
	plans, err := b.Plan(ctx, map[string]any{
		"account": "123456789012",
		"region":  "ap-northeast-1",
		"name":    "secret",
		"template": map[string]any{
			"password": "{{ op://vault/item/field }}",
		},
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

	// verify the result
	want := &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String("arn:aws:secretsmanager:ap-northeast-1:123456789012:secret:secret-abcd"),
		SecretString: aws.String(`{"password":"secret"}`),
	}
	opts := cmpopts.IgnoreUnexported(secretsmanager.PutSecretValueInput{})
	if diff := cmp.Diff(want, got, opts); diff != "" {
		t.Errorf("unexpected result (-want +got):\n%s", diff)
	}
}

func TestPlan_UpToDate(t *testing.T) {
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
			return &secretsmanager.GetSecretValueOutput{
				ARN:          aws.String("arn:aws:secretsmanager:ap-northeast-1:123456789012:secret:secret-abcd"),
				SecretString: aws.String(`{"password":"secret"}`),
			}, nil
		}),
	})

	// do planning
	plans, err := b.Plan(ctx, map[string]any{
		"account": "123456789012",
		"region":  "ap-northeast-1",
		"name":    "secret",
		"template": map[string]any{
			"password": "{{ op://vault/item/field }}",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// verify the plan
	if len(plans) != 0 {
		t.Fatalf("unexpected length: want 0, got %d", len(plans))
	}
}
