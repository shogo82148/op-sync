package template

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/shogo82148/op-sync/internal/services/mock"
)

func TestPlan_NoFile(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dir := t.TempDir()
	tmp := filepath.Join(dir, "output.txt")

	b := New(&Options{
		Injector: mock.Injector(func(ctx context.Context, template string) ([]byte, error) {
			return []byte("template"), nil
		}),
	})

	// do planning
	plans, err := b.Plan(ctx, map[string]any{
		"output":   tmp,
		"template": "template",
	})
	if err != nil {
		t.Fatal(err)
	}

	// verify the plan
	if len(plans) != 1 {
		t.Fatalf("unexpected length: want 1, got %d", len(plans))
	}
	plan, ok := plans[0].(*Plan)
	if !ok {
		t.Fatalf("unexpected type: want *Plan, got %T", plans[0])
	}
	if plan.overwrite {
		t.Error("overwrite should be false")
	}
	if got, want := plan.Preview(), "file \""+tmp+"\" will be created"; got != want {
		t.Errorf("unexpected preview: want %q, got %q", want, got)
	}

	// apply the plan
	if err := plan.Apply(ctx); err != nil {
		t.Fatal(err)
	}

	// verify the output
	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "template" {
		t.Errorf("unexpected output: want %q, got %q", "template", string(data))
	}
}

func TestPlan_UpToDate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dir := t.TempDir()
	tmp := filepath.Join(dir, "output.txt")

	if err := os.WriteFile(tmp, []byte("template"), 0o600); err != nil {
		t.Fatal(err)
	}

	b := New(&Options{
		Injector: mock.Injector(func(ctx context.Context, template string) ([]byte, error) {
			return []byte("template"), nil
		}),
	})

	// do planning
	plans, err := b.Plan(ctx, map[string]any{
		"output":   tmp,
		"template": "template",
	})
	if err != nil {
		t.Fatal(err)
	}

	// verify the plan
	if len(plans) != 0 {
		t.Fatalf("unexpected length: want 0, got %d", len(plans))
	}
}

func TestPlan_Updated(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dir := t.TempDir()
	tmp := filepath.Join(dir, "output.txt")

	if err := os.WriteFile(tmp, []byte("old template"), 0o600); err != nil {
		t.Fatal(err)
	}

	b := New(&Options{
		Injector: mock.Injector(func(ctx context.Context, template string) ([]byte, error) {
			return []byte("template"), nil
		}),
	})

	// do planning
	plans, err := b.Plan(ctx, map[string]any{
		"output":   tmp,
		"template": "template",
	})
	if err != nil {
		t.Fatal(err)
	}

	// verify the plan
	if len(plans) != 1 {
		t.Fatalf("unexpected length: want 1, got %d", len(plans))
	}
	plan, ok := plans[0].(*Plan)
	if !ok {
		t.Fatalf("unexpected type: want *Plan, got %T", plans[0])
	}
	if !plan.overwrite {
		t.Error("overwrite should be true")
	}
	if got, want := plan.Preview(), "file \""+tmp+"\" will be updated"; got != want {
		t.Errorf("unexpected preview: want %q, got %q", want, got)
	}

	// apply the plan
	if err := plan.Apply(ctx); err != nil {
		t.Fatal(err)
	}

	// verify the output
	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "template" {
		t.Errorf("unexpected output: want %q, got %q", "template", string(data))
	}

}
