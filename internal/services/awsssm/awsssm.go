package awsssm

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/shogo82148/op-sync/internal/services"
)

type Service struct {
	mu  sync.Mutex
	svc map[string]*ssm.Client
}

func New() *Service {
	return &Service{}
}

func (s *Service) getClient(ctx context.Context, region string) (*ssm.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.svc == nil {
		s.svc = make(map[string]*ssm.Client)
	}
	if svc, ok := s.svc[region]; !ok {
		return svc, nil
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}
	svc := ssm.NewFromConfig(cfg)
	s.svc[region] = svc
	return svc, nil
}

var _ services.SSMParameterGetter = (*Service)(nil)

func (s *Service) SSMGetParameter(ctx context.Context, region string, in *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	svc, err := s.getClient(ctx, region)
	if err != nil {
		return nil, err
	}

	slog.DebugContext(ctx, "get ssm parameter", slog.String("name", aws.ToString(in.Name)))
	return svc.GetParameter(ctx, in)
}
