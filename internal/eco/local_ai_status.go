package eco

import (
	"context"
	"strings"
	"time"
)

// LocalAIStatus is a user-facing summary of ECO's configured generative local
// AI path. It does not imply that a model answered a particular question; use
// QuestionUsedLocalAI for that distinction.
type LocalAIStatus struct {
	State          string `json:"state"`
	Configured     bool   `json:"configured"`
	Ready          bool   `json:"ready"`
	Engine         string `json:"engine,omitempty"`
	EngineVersion  string `json:"engine_version,omitempty"`
	RuntimeSHA256  string `json:"runtime_sha256,omitempty"`
	ModelName      string `json:"model_name,omitempty"`
	ModelSHA256    string `json:"model_sha256,omitempty"`
	Detail         string `json:"detail,omitempty"`
}

const localAIQuestionSupportPrefix = "Local llama.cpp selected the passages;"

// LocalAIStatus checks the same hash-pinned runtime/model configuration that
// Vault.Ask uses, then performs an offline llama.cpp version probe. It neither
// registers the tool nor mutates the workspace simply to report readiness.
func (v *Vault) LocalAIStatus() LocalAIStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return v.LocalAIStatusContext(ctx)
}

func (v *Vault) LocalAIStatusContext(ctx context.Context) LocalAIStatus {
	if ctx == nil {
		ctx = context.Background()
	}
	cfg, configured, err := loadConfiguredLocalAI()
	if !configured {
		return LocalAIStatus{
			State:      "unavailable",
			Configured: false,
			Ready:      false,
			Detail:     "No hash-pinned local Qwen/llama.cpp configuration is active. ECO will use its deterministic source-backed engine.",
		}
	}
	if err != nil {
		return LocalAIStatus{
			State:      "invalid",
			Configured: true,
			Ready:      false,
			Detail:     "The configured local AI files could not be verified: " + truncate(err.Error(), 500),
		}
	}
	if err := ctx.Err(); err != nil {
		return LocalAIStatus{
			State:      "unavailable",
			Configured: true,
			Ready:      false,
			Detail:     "Local AI readiness check did not complete: " + truncate(err.Error(), 300),
		}
	}
	version, err := llamaCPPVersion(ctx, cfg.Executable)
	if err != nil {
		return LocalAIStatus{
			State:         "invalid",
			Configured:    true,
			Ready:         false,
			Engine:        "llama.cpp",
			RuntimeSHA256: cfg.ExecutableSHA256,
			ModelName:     baseName(cfg.Model),
			ModelSHA256:   cfg.ModelSHA256,
			Detail:        "The hash-pinned local AI files are present, but llama.cpp did not pass its offline version check: " + truncate(err.Error(), 400),
		}
	}
	return LocalAIStatus{
		State:         "ready",
		Configured:    true,
		Ready:         true,
		Engine:        "llama.cpp",
		EngineVersion: version,
		RuntimeSHA256: cfg.ExecutableSHA256,
		ModelName:     baseName(cfg.Model),
		ModelSHA256:   cfg.ModelSHA256,
		Detail:        "Qwen is configured locally through a verified llama.cpp runtime. A question is not marked as model-used until ECO accepts a grounded local-AI result.",
	}
}

// QuestionUsedLocalAI reports whether this persisted QuestionRecord was released
// through ECO's grounded llama.cpp path. Deterministic fallback records do not
// carry this support marker.
func QuestionUsedLocalAI(rec QuestionRecord) bool {
	return strings.HasPrefix(strings.TrimSpace(rec.Support), localAIQuestionSupportPrefix)
}

func baseName(path string) string {
	path = strings.ReplaceAll(strings.TrimSpace(path), "\\", "/")
	if path == "" {
		return ""
	}
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}
