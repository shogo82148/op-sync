package op

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/shogo82148/op-sync/internal/services"
)

type URI struct {
	Vault     string
	Item      string
	Section   string
	Field     string
	Attribute string
}

func (u *URI) String() string {
	var path string
	if u.Section == "" {
		path = fmt.Sprintf("%s/%s", u.Item, u.Field)
	} else {
		path = fmt.Sprintf("%s/%s/%s", u.Item, u.Section, u.Field)
	}

	q := url.Values{}
	if u.Attribute != "" {
		q.Set("attribute", u.Attribute)
	}

	v := url.URL{
		Scheme:   "op",
		Host:     u.Vault,
		Path:     path,
		RawQuery: q.Encode(),
	}
	return v.String()
}

func ParseURI(uri string) (*URI, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "op" {
		return nil, fmt.Errorf("unknown scheme: %q", u.Scheme)
	}

	path := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	var item, section, field, attribute string
	if len(path) == 2 {
		item = path[0]
		field = path[1]
	} else if len(path) == 3 {
		item = path[0]
		section = path[1]
		field = path[2]
	} else {
		return nil, fmt.Errorf("invalid path: %q", u.Path)
	}
	attribute = u.Query().Get("attribute")

	return &URI{
		Vault:     u.Host,
		Item:      item,
		Section:   section,
		Field:     field,
		Attribute: attribute,
	}, nil
}

func command(ctx context.Context, name string, args ...string) *exec.Cmd {
	slog.DebugContext(ctx, "run 1password cli", slog.String("name", name), slog.Any("args", args))
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Cancel = func() error {
		return cmd.Process.Signal(os.Interrupt)
	}
	cmd.WaitDelay = 10 * time.Second
	return cmd
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
