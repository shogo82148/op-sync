package awssts

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/shogo82148/op-sync/internal/services"
)

var _ services.STSCallerIdentityGetter = (*Service)(nil)

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) STSGetCallerIdentity(ctx context.Context) (*sts.GetCallerIdentityOutput, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	svc := sts.NewFromConfig(cfg)
	slog.DebugContext(ctx, "get caller identity")
	return svc.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
}
