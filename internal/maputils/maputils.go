package maputils

import (
	"errors"
	"fmt"
)

type Context struct {
	errors []error
}

func (ctx *Context) Err() error {
	return errors.Join(ctx.errors...)
}

func Get[T any](ctx *Context, m map[string]any, key string) (T, bool) {
	var zero T

	v, ok := m[key]
	if !ok {
		return zero, false
	}
	t, ok := v.(T)
	if !ok {
		ctx.errors = append(ctx.errors, fmt.Errorf("invalid type for parameter %q, want %T", key, zero))
		return zero, false
	}
	return t, true
}

func Must[T any](ctx *Context, m map[string]any, key string) T {
	v, ok := Get[T](ctx, m, key)
	if !ok {
		ctx.errors = append(ctx.errors, fmt.Errorf("parameter %q is required", key))
	}
	return v
}
