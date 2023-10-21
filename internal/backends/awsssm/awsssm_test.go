package awsssm

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/shogo82148/op-sync/internal/services/mock"
)

func TestPlan(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var result *ssm.PutParameterInput
	b := New(&Options{
		OnePasswordReader: mock.OnePasswordReader(func(ctx context.Context, uri string) ([]byte, error) {
			return []byte("secret"), nil
		}),
		STSCallerIdentityGetter: mock.STSCallerIdentityGetter(func(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
			return &sts.GetCallerIdentityOutput{
				Account: aws.String("123456789012"),
			}, nil
		}),
		SSMParameterGetter: mock.SSMParameterGetter(func(ctx context.Context, region string, in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
			return nil, &types.ParameterNotFound{}
		}),
		SSMParameterPutter: mock.SSMParameterPutter(func(ctx context.Context, region string, in *ssm.PutParameterInput) (*ssm.PutParameterOutput, error) {
			result = in
			return &ssm.PutParameterOutput{}, nil
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

	// verify the result
	want := &ssm.PutParameterInput{
		Name:        aws.String("/path/to/secret"),
		Description: aws.String("managed by op-sync: op://vault/item/field"),
		Type:        types.ParameterTypeSecureString,
		Value:       aws.String("secret"),
		Overwrite:   aws.Bool(false),
	}
	opts := cmpopts.IgnoreUnexported(ssm.PutParameterInput{})
	if diff := cmp.Diff(want, result, opts); diff != "" {
		t.Errorf("unexpected result (-want +got):\n%s", diff)
	}
}

func TestPlan_Overwrite(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var result *ssm.PutParameterInput
	b := New(&Options{
		OnePasswordReader: mock.OnePasswordReader(func(ctx context.Context, uri string) ([]byte, error) {
			return []byte("secret"), nil
		}),
		STSCallerIdentityGetter: mock.STSCallerIdentityGetter(func(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
			return &sts.GetCallerIdentityOutput{
				Account: aws.String("123456789012"),
			}, nil
		}),
		SSMParameterGetter: mock.SSMParameterGetter(func(ctx context.Context, region string, in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
			return &ssm.GetParameterOutput{
				Parameter: &types.Parameter{
					Value: aws.String("old-secret"),
				},
			}, nil
		}),
		SSMParameterPutter: mock.SSMParameterPutter(func(ctx context.Context, region string, in *ssm.PutParameterInput) (*ssm.PutParameterOutput, error) {
			result = in
			return &ssm.PutParameterOutput{}, nil
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

	// verify the result
	want := &ssm.PutParameterInput{
		Name:        aws.String("/path/to/secret"),
		Description: aws.String("managed by op-sync: op://vault/item/field"),
		Type:        types.ParameterTypeSecureString,
		Value:       aws.String("secret"),
		Overwrite:   aws.Bool(true),
	}
	opts := cmpopts.IgnoreUnexported(ssm.PutParameterInput{})
	if diff := cmp.Diff(want, result, opts); diff != "" {
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
		SSMParameterGetter: mock.SSMParameterGetter(func(ctx context.Context, region string, in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
			return &ssm.GetParameterOutput{
				Parameter: &types.Parameter{
					Value: aws.String("secret"),
				},
			}, nil
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
	if len(plans) != 0 {
		t.Fatalf("unexpected length: want 0, got %d", len(plans))
	}
}
