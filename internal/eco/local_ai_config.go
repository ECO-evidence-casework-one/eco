package eco

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	envLlamaCPPExecutable       = "ECO_LLAMA_CPP"
	envLlamaCPPExecutableSHA256 = "ECO_LLAMA_CPP_SHA256"
	envLlamaCPPModel            = "ECO_LLAMA_MODEL"
	envLlamaCPPModelSHA256      = "ECO_LLAMA_MODEL_SHA256"
)

type configuredLocalAI struct {
	Executable       string
	ExecutableSHA256 string
	Model            string
	ModelSHA256      string
}

// Ask is ECO's application-facing question route. When a complete, explicitly
// hash-pinned local llama.cpp configuration is present it uses the already
// grounded local-AI workflow through ECO's verified local-tool registry. With
// no local-AI configuration it preserves the deterministic source-backed
// behaviour. A configured engine that cannot be verified or run is audited and
// falls back rather than making Ask ECO unusable.
func (v *Vault) Ask(question string, scopeIDs []string) QuestionRecord {
	v.opMu.RLock()
	defer v.opMu.RUnlock()
	return v.askLocked(question, append([]string(nil), scopeIDs...))
}

// askLocked owns the complete Ask transaction beneath one opMu read lock.
func (v *Vault) askLocked(question string, scopeIDs []string) QuestionRecord {
	if workspaceOnlyIntent(classifyIntent(strings.TrimSpace(question))) {
		return v.askDeterministic(question, scopeIDs)
	}
	cfg, configured, err := loadConfiguredLocalAI()
	if !configured {
		return v.askDeterministic(question, scopeIDs)
	}
	if err == nil {
		err = v.ensureConfiguredLlamaCPPRegistered(cfg)
	}
	var result LlamaCPPAnswerResult
	if err == nil {
		result, err = v.askWithRegisteredLlamaCPPLocked(context.Background(), question, scopeIDs, cfg.Model)
		if err == nil && result.Question.ID != "" {
			return result.Question
		}
		if err == nil {
			err = errors.New("configured local AI returned no accepted question record")
		}
	}
	_ = v.recordConfiguredLocalAIFallback(question, err)
	remaining := maxAskVerificationBytes - result.sourceVerificationBytes
	if remaining < 0 {
		remaining = 0
	}
	return v.askDeterministicWithBudget(question, scopeIDs, remaining, result.sourceVerificationFailures, result.sourceVerificationLimit || remaining == 0)
}

func loadConfiguredLocalAI() (configuredLocalAI, bool, error) {
	cfg := configuredLocalAI{
		Executable:       strings.TrimSpace(os.Getenv(envLlamaCPPExecutable)),
		ExecutableSHA256: normalizeSHA256(os.Getenv(envLlamaCPPExecutableSHA256)),
		Model:            strings.TrimSpace(os.Getenv(envLlamaCPPModel)),
		ModelSHA256:      normalizeSHA256(os.Getenv(envLlamaCPPModelSHA256)),
	}
	if cfg.Executable == "" && cfg.ExecutableSHA256 == "" && cfg.Model == "" && cfg.ModelSHA256 == "" {
		return configuredLocalAI{}, false, nil
	}
	if cfg.Executable == "" || cfg.ExecutableSHA256 == "" || cfg.Model == "" || cfg.ModelSHA256 == "" {
		return configuredLocalAI{}, true, errors.New("local AI configuration is incomplete; executable, executable SHA-256, GGUF model and model SHA-256 are all required")
	}
	if !validSHA256(cfg.ExecutableSHA256) || !validSHA256(cfg.ModelSHA256) {
		return configuredLocalAI{}, true, errors.New("local AI configuration contains an invalid SHA-256 identity")
	}

	executable, err := requireAbsoluteRegularFile(cfg.Executable, "configured llama.cpp executable")
	if err != nil {
		return configuredLocalAI{}, true, err
	}
	executableHash, err := hashFile(executable)
	if err != nil {
		return configuredLocalAI{}, true, fmt.Errorf("fingerprint configured llama.cpp executable: %w", err)
	}
	if !strings.EqualFold(executableHash, cfg.ExecutableSHA256) {
		return configuredLocalAI{}, true, errors.New("configured llama.cpp executable SHA-256 does not match its approved identity")
	}

	model, err := inspectLlamaCPPModel(cfg.Model)
	if err != nil {
		return configuredLocalAI{}, true, err
	}
	if !strings.EqualFold(model.SHA256, cfg.ModelSHA256) {
		return configuredLocalAI{}, true, errors.New("configured llama.cpp GGUF model SHA-256 does not match its approved identity")
	}

	cfg.Executable = executable
	cfg.ExecutableSHA256 = strings.ToLower(executableHash)
	cfg.Model = model.Path
	cfg.ModelSHA256 = strings.ToLower(model.SHA256)
	return cfg, true, nil
}

// ensureConfiguredLlamaCPPRegistered keeps the application-facing AI route on
// the same donor provenance/identity path as ECO's other optional local tools.
// A matching current registration is re-verified. A missing, stale or different
// registration is replaced only by successfully registering the explicitly
// hash-pinned executable from the process configuration.
func (v *Vault) ensureConfiguredLlamaCPPRegistered(cfg configuredLocalAI) error {
	registration, err := v.RegisteredLocalTool("llama.cpp")
	if err == nil && registration.SHA256 == cfg.ExecutableSHA256 {
		verified, verifyErr := v.VerifyRegisteredLocalTool("llama.cpp")
		if verifyErr == nil && verified.SHA256 == cfg.ExecutableSHA256 {
			return nil
		}
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read registered llama.cpp runtime: %w", err)
	}

	registration, err = v.RegisterLocalTool("llama.cpp", cfg.Executable)
	if err != nil {
		return fmt.Errorf("register configured llama.cpp runtime: %w", err)
	}
	if registration.SHA256 != cfg.ExecutableSHA256 {
		return errors.New("registered llama.cpp runtime does not match the configured approved SHA-256 identity")
	}
	return nil
}

func normalizeSHA256(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

func (v *Vault) recordConfiguredLocalAIFallback(question string, cause error) error {
	reason := "configured local AI did not produce an accepted grounded answer"
	if cause != nil {
		reason = truncate(cause.Error(), 500)
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	oldChanges := append([]ChangeRecord(nil), v.Workspace.Changes...)
	oldUpdatedAt := v.Workspace.UpdatedAt
	oldBuildID := v.Workspace.BuildID
	v.addChangeUnlocked("local-ai", "configured-local-ai-fallback", "Configured local AI failed; ECO used its deterministic source-backed fallback", map[string]any{
		"question": truncate(strings.TrimSpace(question), 300),
		"reason":   reason,
	})
	if err := v.saveUnlocked(); err != nil {
		v.Workspace.Changes = oldChanges
		v.Workspace.UpdatedAt = oldUpdatedAt
		v.Workspace.BuildID = oldBuildID
		return err
	}
	return nil
}
