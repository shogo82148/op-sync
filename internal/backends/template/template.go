// Package template provides the template backend.
package template

import (
	"bytes"
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

	// inject the template
	newData, err := b.opts.Inject(ctx, template)
	if err != nil {
		return nil, err
	}

	var overwrite bool
	oldData, err := os.ReadFile(output)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		if bytes.Equal(oldData, newData) {
			return []backends.Plan{}, nil
		}
	}
	return []backends.Plan{
		&Plan{
			backend:   b,
			output:    output,
			newData:   newData,
			overwrite: overwrite,
		},
	}, nil
}

var _ backends.Plan = (*Plan)(nil)

type Plan struct {
	backend   *Backend
	output    string
	newData   []byte
	overwrite bool
}

func (p *Plan) Preview() string {
	if p.overwrite {
		return fmt.Sprintf("file %q will be updated", p.output)
	}
	return fmt.Sprintf("file %q will be created", p.output)
}

func (p *Plan) Apply(ctx context.Context) error {
	tmp := fmt.Sprintf("%s.%d.tmp", p.output, os.Getpid())
	defer os.Remove(tmp)
	if err := os.WriteFile(tmp, p.newData, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, p.output)
}
