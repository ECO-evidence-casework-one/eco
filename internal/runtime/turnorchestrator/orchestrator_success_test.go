package turnorchestrator

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestAcceptedTurnIsVerifiedErasedAndReceipted(t *testing.T) {
	d, l := baseDeps(t)
	r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-1", Text: "Hello"})
	if r.Outcome != OutcomeAccepted || r.Text != "accepted answer" {
		t.Fatalf("unexpected result: %+v", r)
	}
	if l.releases.Load() != 1 || !r.Receipt.LeaseReleased || !r.Receipt.ErasurePassed {
		t.Fatalf("cleanup not proven: %+v", r.Receipt)
	}
	if r.Receipt.ChunkCount != 2 || r.Receipt.GeneratedBytes != 12 || r.Receipt.TokenCount != 2 {
		t.Fatalf("bad receipt counts: %+v", r.Receipt)
	}
	if r.Receipt.ReceiptSHA256 == "" || r.Receipt.Reason != ReasonNone {
		t.Fatalf("bad receipt: %+v", r.Receipt)
	}
	raw, _ := json.Marshal(r.Receipt)
	if strings.Contains(string(raw), "Hello") || strings.Contains(string(raw), "accepted answer") || strings.Contains(string(raw), "draft answer") {
		t.Fatalf("receipt leaked content: %s", raw)
	}
}

func TestLateChunksAreSuppressed(t *testing.T) {
	d, _ := baseDeps(t)
	lateDone := make(chan struct{})
	d.Runner = runnerFunc(func(_ context.Context, in GenerationInput, emit func(Chunk) error) (GenerationResult, error) {
		if err := emit(Chunk{TurnID: in.TurnID, RunID: in.RunID, Sequence: 0, Data: []byte("answer")}); err != nil {
			return GenerationResult{}, err
		}
		go func() {
			time.Sleep(2 * time.Millisecond)
			_ = emit(Chunk{TurnID: in.TurnID, RunID: in.RunID, Sequence: 1, Data: []byte("late")})
			close(lateDone)
		}()
		return GenerationResult{TurnID: in.TurnID, RunID: in.RunID, Completed: true, TokenCount: 1, FinishReason: "stop"}, nil
	})
	d.Verifier = verifierFunc(func(_ context.Context, in VerificationInput) (VerifiedOutput, error) {
		if string(in.Generated) != "answer" {
			t.Fatalf("late content entered verifier: %q", in.Generated)
		}
		<-lateDone
		return VerifiedOutput{Accepted: true, Text: "ok", VerificationID: "verify-late"}, nil
	})
	r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-7", Text: "q"})
	if r.Outcome != OutcomeAccepted || r.Text != "ok" {
		t.Fatalf("unexpected: %+v", r)
	}
	if r.Receipt.LateChunksSuppressed < 1 {
		t.Fatalf("late chunk not counted: %+v", r.Receipt)
	}
}

func TestConcurrentTurnsRemainIsolated(t *testing.T) {
	d, _ := baseDeps(t)
	d.Runner = runnerFunc(func(_ context.Context, in GenerationInput, emit func(Chunk) error) (GenerationResult, error) {
		for i := 0; i < 3; i++ {
			if err := emit(Chunk{TurnID: in.TurnID, RunID: in.RunID, Sequence: uint64(i), Data: []byte(in.TurnID)}); err != nil {
				return GenerationResult{}, err
			}
		}
		return GenerationResult{TurnID: in.TurnID, RunID: in.RunID, Completed: true, TokenCount: 3, FinishReason: "stop"}, nil
	})
	d.Supervisor = supervisorFunc(func(_ context.Context, r Route) (Lease, error) { return &testLease{id: "lease-" + r.RouteID}, nil })
	d.Verifier = verifierFunc(func(_ context.Context, in VerificationInput) (VerifiedOutput, error) {
		return VerifiedOutput{Accepted: true, Text: string(in.Generated), VerificationID: "verify-" + in.TurnID}, nil
	})
	o := newTestOrchestrator(t, d)
	ids := []string{"turn-A", "turn-B", "turn-C", "turn-D"}
	var wg sync.WaitGroup
	out := make(chan Result, len(ids))
	for _, id := range ids {
		wg.Add(1)
		go func(id string) { defer wg.Done(); out <- o.Run(context.Background(), Request{TurnID: id, Text: "q"}) }(id)
	}
	wg.Wait()
	close(out)
	seen := map[string]bool{}
	for r := range out {
		if r.Outcome != OutcomeAccepted || r.Text != r.Receipt.TurnID+r.Receipt.TurnID+r.Receipt.TurnID {
			t.Fatalf("cross-turn contamination: %+v", r)
		}
		seen[r.Receipt.TurnID] = true
	}
	if len(seen) != len(ids) {
		t.Fatalf("missing turns: %v", seen)
	}
}

func TestStageOrderIsFixed(t *testing.T) {
	d, _ := baseDeps(t)
	var mu sync.Mutex
	var stages []string
	add := func(s string) { mu.Lock(); stages = append(stages, s); mu.Unlock() }
	baseCompiler := d.Compiler
	baseRouter := d.Router
	baseSupervisor := d.Supervisor
	baseRunner := d.Runner
	baseVerifier := d.Verifier
	baseEraser := d.Eraser
	d.Compiler = compilerFunc(func(ctx context.Context, r Request) (CompiledContext, error) {
		add("compile")
		return baseCompiler.Compile(ctx, r)
	})
	d.Router = routerFunc(func(ctx context.Context, c CompiledContext) (Route, error) {
		add("route")
		return baseRouter.Route(ctx, c)
	})
	d.Supervisor = supervisorFunc(func(ctx context.Context, r Route) (Lease, error) {
		add("admit")
		l, err := baseSupervisor.Admit(ctx, r)
		if err != nil {
			return nil, err
		}
		return &orderedLease{Lease: l, add: add}, nil
	})
	d.Runner = runnerFunc(func(ctx context.Context, in GenerationInput, emit func(Chunk) error) (GenerationResult, error) {
		add("generate")
		return baseRunner.Generate(ctx, in, emit)
	})
	d.Verifier = verifierFunc(func(ctx context.Context, in VerificationInput) (VerifiedOutput, error) {
		add("verify")
		return baseVerifier.Verify(ctx, in)
	})
	d.Eraser = eraserFunc(func(ctx context.Context, x *Transient) error { add("erase"); return baseEraser.Erase(ctx, x) })
	r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-order", Text: "q"})
	if r.Outcome != OutcomeAccepted {
		t.Fatalf("unexpected: %+v", r)
	}
	got := strings.Join(stages, ",")
	want := "compile,route,admit,generate,release,verify,erase"
	if got != want {
		t.Fatalf("stage order=%q want=%q", got, want)
	}
}

type orderedLease struct {
	Lease
	add func(string)
}

func (l *orderedLease) Release(ctx context.Context, r ReasonCode) error {
	l.add("release")
	return l.Lease.Release(ctx, r)
}

func TestAllOwnedBuffersAreErased(t *testing.T) {
	d, _ := baseDeps(t)
	var retained *Transient
	var before [4]int
	d.Eraser = eraserFunc(func(_ context.Context, x *Transient) error {
		retained = x
		before = [4]int{len(x.Prompt), len(x.RuntimePrompt), len(x.Generated), len(x.VerificationCopy)}
		x.Zero()
		return nil
	})
	r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-zero", Text: "q"})
	if r.Outcome != OutcomeAccepted {
		t.Fatalf("unexpected: %+v", r)
	}
	for i, n := range before {
		if n == 0 {
			t.Fatalf("buffer %d was not populated before erasure: %v", i, before)
		}
	}
	if retained == nil || retained.Prompt != nil || retained.RuntimePrompt != nil || retained.Generated != nil || retained.VerificationCopy != nil {
		t.Fatalf("buffers retained: %+v", retained)
	}
}

func TestReceiptDigestRecomputesAndContainsNoContentFingerprints(t *testing.T) {
	d, _ := baseDeps(t)
	r := newTestOrchestrator(t, d).Run(context.Background(), Request{TurnID: "turn-receipt", Text: "unique secret phrase"})
	if r.Receipt.ReceiptSHA256 != receiptDigest(r.Receipt) {
		t.Fatalf("receipt digest mismatch")
	}
	raw, _ := json.Marshal(r.Receipt)
	s := string(raw)
	for _, forbidden := range []string{"unique secret phrase", "draft answer", "accepted answer", "request_sha256", "context_sha256", "generated_sha256", "output_sha256"} {
		if strings.Contains(s, forbidden) {
			t.Fatalf("receipt contains %q: %s", forbidden, s)
		}
	}
	if r.Receipt.RequestBytes != len("unique secret phrase") || r.Receipt.PromptBytes == 0 {
		t.Fatalf("bounded size metadata missing: %+v", r.Receipt)
	}
}
