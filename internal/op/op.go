package op

import (
	"context"
	"encoding/json"
	"log"
	"os/exec"
	"strings"
	"time"
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

func (s *Service) Inject(ctx context.Context, tmpl string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "op", "inject")
	cmd.Stdin = strings.NewReader(tmpl)
	return cmd.Output()
}

func (s *Service) Read(ctx context.Context, uri string) ([]byte, error) {
	return s.run(ctx, []string{"read", "--no-newline", uri})
}

type ItemInfo struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Version int    `json:"version"`
	Vault   struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"vault"`
	Category              string    `json:"category"`
	LastEditedBy          string    `json:"last_edited_by"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	AdditionalInformation string    `json:"additional_information"`
}

func (s *Service) GetItem(ctx context.Context, vault, item string) (*ItemInfo, error) {
	data, err := s.run(ctx, []string{"item", "get", "--format=json", "--vault", vault, item})
	if err != nil {
		return nil, err
	}
	var info ItemInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

func (s *Service) run(ctx context.Context, args []string) ([]byte, error) {
	log.Println(args)
	cmd := exec.CommandContext(ctx, "op", args...)
	return cmd.Output()
}
