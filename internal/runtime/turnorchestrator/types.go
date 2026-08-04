// Package turnorchestrator coordinates one bounded local-AI turn.
//
// It deliberately owns no model, worker, evidence, workspace or UI behaviour.
// Those capabilities are injected through narrow interfaces so existing
// foundations remain authoritative and the application receives one auditable
// transaction boundary.
package turnorchestrator

import (
	"context"
	"time"
)

type Stage string

const (
	StageValidate Stage = "validate"
	StageCompile  Stage = "compile"
	StageRoute    Stage = "route"
	StageAdmit    Stage = "admit"
	StageGenerate Stage = "generate"
	StageRelease  Stage = "release"
	StageVerify   Stage = "verify"
	StageErase    Stage = "erase"
	StageFallback Stage = "fallback"
	StageComplete Stage = "complete"
)

type Outcome string

const (
	OutcomeAccepted  Outcome = "accepted"
	OutcomeFallback  Outcome = "fallback"
	OutcomeCancelled Outcome = "cancelled"
	OutcomeFailed    Outcome = "failed"
)

type ReasonCode string

const (
	ReasonNone                    ReasonCode = "none"
	ReasonInvalidRequest          ReasonCode = "invalid_request"
	ReasonCancelled               ReasonCode = "cancelled"
	ReasonTimedOut                ReasonCode = "timed_out"
	ReasonCompileFailed           ReasonCode = "compile_failed"
	ReasonInvalidContext          ReasonCode = "invalid_context"
	ReasonRouteFailed             ReasonCode = "route_failed"
	ReasonInvalidRoute            ReasonCode = "invalid_route"
	ReasonAdmissionDenied         ReasonCode = "admission_denied"
	ReasonGenerationFailed        ReasonCode = "generation_failed"
	ReasonInvalidGeneration       ReasonCode = "invalid_generation"
	ReasonRunIdentityFailed       ReasonCode = "run_identity_failed"
	ReasonStreamIdentityMismatch  ReasonCode = "stream_identity_mismatch"
	ReasonStreamSequenceViolation ReasonCode = "stream_sequence_violation"
	ReasonChunkTooLarge           ReasonCode = "chunk_too_large"
	ReasonOutputTooLarge          ReasonCode = "output_too_large"
	ReasonLeaseReleaseFailed      ReasonCode = "lease_release_failed"
	ReasonVerificationFailed      ReasonCode = "verification_failed"
	ReasonVerificationRejected    ReasonCode = "verification_rejected"
	ReasonInvalidVerifiedOutput   ReasonCode = "invalid_verified_output"
	ReasonFallbackFailed          ReasonCode = "fallback_failed"
	ReasonInvalidFallback         ReasonCode = "invalid_fallback"
	ReasonErasureFailed           ReasonCode = "erasure_failed"
	ReasonDependencyPanic         ReasonCode = "dependency_panic"
)

type Request struct {
	TurnID  string
	Text    string
	Timeout time.Duration
}

type CompiledContext struct {
	TurnID    string
	ContextID string
	Prompt    []byte
}

type Route struct {
	RouteID        string
	ModelSHA256    string
	RuntimeSHA256  string
	AdapterSHA256  string
	MaxOutputBytes int
}

type GenerationInput struct {
	TurnID string
	RunID  string
	Prompt []byte
	Route  Route
}

type Chunk struct {
	TurnID   string
	RunID    string
	Sequence uint64
	Data     []byte
}

type GenerationResult struct {
	TurnID       string
	RunID        string
	Completed    bool
	TokenCount   int
	FinishReason string
}

type VerificationInput struct {
	TurnID    string
	ContextID string
	Route     Route
	Generated []byte
}

type VerifiedOutput struct {
	Accepted       bool
	Text           []byte
	VerificationID string
}

type FallbackOutput struct {
	Text          []byte
	FallbackID    string
	Deterministic bool
}

type Result struct {
	Outcome Outcome
	Text    string
	Receipt Receipt
}

type Receipt struct {
	Schema      int        `json:"schema"`
	TurnID      string     `json:"turn_id"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt time.Time  `json:"completed_at"`
	Outcome     Outcome    `json:"outcome"`
	FinalStage  Stage      `json:"final_stage"`
	Reason      ReasonCode `json:"reason"`

	RouteSHA256 string `json:"route_sha256,omitempty"`

	ContextID      string `json:"context_id,omitempty"`
	RouteID        string `json:"route_id,omitempty"`
	ModelSHA256    string `json:"model_sha256,omitempty"`
	RuntimeSHA256  string `json:"runtime_sha256,omitempty"`
	AdapterSHA256  string `json:"adapter_sha256,omitempty"`
	LeaseID        string `json:"lease_id,omitempty"`
	RunID          string `json:"run_id,omitempty"`
	VerificationID string `json:"verification_id,omitempty"`
	FallbackID     string `json:"fallback_id,omitempty"`

	RequestBytes   int    `json:"request_bytes"`
	PromptBytes    int    `json:"prompt_bytes"`
	ChunkCount     int    `json:"chunk_count"`
	GeneratedBytes int    `json:"generated_bytes"`
	OutputBytes    int    `json:"output_bytes"`
	TokenCount     int    `json:"token_count"`
	FinishReason   string `json:"finish_reason,omitempty"`

	LeaseReleased        bool   `json:"lease_released"`
	ErasurePassed        bool   `json:"erasure_passed"`
	LateChunksSuppressed int    `json:"late_chunks_suppressed"`
	ReceiptSHA256        string `json:"receipt_sha256"`
}

type ContextCompiler interface {
	Compile(context.Context, Request) (CompiledContext, error)
}

type Router interface {
	Route(context.Context, CompiledContext) (Route, error)
}

type Lease interface {
	ID() string
	Release(context.Context, ReasonCode) error
}

type Supervisor interface {
	Admit(context.Context, Route) (Lease, error)
}

type Runner interface {
	Generate(context.Context, GenerationInput, func(Chunk) error) (GenerationResult, error)
}

type Verifier interface {
	Verify(context.Context, VerificationInput) (VerifiedOutput, error)
}

type Fallback interface {
	Resolve(context.Context, Request, Stage, ReasonCode) (FallbackOutput, error)
}

type Eraser interface {
	Erase(context.Context, *Transient) error
}

type Clock interface {
	Now() time.Time
}

type Dependencies struct {
	Compiler   ContextCompiler
	Router     Router
	Supervisor Supervisor
	Runner     Runner
	Verifier   Verifier
	Fallback   Fallback
	Eraser     Eraser
	Clock      Clock
}

type Policy struct {
	DefaultTimeout  time.Duration
	MaxTimeout      time.Duration
	MaxRequestBytes int
	MaxPromptBytes  int
	MaxOutputBytes  int
	MaxChunkBytes   int
	CleanupTimeout  time.Duration
}

type Transient struct {
	Prompt           []byte
	RuntimePrompt    []byte
	Generated        []byte
	VerificationCopy []byte
	Verified         []byte
}

// Zero overwrites all owned transient byte slices and clears their lengths.
func (t *Transient) Zero() {
	if t == nil {
		return
	}
	for i := range t.Prompt {
		t.Prompt[i] = 0
	}
	for i := range t.RuntimePrompt {
		t.RuntimePrompt[i] = 0
	}
	for i := range t.Generated {
		t.Generated[i] = 0
	}
	for i := range t.VerificationCopy {
		t.VerificationCopy[i] = 0
	}
	for i := range t.Verified {
		t.Verified[i] = 0
	}
	t.Prompt = nil
	t.RuntimePrompt = nil
	t.Generated = nil
	t.VerificationCopy = nil
	t.Verified = nil
}

type ZeroEraser struct{}

func (ZeroEraser) Erase(_ context.Context, t *Transient) error {
	t.Zero()
	return nil
}
