package op

import (
	"context"
	"encoding/json"
	"log/slog"
	"os/exec"

	"github.com/shogo82148/op-sync/internal/services"
)

func command(ctx context.Context, name string, args ...string) *exec.Cmd {
	slog.DebugContext(ctx, "run 1password cli", slog.String("name", name), slog.Any("args", args))
	return exec.CommandContext(ctx, name, args...)
}

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

var _ services.WhoAmIer = (*Service)(nil)

// WhoAmI returns the information about a signed-in account.
func (s *Service) WhoAmI(ctx context.Context) (*services.OnePasswordUser, error) {
	cmd := command(ctx, "op", "whoami", "--format=json")
	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var info services.OnePasswordUser
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}
