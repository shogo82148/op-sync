package op

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/shogo82148/op-sync/internal/services"
)

func command(ctx context.Context, name string, args ...string) *exec.Cmd {
	slog.DebugContext(ctx, "run 1password cli", slog.String("name", name), slog.Any("args", args))
	return exec.CommandContext(ctx, name, args...)
}

func wrap(err error) error {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return fmt.Errorf("failed to run op command: %q: %w", string(exitErr.Stderr), err)
	}
	return fmt.Errorf("failed to run op command: %w", err)
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
		return nil, wrap(err)
	}
	var info services.OnePasswordUser
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to parse the output of op whoami: %w", err)
	}
	return &info, nil
}

var _ services.Injector = (*Service)(nil)

// Injector inject the secrets into the template.
func (s *Service) Inject(ctx context.Context, tmpl string) ([]byte, error) {
	cmd := command(ctx, "op", "inject")
	cmd.Stdin = strings.NewReader(tmpl)
	data, err := cmd.Output()
	if err != nil {
		return nil, wrap(err)
	}
	return data, nil
}

// GetOnePasswordItem gets the item from 1password.
func (s *Service) GetOnePasswordItem(ctx context.Context, vault, item string) (*services.OnePasswordItem, error) {
	cmd := command(ctx, "op", "item", "get", item, "--vault", vault, "--format=json")
	data, err := cmd.Output()
	if err != nil {
		return nil, wrap(err)
	}
	var info services.OnePasswordItem
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to parse the output of op get item: %w", err)
	}
	return &info, nil
}

// ReadOnePassword reads the secret from 1password.
func (s *Service) ReadOnePassword(ctx context.Context, uri string) ([]byte, error) {
	cmd := command(ctx, "op", "read", "--no-newline", uri)
	data, err := cmd.Output()
	if err != nil {
		return nil, wrap(err)
	}
	return data, nil
}
