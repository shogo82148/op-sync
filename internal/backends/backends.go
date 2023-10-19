package backends

import "context"

type Backend interface {
	Plan(ctx context.Context, cfg map[string]any) ([]Plan, error)
}

type Plan interface {
	Preview() string
	Apply(ctx context.Context) error
}
