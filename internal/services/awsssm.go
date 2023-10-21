package services

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type SSMParameterGetter interface {
	SSMGetParameter(ctx context.Context, region string, in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error)
}
