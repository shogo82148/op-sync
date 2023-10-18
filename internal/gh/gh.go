package gh

import (
	"bytes"
	"context"
	"os/exec"
	"sync"

	"github.com/google/go-github/v56/github"
)

type Service struct {
	mu sync.Mutex
	c  *github.Client
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) client(ctx context.Context) (*github.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.c != nil {
		return s.c, nil
	}

	cmd := exec.CommandContext(ctx, "gh", "auth", "token")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	token := string(bytes.TrimSpace(out))
	s.c = github.NewClient(nil).WithAuthToken(token)
	return s.c, nil
}

func (s *Service) UserInfo(ctx context.Context) (string, error) {
	client, err := s.client(ctx)
	if err != nil {
		return "", err
	}

	u, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}
	return u.GetLogin(), nil
}
