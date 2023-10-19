package backends

import "context"

// Backend is a backend of op-sync.
type Backend interface {
	Plan(ctx context.Context, cfg map[string]any) ([]Plan, error)
}

// Plan is a plan of op-sync.
type Plan interface {
	Preview() string
	Apply(ctx context.Context) error
}
