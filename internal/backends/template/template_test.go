package template

import (
	"context"
	"path/filepath"
	"testing"
)

func TestPlan(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dir := t.TempDir()
	tmp := filepath.Join(dir, "output.txt")

	b := New(&Options{})
	plans, err := b.Plan(ctx, map[string]any{
		"output":   tmp,
		"template": "template",
	})
	if err != nil {
		t.Fatal(err)
	}

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
}
