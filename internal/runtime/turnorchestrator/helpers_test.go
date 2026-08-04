package turnorchestrator

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type fixedClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *fixedClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(time.Millisecond)
	return c.t
}

type compilerFunc func(context.Context, Request) (CompiledContext, error)

func (f compilerFunc) Compile(c context.Context, r Request) (CompiledContext, error) { return f(c, r) }

type routerFunc func(context.Context, CompiledContext) (Route, error)

func (f routerFunc) Route(c context.Context, x CompiledContext) (Route, error) { return f(c, x) }

type supervisorFunc func(context.Context, Route) (Lease, error)

func (f supervisorFunc) Admit(c context.Context, r Route) (Lease, error) { return f(c, r) }

type runnerFunc func(context.Context, GenerationInput, func(Chunk) error) (GenerationResult, error)

func (f runnerFunc) Generate(c context.Context, i GenerationInput, e func(Chunk) error) (GenerationResult, error) {
	return f(c, i, e)
}

type verifierFunc func(context.Context, VerificationInput) (VerifiedOutput, error)

func (f verifierFunc) Verify(c context.Context, i VerificationInput) (VerifiedOutput, error) {
	return f(c, i)
}

type fallbackFunc func(context.Context, Request, Stage, ReasonCode) (FallbackOutput, error)

func (f fallbackFunc) Resolve(c context.Context, r Request, s Stage, rc ReasonCode) (FallbackOutput, error) {
	return f(c, r, s, rc)
}

type eraserFunc func(context.Context, *Transient) error

func (f eraserFunc) Erase(c context.Context, x *Transient) error { return f(c, x) }

type testLease struct {
	id       string
	releases atomic.Int32
	reason   ReasonCode
	err      error
}

func (l *testLease) ID() string { return l.id }
func (l *testLease) Release(_ context.Context, r ReasonCode) error {
	l.releases.Add(1)
	l.reason = r
	return l.err
}

const digest = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func baseDeps(t *testing.T) (Dependencies, *testLease) {
	t.Helper()
	lease := &testLease{id: "lease-1"}
	d := Dependencies{
		Compiler: compilerFunc(func(_ context.Context, r Request) (CompiledContext, error) {
			return CompiledContext{TurnID: r.TurnID, ContextID: "ctx-1", Prompt: []byte("bounded prompt")}, nil
		}),
		Router: routerFunc(func(context.Context, CompiledContext) (Route, error) {
			return Route{RouteID: "route-1", ModelSHA256: digest, RuntimeSHA256: digest, AdapterSHA256: digest, MaxOutputBytes: 1024}, nil
		}),
		Supervisor: supervisorFunc(func(context.Context, Route) (Lease, error) { return lease, nil }),
		Runner: runnerFunc(func(_ context.Context, in GenerationInput, emit func(Chunk) error) (GenerationResult, error) {
			if err := emit(Chunk{TurnID: in.TurnID, RunID: in.RunID, Sequence: 0, Data: []byte("draft ")}); err != nil {
				return GenerationResult{}, err
			}
			if err := emit(Chunk{TurnID: in.TurnID, RunID: in.RunID, Sequence: 1, Data: []byte("answer")}); err != nil {
				return GenerationResult{}, err
			}
			return GenerationResult{TurnID: in.TurnID, RunID: in.RunID, Completed: true, TokenCount: 2, FinishReason: "stop"}, nil
		}),
		Verifier: verifierFunc(func(_ context.Context, in VerificationInput) (VerifiedOutput, error) {
			if string(in.Generated) != "draft answer" {
				t.Fatalf("unexpected generated: %q", in.Generated)
			}
			return VerifiedOutput{Accepted: true, Text: []byte("accepted answer"), VerificationID: "verify-1"}, nil
		}),
		Fallback: fallbackFunc(func(_ context.Context, _ Request, _ Stage, _ ReasonCode) (FallbackOutput, error) {
			return FallbackOutput{Text: []byte("deterministic fallback"), FallbackID: "fallback-1", Deterministic: true}, nil
		}),
		Eraser: ZeroEraser{},
		Clock:  &fixedClock{t: time.Date(2026, 8, 4, 9, 5, 0, 0, time.UTC)},
	}
	return d, lease
}

func newTestOrchestrator(t *testing.T, d Dependencies) *Orchestrator {
	t.Helper()
	o, err := New(d, Policy{DefaultTimeout: time.Second, MaxTimeout: 2 * time.Second, MaxRequestBytes: 1024, MaxPromptBytes: 4096, MaxOutputBytes: 1024, MaxChunkBytes: 128})
	if err != nil {
		t.Fatal(err)
	}
	return o
}
