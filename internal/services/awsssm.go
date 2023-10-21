package services

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sts"
)

type STSCallerIdentityGetter interface {
	STSGetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error)
}
