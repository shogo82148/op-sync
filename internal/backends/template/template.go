// Package template provides the template backend.
package template

import (
	"context"
	"fmt"
	"os"

	"github.com/shogo82148/op-sync/internal/backends"
	"github.com/shogo82148/op-sync/internal/maputils"
	"github.com/shogo82148/op-sync/internal/services"
)

var _ backends.Backend = (*Backend)(nil)

type Backend struct {
	opts *Options
}

type Options struct {
	services.Injector
}

func New(opts *Options) *Backend {
	return &Backend{opts: opts}
}

func (b *Backend) Plan(ctx context.Context, params map[string]any) ([]backends.Plan, error) {
	c := new(maputils.Context)
	output := maputils.Must[string](c, params, "output")
	template := maputils.Must[string](c, params, "template")
	if err := c.Err(); err != nil {
		return nil, fmt.Errorf("template: validation failed: %w", err)
	}

	var overwrite bool
	stat, err := os.Stat(output)
	if err == nil {
		if stat.IsDir() {
			return nil, fmt.Errorf("template: %q is a directory", output)
		}
		overwrite = true
	}
	return []backends.Plan{
		&Plan{
			output:    output,
			template:  template,
			overwrite: overwrite,
		},
	}, nil
}

var _ backends.Plan = (*Plan)(nil)

type Plan struct {
	output    string
	template  string
	overwrite bool
}

func (p *Plan) Preview() string {
	if p.overwrite {
		return fmt.Sprintf("file %q will be updated", p.output)
	}
	return fmt.Sprintf("file %q will be created", p.output)
}

func (p *Plan) Apply(ctx context.Context) error {
	return nil
}
