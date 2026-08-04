package turnorchestrator

import (
	"context"
	"testing"
	"time"
)

func TestNewRejectsIncompleteDependenciesAndBadPolicy(t *testing.T) {
	d, _ := baseDeps(t)
	d.Runner = nil
	if _, err := New(d, Policy{}); err == nil {
		t.Fatal("expected dependency error")
	}
	d, _ = baseDeps(t)
	if _, err := New(d, Policy{DefaultTimeout: 2 * time.Second, MaxTimeout: time.Second}); err == nil {
		t.Fatal("expected timeout policy error")
	}
	if _, err := New(d, Policy{MaxOutputBytes: 10, MaxChunkBytes: 11}); err == nil {
		t.Fatal("expected chunk policy error")
	}
}

func TestCleanupDependenciesReceiveBoundedContexts(t *testing.T) {
	d, l := baseDeps(t)
	var releaseDeadline, eraseDeadline bool
	d.Supervisor = supervisorFunc(func(context.Context, Route) (Lease, error) {
		return &deadlineLease{testLease: l, seen: &releaseDeadline}, nil
	})
	d.Eraser = eraserFunc(func(ctx context.Context, x *Transient) error { _, eraseDeadline = ctx.Deadline(); x.Zero(); return nil })
	o, err := New(d, Policy{DefaultTimeout: time.Second, MaxTimeout: 2 * time.Second, MaxRequestBytes: 1024, MaxPromptBytes: 4096, MaxOutputBytes: 1024, MaxChunkBytes: 128, CleanupTimeout: 50 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	r := o.Run(context.Background(), Request{TurnID: "turn-cleanup", Text: "q"})
	if r.Outcome != OutcomeAccepted || !releaseDeadline || !eraseDeadline {
		t.Fatalf("cleanup contexts not bounded: result=%+v release=%v erase=%v", r, releaseDeadline, eraseDeadline)
	}
}

type deadlineLease struct {
	testLease *testLease
	seen      *bool
}

func (l *deadlineLease) ID() string { return l.testLease.ID() }
func (l *deadlineLease) Release(ctx context.Context, r ReasonCode) error {
	_, *l.seen = ctx.Deadline()
	return l.testLease.Release(ctx, r)
}

func TestRequestTimeoutIsCappedByPolicy(t *testing.T) {
	d, _ := baseDeps(t)
	var remaining time.Duration
	d.Compiler = compilerFunc(func(ctx context.Context, r Request) (CompiledContext, error) {
		deadline, _ := ctx.Deadline()
		remaining = time.Until(deadline)
		return CompiledContext{TurnID: r.TurnID, ContextID: "ctx-cap", Prompt: []byte("x")}, nil
	})
	o, err := New(d, Policy{DefaultTimeout: 100 * time.Millisecond, MaxTimeout: 200 * time.Millisecond, MaxRequestBytes: 1024, MaxPromptBytes: 4096, MaxOutputBytes: 1024, MaxChunkBytes: 128})
	if err != nil {
		t.Fatal(err)
	}
	r := o.Run(context.Background(), Request{TurnID: "turn-cap", Text: "q", Timeout: time.Hour})
	if r.Outcome != OutcomeAccepted {
		t.Fatalf("unexpected: %+v", r)
	}
	if remaining <= 0 || remaining > 250*time.Millisecond {
		t.Fatalf("timeout not capped: %v", remaining)
	}
}
