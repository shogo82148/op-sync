package opsync

import (
	"context"
	"log/slog"

	"github.com/shogo82148/op-sync/internal/services/op"
)

func Run(ctx context.Context, args []string) int {
	if err := run(ctx, args); err != nil {
		slog.ErrorContext(ctx, "op-sync error", slog.String("error", err.Error()))
		return 1
	}
	return 0
}

func run(ctx context.Context, args []string) error {
	cfg, err := ParseConfig(".op-sync.yml")
	if err != nil {
		return err
	}

	op := op.NewService()
	planner := NewPlanner(&PlannerOptions{
		Config:   cfg,
		WhoAmIer: op,
		Injector: op,
	})
	plans, err := planner.Plan(ctx)
	if err != nil {
		return err
	}

	for _, plan := range plans {
		slog.InfoContext(ctx, plan.Preview())
	}

	// TODO: ask user to continue

	for _, plan := range plans {
		if err := plan.Apply(ctx); err != nil {
			return err
		}
	}

	return nil
}
