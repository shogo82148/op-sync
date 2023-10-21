package mock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/shogo82148/op-sync/internal/services"
)

var _ services.STSCallerIdentityGetter = STSCallerIdentityGetter(nil)

type STSCallerIdentityGetter func(ctx context.Context) (*sts.GetCallerIdentityOutput, error)

func (f STSCallerIdentityGetter) STSGetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
	return f(ctx)
}
