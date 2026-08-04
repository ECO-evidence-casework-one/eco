package turnorchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCompileFailureUsesDeterministicFallback(t *testing.T) {
	d, l := baseDeps(t)
	d.Compiler = compilerFunc(func(context.Context, Request) (CompiledContext, error) {
		return CompiledContext{}, errors.New("secret internal failure")
	})
	r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-2", Text: "question"})
	if r.Outcome != OutcomeFallback || r.Text != "deterministic fallback" || r.Receipt.Reason != ReasonCompileFailed {
		t.Fatalf("unexpected: %+v", r)
	}
	if l.releases.Load() != 0 || !r.Receipt.ErasurePassed {
		t.Fatalf("unexpected cleanup: %+v", r.Receipt)
	}
}

func TestInvalidRequestDoesNotInvokeDependencies(t *testing.T) {
	d, _ := baseDeps(t)
	var called atomic.Bool
	d.Compiler = compilerFunc(func(context.Context, Request) (CompiledContext, error) {
		called.Store(true)
		return CompiledContext{}, nil
	})
	r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "bad id", Text: " "})
	if r.Outcome != OutcomeFailed || r.Receipt.Reason != ReasonInvalidRequest || called.Load() {
		t.Fatalf("unexpected: %+v called=%v", r, called.Load())
	}
}

func TestOutOfOrderStreamIsRejectedBeforeVerification(t *testing.T) {
	d, l := baseDeps(t)
	var verified atomic.Bool
	d.Runner = runnerFunc(func(_ context.Context, in GenerationInput, emit func(Chunk) error) (GenerationResult, error) {
		_ = emit(Chunk{TurnID: in.TurnID, RunID: in.RunID, Sequence: 1, Data: []byte("bad")})
		return GenerationResult{TurnID: in.TurnID, RunID: in.RunID, Completed: true}, nil
	})
	d.Verifier = verifierFunc(func(context.Context, VerificationInput) (VerifiedOutput, error) {
		verified.Store(true)
		return VerifiedOutput{}, nil
	})
	r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-3", Text: "question"})
	if r.Outcome != OutcomeFallback || r.Receipt.Reason != ReasonStreamSequenceViolation || verified.Load() {
		t.Fatalf("unexpected: %+v verified=%v", r, verified.Load())
	}
	if l.releases.Load() != 1 {
		t.Fatalf("release count %d", l.releases.Load())
	}
}

func TestIdentityMismatchAndOversizeFailClosed(t *testing.T) {
	cases := []struct {
		name  string
		chunk func(GenerationInput) Chunk
		want  ReasonCode
	}{
		{"identity", func(in GenerationInput) Chunk {
			return Chunk{TurnID: "other", RunID: in.RunID, Sequence: 0, Data: []byte("x")}
		}, ReasonStreamIdentityMismatch},
		{"chunk", func(in GenerationInput) Chunk {
			return Chunk{TurnID: in.TurnID, RunID: in.RunID, Sequence: 0, Data: make([]byte, 129)}
		}, ReasonChunkTooLarge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, _ := baseDeps(t)
			d.Runner = runnerFunc(func(_ context.Context, in GenerationInput, emit func(Chunk) error) (GenerationResult, error) {
				_ = emit(tc.chunk(in))
				return GenerationResult{TurnID: in.TurnID, RunID: in.RunID, Completed: true}, nil
			})
			r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-x", Text: "q"})
			if r.Outcome != OutcomeFallback || r.Receipt.Reason != tc.want {
				t.Fatalf("unexpected: %+v", r)
			}
		})
	}
}

func TestVerifierRejectionNeverPublishesGeneratedText(t *testing.T) {
	d, _ := baseDeps(t)
	d.Verifier = verifierFunc(func(context.Context, VerificationInput) (VerifiedOutput, error) {
		return VerifiedOutput{Accepted: false}, nil
	})
	r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-4", Text: "q"})
	if r.Outcome != OutcomeFallback || r.Text == "draft answer" || r.Receipt.Reason != ReasonVerificationRejected {
		t.Fatalf("unexpected: %+v", r)
	}
}

func TestCancellationStopsWithoutFallback(t *testing.T) {
	d, l := baseDeps(t)
	entered := make(chan struct{})
	d.Runner = runnerFunc(func(ctx context.Context, in GenerationInput, emit func(Chunk) error) (GenerationResult, error) {
		close(entered)
		<-ctx.Done()
		return GenerationResult{}, ctx.Err()
	})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan Result, 1)
	go func() { done <- newTestOrchestrator(t, d).Run(ctx, Request{TurnID: "turn-5", Text: "q"}) }()
	<-entered
	cancel()
	r := <-done
	if r.Outcome != OutcomeCancelled || r.Text != "" || r.Receipt.Reason != ReasonCancelled {
		t.Fatalf("unexpected: %+v", r)
	}
	if l.releases.Load() != 1 || !r.Receipt.ErasurePassed {
		t.Fatalf("cleanup failed: %+v", r.Receipt)
	}
}

func TestTimeoutStopsWithoutFallback(t *testing.T) {
	d, _ := baseDeps(t)
	d.Runner = runnerFunc(func(ctx context.Context, in GenerationInput, emit func(Chunk) error) (GenerationResult, error) {
		<-ctx.Done()
		return GenerationResult{}, ctx.Err()
	})
	r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-6", Text: "q", Timeout: 5 * time.Millisecond})
	if r.Outcome != OutcomeCancelled || r.Receipt.Reason != ReasonTimedOut {
		t.Fatalf("unexpected: %+v", r)
	}
}

func TestErasureFailureBlocksAcceptedAndFallbackOutput(t *testing.T) {
	for _, compileFails := range []bool{false, true} {
		t.Run(map[bool]string{false: "accepted", true: "fallback"}[compileFails], func(t *testing.T) {
			d, _ := baseDeps(t)
			if compileFails {
				d.Compiler = compilerFunc(func(context.Context, Request) (CompiledContext, error) { return CompiledContext{}, errors.New("x") })
			}
			d.Eraser = eraserFunc(func(context.Context, *Transient) error { return errors.New("erase failed") })
			r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-8", Text: "q"})
			if r.Outcome != OutcomeFailed || r.Text != "" || r.Receipt.Reason != ReasonErasureFailed {
				t.Fatalf("unexpected: %+v", r)
			}
		})
	}
}

func TestLeaseReleaseFailureBlocksOutput(t *testing.T) {
	d, l := baseDeps(t)
	l.err = errors.New("release failed")
	r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-9", Text: "q"})
	if r.Outcome != OutcomeFailed || r.Text != "" || r.Receipt.Reason != ReasonLeaseReleaseFailed {
		t.Fatalf("unexpected: %+v", r)
	}
	if l.releases.Load() != 1 {
		t.Fatalf("release count=%d", l.releases.Load())
	}
}

func TestNonDeterministicFallbackIsRejected(t *testing.T) {
	d, _ := baseDeps(t)
	d.Compiler = compilerFunc(func(context.Context, Request) (CompiledContext, error) { return CompiledContext{}, errors.New("x") })
	d.Fallback = fallbackFunc(func(context.Context, Request, Stage, ReasonCode) (FallbackOutput, error) {
		return FallbackOutput{Text: "maybe", FallbackID: "fb", Deterministic: false}, nil
	})
	r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-10", Text: "q"})
	if r.Outcome != OutcomeFailed || r.Receipt.Reason != ReasonInvalidFallback {
		t.Fatalf("unexpected: %+v", r)
	}
}

func TestDependencyPanicBecomesBoundedFailure(t *testing.T) {
	d, _ := baseDeps(t)
	d.Compiler = compilerFunc(func(context.Context, Request) (CompiledContext, error) { panic("sensitive panic") })
	r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-11", Text: "q"})
	if r.Outcome != OutcomeFallback || r.Receipt.Reason != ReasonDependencyPanic {
		t.Fatalf("unexpected: %+v", r)
	}
	raw, _ := json.Marshal(r)
	if strings.Contains(string(raw), "sensitive panic") {
		t.Fatal("panic text leaked")
	}
}

func TestOutputTotalLimitCancelsStream(t *testing.T) {
	d, _ := baseDeps(t)
	d.Router = routerFunc(func(context.Context, CompiledContext) (Route, error) {
		return Route{RouteID: "route-small", ModelSHA256: digest, RuntimeSHA256: digest, AdapterSHA256: digest, MaxOutputBytes: 10}, nil
	})
	d.Runner = runnerFunc(func(_ context.Context, in GenerationInput, emit func(Chunk) error) (GenerationResult, error) {
		if err := emit(Chunk{TurnID: in.TurnID, RunID: in.RunID, Sequence: 0, Data: []byte("123456")}); err != nil {
			return GenerationResult{}, err
		}
		_ = emit(Chunk{TurnID: in.TurnID, RunID: in.RunID, Sequence: 1, Data: []byte("789012")})
		return GenerationResult{TurnID: in.TurnID, RunID: in.RunID, Completed: true}, nil
	})
	r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-limit", Text: "q"})
	if r.Outcome != OutcomeFallback || r.Receipt.Reason != ReasonOutputTooLarge {
		t.Fatalf("unexpected: %+v", r)
	}
}

func TestInvalidContextRouteAndGenerationFailClosed(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Dependencies)
		want   ReasonCode
	}{
		{"context", func(d *Dependencies) {
			d.Compiler = compilerFunc(func(context.Context, Request) (CompiledContext, error) {
				return CompiledContext{TurnID: "other", ContextID: "ctx", Prompt: []byte("x")}, nil
			})
		}, ReasonInvalidContext},
		{"route", func(d *Dependencies) {
			d.Router = routerFunc(func(context.Context, CompiledContext) (Route, error) {
				return Route{RouteID: "route", ModelSHA256: "bad", RuntimeSHA256: digest, AdapterSHA256: digest, MaxOutputBytes: 10}, nil
			})
		}, ReasonInvalidRoute},
		{"generation", func(d *Dependencies) {
			d.Runner = runnerFunc(func(_ context.Context, in GenerationInput, emit func(Chunk) error) (GenerationResult, error) {
				_ = emit(Chunk{TurnID: in.TurnID, RunID: in.RunID, Sequence: 0, Data: []byte("x")})
				return GenerationResult{TurnID: "other", RunID: in.RunID, Completed: true}, nil
			})
		}, ReasonInvalidGeneration},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, _ := baseDeps(t)
			tc.mutate(&d)
			r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-invalid", Text: "q"})
			if r.Outcome != OutcomeFallback || r.Receipt.Reason != tc.want {
				t.Fatalf("unexpected: %+v", r)
			}
		})
	}
}

type panicIDLease struct{ releases atomic.Int32 }

func (*panicIDLease) ID() string                                  { panic("secret lease panic") }
func (l *panicIDLease) Release(context.Context, ReasonCode) error { l.releases.Add(1); return nil }

func TestLeaseIDPanicIsBoundedAndLeaseStillReleased(t *testing.T) {
	d, _ := baseDeps(t)
	l := &panicIDLease{}
	d.Supervisor = supervisorFunc(func(context.Context, Route) (Lease, error) { return l, nil })
	r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-idpanic", Text: "q"})
	if r.Outcome != OutcomeFallback || r.Receipt.Reason != ReasonDependencyPanic {
		t.Fatalf("unexpected: %+v", r)
	}
	if l.releases.Load() != 1 {
		t.Fatalf("lease not released: %d", l.releases.Load())
	}
}

func TestRunnerAndVerifierPanicsNeverEscape(t *testing.T) {
	for _, which := range []string{"runner", "verifier"} {
		t.Run(which, func(t *testing.T) {
			d, _ := baseDeps(t)
			if which == "runner" {
				d.Runner = runnerFunc(func(context.Context, GenerationInput, func(Chunk) error) (GenerationResult, error) {
					panic("runner secret")
				})
			} else {
				d.Verifier = verifierFunc(func(context.Context, VerificationInput) (VerifiedOutput, error) { panic("verifier secret") })
			}
			r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-panic-" + which, Text: "q"})
			if r.Outcome != OutcomeFallback || r.Receipt.Reason != ReasonDependencyPanic {
				t.Fatalf("unexpected: %+v", r)
			}
		})
	}
}
