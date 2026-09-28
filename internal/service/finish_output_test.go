package service

import (
	"context"
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
		RegisterTolerance: 2.0,
	})
	if err != nil {
		t.Fatalf("new runtime: %v", err)
	}
	if err := runtime.Start(context.Background()); err != nil {
		t.Fatalf("start runtime: %v", err)
	}
	return runtime
}

func folderLots(t *testing.T, runtime *Runtime) int {
	t.Helper()
	for _, equipment := range runtime.Equipment() {
		if equipment["name"] == "folder" {
			lots, _ := equipment["lots"].(int)
			return lots
		}
	}
	t.Fatal("folder equipment snapshot is missing")
	return 0
}

func folderFault(t *testing.T, runtime *Runtime) string {
	t.Helper()
	for _, equipment := range runtime.Equipment() {
		if equipment["name"] == "folder" {
			fault, _ := equipment["fault"].(string)
			return fault
		}
	}
	t.Fatal("folder equipment snapshot is missing")
	return ""
}

func TestFinishOutputRejectsJobThatNeverProduced(t *testing.T) {
	runtime := newTestRuntime(t)

	if err := runtime.ExecuteOperation(context.Background(), "finish_output"); err == nil {
		t.Fatal("finish_output succeeded on a freshly created job with no roll, web feed, or released interval")
	} else if !strings.Contains(err.Error(), "finish output rejected") {
		t.Fatalf("rejection should explain the gate failure, got %q", err.Error())
	}

	if lots := folderLots(t, runtime); lots != 0 {
		t.Fatalf("no folder lot may be counted without production, got %d lots", lots)
	}
	fault := folderFault(t, runtime)
	if !strings.Contains(fault, "no roll has been loaded") {
		t.Fatalf("folder fault should preserve the rejection reason, got %q", fault)
	}
	incidents := runtime.Incidents()
	if len(incidents) != 1 || incidents[0].Component != "folder" {
		t.Fatalf("rejected finish_output should leave one folder incident, got %+v", incidents)
	}
}

func TestFinishOutputSucceedsAfterRequiredOperations(t *testing.T) {
	runtime := newTestRuntime(t)
	ctx := context.Background()

	for _, kind := range []string{"load_roll", "feed_web", "adjust_phase", "release_interval"} {
		if err := runtime.ExecuteOperation(ctx, kind); err != nil {
			t.Fatalf("operation %s failed: %v", kind, err)
		}
	}

	if err := runtime.ExecuteOperation(ctx, "finish_output"); err != nil {
		t.Fatalf("finish_output should succeed once production gates are met: %v", err)
	}
	if lots := folderLots(t, runtime); lots != 1 {
		t.Fatalf("expected exactly one completed lot, got %d", lots)
	}

	// The same released interval cannot back a second lot.
	err := runtime.ExecuteOperation(ctx, "finish_output")
	if err == nil {
		t.Fatal("finish_output succeeded twice against one released interval")
	}
	if !strings.Contains(err.Error(), "already produced a folder lot") {
		t.Fatalf("expected duplicate lot rejection, got %v", err)
	}
	if lots := folderLots(t, runtime); lots != 1 {
		t.Fatalf("rejected retry must not count another lot, got %d", lots)
	}
}

func TestFinishOutputRejectsMissingWebAndIntervalGates(t *testing.T) {
	runtime := newTestRuntime(t)
	ctx := context.Background()

	if err := runtime.ExecuteOperation(ctx, "load_roll"); err != nil {
		t.Fatalf("load_roll failed: %v", err)
	}
	if err := runtime.ExecuteOperation(ctx, "finish_output"); err == nil ||
		!strings.Contains(err.Error(), "paper has not advanced") {
		t.Fatalf("expected no-web-advance rejection, got %v", err)
	}

	if err := runtime.ExecuteOperation(ctx, "feed_web"); err != nil {
		t.Fatalf("feed_web failed: %v", err)
	}
	if err := runtime.ExecuteOperation(ctx, "adjust_phase"); err != nil {
		t.Fatalf("adjust_phase failed: %v", err)
	}
	if err := runtime.ExecuteOperation(ctx, "finish_output"); err == nil ||
		!strings.Contains(err.Error(), "has not released a paper interval") {
		t.Fatalf("expected missing interval rejection, got %v", err)
	}
	if lots := folderLots(t, runtime); lots != 0 {
		t.Fatalf("gate failures must never count a lot, got %d", lots)
	}
}
