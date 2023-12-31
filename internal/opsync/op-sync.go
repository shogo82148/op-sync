package opsync

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/Songmu/prompter"
	"github.com/shogo82148/op-sync/internal/backends"
	"github.com/shogo82148/op-sync/internal/services/awssecretsmanager"
	"github.com/shogo82148/op-sync/internal/services/awsssm"
	"github.com/shogo82148/op-sync/internal/services/awssts"
	"github.com/shogo82148/op-sync/internal/services/gh"
	"github.com/shogo82148/op-sync/internal/services/op"
)

var app = New()

type App struct {
	// Config is file path.
	Config string

	// Help shows help message.
	Help bool

	// Debug enables debug log.
	Debug bool

	// Force enables force mode.
	Force bool

	// Type is the type of the secret to sync.
	Type string

	fset *flag.FlagSet
}

func New() *App {
	fset := flag.NewFlagSet("op-sync", flag.ExitOnError)
	app := &App{
		fset: fset,
	}
	fset.StringVar(&app.Config, "config", ".op-sync.yml", "config file path")
	fset.BoolVar(&app.Help, "help", false, "show help message")
	fset.BoolVar(&app.Debug, "debug", false, "enable debug log")
	fset.BoolVar(&app.Force, "force", false, "enable force mode")
	fset.StringVar(&app.Type, "type", "", "the type of the secret to sync")
	return app
}

func Run(ctx context.Context, args []string) int {
	if err := app.Parse(args); err != nil {
		slog.ErrorContext(ctx, "op-sync error", slog.String("error", err.Error()))
	}
	if err := app.Run(ctx); err != nil {
		slog.ErrorContext(ctx, "op-sync error", slog.String("error", err.Error()))
		return 1
	}
	return 0
}

// Parse parses command line arguments.
func (app *App) Parse(args []string) error {
	return app.fset.Parse(args)
}

// Run runs the application.
func (app *App) Run(ctx context.Context) error {
	if app.Help {
		// show help message
		app.fset.Usage()
		return nil
	}

	if app.Debug {
		// enable debug log
		h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		slog.SetDefault(slog.New(h))
	}

	// parse configure file
	slog.DebugContext(ctx, "parse config", slog.String("path", app.Config))
	cfg, err := ParseConfig(app.Config)
	if err != nil {
		return fmt.Errorf("failed to parse %q: %w", app.Config, err)
	}

	planner := NewPlanner(&PlannerOptions{
		Config:            cfg,
		OnePassword:       op.NewService(),
		GitHub:            gh.NewService(),
		AWSSTS:            awssts.New(),
		AWSSSM:            awsssm.New(),
		AWSSecretsManager: awssecretsmanager.New(),
	})

	var plans []backends.Plan
	if app.Type != "" {
		plans, err = planner.PlanWithType(ctx, app.Type)
	} else if app.fset.NArg() == 0 {
		plans, err = planner.Plan(ctx)
	} else {
		plans, err = planner.PlanWithSecrets(ctx, app.fset.Args())
	}
	if err != nil {
		return err
	}

	if len(plans) == 0 {
		fmt.Println("No changes will be applied.")
		return nil
	}

	fmt.Println("The following changes will be applied:")
	for _, plan := range plans {
		fmt.Println(plan.Preview())
	}

	if !app.Force {
		if !prompter.YN("Do you want to continue?", false) {
			return nil
		}
	}

	for _, plan := range plans {
		if err := plan.Apply(ctx); err != nil {
			return err
		}
	}

	return nil
}
