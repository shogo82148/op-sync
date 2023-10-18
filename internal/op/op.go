package op

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

type UserInfo struct {
	URL         string `json:"url"`
	Email       string `json:"email"`
	UserUUID    string `json:"user_uuid"`
	AccountUUID string `json:"account_uuid"`
	Shorthand   string `json:"shorthand"`
}

// Whoami returns the information about a signed-in account.
func (s *Service) Whoami(ctx context.Context) (*UserInfo, error) {
	data, err := s.run(ctx, []string{"whoami", "--format=json"})
	if err != nil {
		return nil, err
	}
	var info UserInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (s Service) Inject(ctx context.Context, tmpl string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "op", "inject")
	cmd.Stdin = strings.NewReader(tmpl)
	return cmd.Output()
}

func (s *Service) run(ctx context.Context, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "op", args...)
	return cmd.Output()
}
