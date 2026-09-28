package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func newTestRuntime(t *testing.T) *Runtime {
	t.Helper()
	runtime, err := NewRuntime(Config{
		DataDir:           t.TempDir(),
		FestoonCapacity:   120,
		RegisterGap:       500 * time.Millisecond,
		RegisterTolerance: 0.25,
	})
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatalf("start runtime: %v", err)
	}
	return runtime
}

func completedLots(runtime *Runtime) int {
	_, _, _, _, lots := runtime.folder.Snapshot()
	return lots
}

func TestFinishOutputRejectedForFreshJob(t *testing.T) {
	runtime := newTestRuntime(t)
	if _, ok := runtime.CurrentJob(); !ok {
		t.Fatal("expected an active job")
	}

	err := runtime.ExecuteOperation(context.Background(), "finish_output")
	if err == nil {
		t.Fatal("finish_output on a fresh job must be rejected")
	}
	var gateErr *FinishGateError
	if !errors.As(err, &gateErr) {
		t.Fatalf("expected *FinishGateError, got %T: %v", err, err)
	}
	if len(gateErr.Reasons) == 0 {
		t.Fatal("rejection must retain at least one reason")
	}

	join := strings.Join(gateErr.Reasons, "; ")
	if !strings.Contains(join, "roll") {
		t.Fatalf("rejection must mention the missing roll, got %q", join)
	}
	if !strings.Contains(join, "interval") {
		t.Fatalf("rejection must mention the missing released interval, got %q", join)
	}
	if completedLots(runtime) != 0 {
		t.Fatalf("rejected finish must not create a lot, got %d", completedLots(runtime))
	}
	if ready, _, _, _, _ := runtime.folder.Snapshot(); ready {
		t.Fatal("folder must not be left ready after a rejected finish")
	}

	incidents := runtime.Incidents()
	if len(incidents) == 0 || incidents[0].Component != "folder" {
		t.Fatalf("rejected finish must record a folder incident, got %#v", incidents)
	}

	// The rejected attempt must not have changed the active job.
	if _, ok := runtime.CurrentJob(); !ok {
		t.Fatal("job must remain active after rejection")
	}
}

func TestFinishOutputSucceedsAfterProductionSteps(t *testing.T) {
	runtime := newTestRuntime(t)
	ctx := context.Background()
	steps := []string{
		"load_roll",
		"feed_web",
		"adjust_phase",
		"start_dryer",
		"release_interval",
	}
	for _, kind := range steps {
		if err := runtime.ExecuteOperation(ctx, kind); err != nil {
			t.Fatalf("operation %s failed: %v", kind, err)
		}
	}

	if err := runtime.ExecuteOperation(ctx, "finish_output"); err != nil {
		t.Fatalf("finish_output after required operations must succeed: %v", err)
	}
	if completedLots(runtime) != 1 {
		t.Fatalf("normal completion must create exactly one lot, got %d", completedLots(runtime))
	}
	if ready, _, _, _, _ := runtime.folder.Snapshot(); !ready {
		t.Fatal("folder must remain ready after a normal completion")
	}
}

func TestFinishOutputRejectedWithoutReleasedInterval(t *testing.T) {
	runtime := newTestRuntime(t)
	ctx := context.Background()
	for _, kind := range []string{"load_roll", "feed_web"} {
		if err := runtime.ExecuteOperation(ctx, kind); err != nil {
			t.Fatalf("operation %s failed: %v", kind, err)
		}
	}

	err := runtime.ExecuteOperation(ctx, "finish_output")
	if err == nil {
		t.Fatal("finish_output before interval release must be rejected")
	}
	var gateErr *FinishGateError
	if !errors.As(err, &gateErr) {
		t.Fatalf("expected *FinishGateError, got %T: %v", err, err)
	}
	join := strings.Join(gateErr.Reasons, "; ")
	if !strings.Contains(join, "interval") {
		t.Fatalf("rejection must name the missing released interval, got %q", join)
	}
	if strings.Contains(join, "roll") || strings.Contains(join, "web") {
		t.Fatalf("roll and web evidence are present, reasons were %q", join)
	}
	if completedLots(runtime) != 0 {
		t.Fatalf("rejected finish must not create a lot, got %d", completedLots(runtime))
	}
}
