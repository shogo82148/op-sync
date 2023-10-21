package mock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/shogo82148/op-sync/internal/services"
)

var _ services.SSMParameterGetter = SSMParameterGetter(nil)

type SSMParameterGetter func(ctx context.Context, region string, in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error)

func (f SSMParameterGetter) SSMGetParameter(ctx context.Context, region string, in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	return f(ctx, region, in)
}

var _ services.SSMParameterPutter = SSMParameterPutter(nil)

type SSMParameterPutter func(ctx context.Context, region string, in *ssm.PutParameterInput) (*ssm.PutParameterOutput, error)

func (f SSMParameterPutter) SSMPutParameter(ctx context.Context, region string, in *ssm.PutParameterInput) (*ssm.PutParameterOutput, error) {
	return f(ctx, region, in)
}
