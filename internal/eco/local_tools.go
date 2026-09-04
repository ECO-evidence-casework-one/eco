package eco

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	localToolRegistrationChangeType = "local-tool-registered"
	maxLocalToolPathRunes           = 32768
	maxLocalToolIdentityRunes       = 512
)

type localToolSpec struct {
	Kind     string
	Upstream string
	License  string
}

var localToolSpecs = map[string]localToolSpec{
	"tesseract": {Kind: "tesseract", Upstream: "tesseract-ocr/tesseract", License: "Apache-2.0"},
	"docling":   {Kind: "docling", Upstream: "docling-project/docling", License: "MIT"},
	"ocrmypdf":  {Kind: "ocrmypdf", Upstream: "ocrmypdf/OCRmyPDF", License: "MPL-2.0"},
	"llama.cpp": {Kind: "llama.cpp", Upstream: "ggml-org/llama.cpp", License: "MIT"},
	"pdfcpu":    {Kind: "pdfcpu", Upstream: "pdfcpu/pdfcpu", License: "Apache-2.0"},
}

type LocalToolRegistration struct {
	Kind          string    `json:"kind"`
	Executable    string    `json:"executable"`
	SHA256        string    `json:"sha256"`
	Size          int64     `json:"size"`
	Version       string    `json:"version"`
	Upstream      string    `json:"upstream"`
	License       string    `json:"license"`
	RegisteredAt  time.Time `json:"registered_at"`
	AuditChangeID string    `json:"audit_change_id,omitempty"`
}

type localToolVersionProbe func(context.Context, string, string) (string, error)

func canonicalLocalToolKind(kind string) (localToolSpec, error) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	switch kind {
	case "tesseract", "tesseract-ocr":
		kind = "tesseract"
	case "docling":
		kind = "docling"
	case "ocrmypdf", "ocrmypdf-cli", "ocrmypdf.exe":
		kind = "ocrmypdf"
	case "llama.cpp", "llamacpp", "llama-cli":
		kind = "llama.cpp"
	case "pdfcpu":
		kind = "pdfcpu"
	default:
		return localToolSpec{}, fmt.Errorf("unsupported local tool kind %q", strings.TrimSpace(kind))
	}
	spec, ok := localToolSpecs[kind]
	if !ok {
		return localToolSpec{}, fmt.Errorf("unsupported local tool kind %q", kind)
	}
	return spec, nil
}

func localToolVersion(ctx context.Context, kind, executable string) (string, error) {
	spec, err := canonicalLocalToolKind(kind)
	if err != nil {
		return "", err
	}
	switch spec.Kind {
	case "tesseract":
		return tesseractVersion(ctx, executable)
	case "docling":
		return doclingVersion(ctx, executable)
	case "ocrmypdf":
		return ocrmyPDFVersion(ctx, executable)
	case "llama.cpp":
		return llamaCPPVersion(ctx, executable)
	case "pdfcpu":
		return pdfcpuVersion(ctx, executable)
	default:
		return "", errors.New("local tool version probe is unavailable")
	}
}

func (v *Vault) RegisterLocalTool(ctx context.Context, kind, executable string) (LocalToolRegistration, error) {
	return v.registerLocalToolWithProbe(ctx, kind, executable, localToolVersion)
}

func (v *Vault) registerLocalToolWithProbe(ctx context.Context, kind, executable string, probe localToolVersionProbe) (LocalToolRegistration, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return LocalToolRegistration{}, err
	}
	if probe == nil {
		return LocalToolRegistration{}, errors.New("local tool version probe is required")
	}
	spec, err := canonicalLocalToolKind(kind)
	if err != nil {
		return LocalToolRegistration{}, err
	}
	path, err := requireAbsoluteRegularFile(executable, spec.Kind+" executable")
	if err != nil {
		return LocalToolRegistration{}, err
	}
	before, err := os.Stat(path)
	if err != nil {
		return LocalToolRegistration{}, err
	}
	if before.Size() <= 0 {
		return LocalToolRegistration{}, errors.New("local tool executable is empty")
	}
	if len([]rune(path)) > maxLocalToolPathRunes {
		return LocalToolRegistration{}, errors.New("local tool executable path is unbounded")
	}

	hashBefore, err := hashFile(path)
	if err != nil {
		return LocalToolRegistration{}, fmt.Errorf("fingerprint local tool executable: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return LocalToolRegistration{}, err
	}
	version, err := probe(ctx, spec.Kind, path)
	if err != nil {
		return LocalToolRegistration{}, err
	}
	version = strings.TrimSpace(version)
	if version == "" || len([]rune(version)) > maxLocalToolIdentityRunes {
		return LocalToolRegistration{}, errors.New("local tool version identity is missing or unbounded")
	}
	hashAfter, err := hashFile(path)
	if err != nil {
		return LocalToolRegistration{}, fmt.Errorf("re-fingerprint local tool executable: %w", err)
	}
	after, err := os.Stat(path)
	if err != nil {
		return LocalToolRegistration{}, err
	}
	if !sameStableFile(before, after) || before.Size() != after.Size() || hashBefore != hashAfter {
		return LocalToolRegistration{}, errors.New("local tool executable changed while it was being registered")
	}
	if !sha256TextPattern.MatchString(hashAfter) {
		return LocalToolRegistration{}, errors.New("local tool SHA-256 is invalid")
	}

	registration := LocalToolRegistration{
		Kind:         spec.Kind,
		Executable:   path,
		SHA256:       hashAfter,
		Size:         after.Size(),
		Version:      version,
		Upstream:     spec.Upstream,
		License:      spec.License,
		RegisteredAt: time.Now().UTC(),
	}
	if err := validateLocalToolRegistration(registration, spec); err != nil {
		return LocalToolRegistration{}, err
	}

	details := map[string]any{
		"kind":          registration.Kind,
		"executable":    registration.Executable,
		"sha256":        registration.SHA256,
		"size":          strconv.FormatInt(registration.Size, 10),
		"version":       registration.Version,
		"upstream":      registration.Upstream,
		"license":       registration.License,
		"registered_at": registration.RegisteredAt.Format(time.RFC3339Nano),
		"verified":      true,
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	oldChanges := append([]ChangeRecord(nil), v.Workspace.Changes...)
	oldUpdatedAt := v.Workspace.UpdatedAt
	oldBuildID := v.Workspace.BuildID
	v.addChangeUnlocked("user", localToolRegistrationChangeType, "Registered verified local "+registration.Kind+" runtime", details)
	registration.AuditChangeID = v.Workspace.Changes[0].ID
	if err := v.saveUnlocked(); err != nil {
		v.Workspace.Changes = oldChanges
		v.Workspace.UpdatedAt = oldUpdatedAt
		v.Workspace.BuildID = oldBuildID
		return LocalToolRegistration{}, fmt.Errorf("persist local tool registration: %w", err)
	}
	return registration, nil
}

func (v *Vault) RegisteredLocalTool(kind string) (LocalToolRegistration, error) {
	spec, err := canonicalLocalToolKind(kind)
	if err != nil {
		return LocalToolRegistration{}, err
	}
	ws := v.Snapshot()
	for _, change := range ws.Changes {
		if change.Type != localToolRegistrationChangeType || detailString(change.Details, "kind") != spec.Kind {
			continue
		}
		registration, err := localToolRegistrationFromChange(change)
		if err != nil {
			return LocalToolRegistration{}, fmt.Errorf("local tool registration %s is invalid: %w", change.ID, err)
		}
		if err := validateLocalToolRegistration(registration, spec); err != nil {
			return LocalToolRegistration{}, fmt.Errorf("local tool registration %s is inconsistent: %w", change.ID, err)
		}
		return registration, nil
	}
	return LocalToolRegistration{}, os.ErrNotExist
}

func (v *Vault) VerifyRegisteredLocalTool(ctx context.Context, kind string) (LocalToolRegistration, error) {
	return v.verifyRegisteredLocalToolWithProbe(ctx, kind, localToolVersion)
}

func (v *Vault) verifyRegisteredLocalToolWithProbe(ctx context.Context, kind string, probe localToolVersionProbe) (LocalToolRegistration, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if probe == nil {
		return LocalToolRegistration{}, errors.New("local tool version probe is required")
	}
	registration, err := v.RegisteredLocalTool(kind)
	if err != nil {
		return LocalToolRegistration{}, err
	}
	path, err := requireAbsoluteRegularFile(registration.Executable, registration.Kind+" registered executable")
	if err != nil {
		return LocalToolRegistration{}, fmt.Errorf("registered local tool is unavailable; re-register it: %w", err)
	}
	if path != registration.Executable {
		return LocalToolRegistration{}, errors.New("registered local tool resolved path changed; re-register it")
	}
	before, err := os.Stat(path)
	if err != nil {
		return LocalToolRegistration{}, err
	}
	if before.Size() != registration.Size {
		return LocalToolRegistration{}, errors.New("registered local tool size changed; re-register it")
	}
	hashBefore, err := hashFile(path)
	if err != nil {
		return LocalToolRegistration{}, err
	}
	if hashBefore != registration.SHA256 {
		return LocalToolRegistration{}, errors.New("registered local tool SHA-256 changed; re-register it")
	}
	version, err := probe(ctx, registration.Kind, path)
	if err != nil {
		return LocalToolRegistration{}, fmt.Errorf("registered local tool version check failed; re-register or repair it: %w", err)
	}
	if strings.TrimSpace(version) != registration.Version {
		return LocalToolRegistration{}, errors.New("registered local tool version changed; re-register it")
	}
	hashAfter, err := hashFile(path)
	if err != nil {
		return LocalToolRegistration{}, err
	}
	after, err := os.Stat(path)
	if err != nil {
		return LocalToolRegistration{}, err
	}
	if !sameStableFile(before, after) || hashAfter != registration.SHA256 || hashBefore != hashAfter {
		return LocalToolRegistration{}, errors.New("registered local tool changed during verification; re-register it")
	}
	return registration, nil
}

func localToolRegistrationFromChange(change ChangeRecord) (LocalToolRegistration, error) {
	if change.Type != localToolRegistrationChangeType {
		return LocalToolRegistration{}, errors.New("wrong change type")
	}
	if verified, ok := change.Details["verified"].(bool); !ok || !verified {
		return LocalToolRegistration{}, errors.New("verified marker is missing")
	}
	size, err := strconv.ParseInt(detailString(change.Details, "size"), 10, 64)
	if err != nil {
		return LocalToolRegistration{}, errors.New("registered executable size is missing or invalid")
	}
	registeredAt, err := time.Parse(time.RFC3339Nano, detailString(change.Details, "registered_at"))
	if err != nil {
		return LocalToolRegistration{}, errors.New("registration timestamp is missing or invalid")
	}
	return LocalToolRegistration{
		Kind:          detailString(change.Details, "kind"),
		Executable:    detailString(change.Details, "executable"),
		SHA256:        detailString(change.Details, "sha256"),
		Size:          size,
		Version:       detailString(change.Details, "version"),
		Upstream:      detailString(change.Details, "upstream"),
		License:       detailString(change.Details, "license"),
		RegisteredAt:  registeredAt.UTC(),
		AuditChangeID: change.ID,
	}, nil
}

func validateLocalToolRegistration(registration LocalToolRegistration, spec localToolSpec) error {
	if registration.Kind != spec.Kind || registration.Upstream != spec.Upstream || registration.License != spec.License {
		return errors.New("registered local tool identity does not match ECO's approved donor specification")
	}
	if registration.Executable == "" || len([]rune(registration.Executable)) > maxLocalToolPathRunes {
		return errors.New("registered local tool path is missing or unbounded")
	}
	if !sha256TextPattern.MatchString(registration.SHA256) || registration.Size <= 0 {
		return errors.New("registered local tool fingerprint is invalid")
	}
	if registration.Version == "" || len([]rune(registration.Version)) > maxLocalToolIdentityRunes {
		return errors.New("registered local tool version is missing or unbounded")
	}
	if registration.RegisteredAt.IsZero() {
		return errors.New("registered local tool timestamp is missing")
	}
	return nil
}

// The registered wrappers remove raw executable paths from ordinary ECO calls.
// Every wrapper verifies the approved executable immediately before passing it
// into the existing source-safe adapter/workflow.
func (v *Vault) OCRImageWithRegisteredTesseract(evidenceID, language string) error {
	return v.OCRImageWithRegisteredTesseractContext(context.Background(), evidenceID, language)
}

func (v *Vault) OCRImageWithRegisteredTesseractContext(ctx context.Context, evidenceID, language string) error {
	tool, err := v.VerifyRegisteredLocalTool(ctx, "tesseract")
	if err != nil {
		return err
	}
	return v.OCRImageWithTesseractContext(ctx, evidenceID, tool.Executable, language)
}

func (v *Vault) ExtractEvidenceWithRegisteredDocling(evidenceID, artifactsPath string) error {
	return v.ExtractEvidenceWithRegisteredDoclingContext(context.Background(), evidenceID, artifactsPath)
}

func (v *Vault) ExtractEvidenceWithRegisteredDoclingContext(ctx context.Context, evidenceID, artifactsPath string) error {
	tool, err := v.VerifyRegisteredLocalTool(ctx, "docling")
	if err != nil {
		return err
	}
	return v.ExtractEvidenceWithDoclingContext(ctx, evidenceID, tool.Executable, artifactsPath)
}

func (v *Vault) ExtractEvidenceWithRegisteredOCRmyPDF(evidenceID, language string) (OCRmyPDFResult, error) {
	return v.ExtractEvidenceWithRegisteredOCRmyPDFContext(context.Background(), evidenceID, language)
}

func (v *Vault) ExtractEvidenceWithRegisteredOCRmyPDFContext(ctx context.Context, evidenceID, language string) (OCRmyPDFResult, error) {
	tool, err := v.VerifyRegisteredLocalTool(ctx, "ocrmypdf")
	if err != nil {
		return OCRmyPDFResult{}, err
	}
	return v.ExtractEvidenceWithOCRmyPDFContext(ctx, evidenceID, tool.Executable, language)
}

func (v *Vault) InspectEvidencePDFWithRegisteredPDFCPU(evidenceID string) (PDFAssessment, error) {
	return v.InspectEvidencePDFWithRegisteredPDFCPUContext(context.Background(), evidenceID)
}

func (v *Vault) InspectEvidencePDFWithRegisteredPDFCPUContext(ctx context.Context, evidenceID string) (PDFAssessment, error) {
	tool, err := v.VerifyRegisteredLocalTool(ctx, "pdfcpu")
	if err != nil {
		return PDFAssessment{}, err
	}
	return v.InspectEvidencePDFWithPDFCPUContext(ctx, evidenceID, tool.Executable)
}

func (v *Vault) AskWithRegisteredLlamaCPP(question string, scopeIDs []string, modelPath string) (LlamaCPPAnswerResult, error) {
	return v.AskWithRegisteredLlamaCPPContext(context.Background(), question, scopeIDs, modelPath)
}

func (v *Vault) AskWithRegisteredLlamaCPPContext(ctx context.Context, question string, scopeIDs []string, modelPath string) (LlamaCPPAnswerResult, error) {
	tool, err := v.VerifyRegisteredLocalTool(ctx, "llama.cpp")
	if err != nil {
		return LlamaCPPAnswerResult{}, err
	}
	return v.AskWithLlamaCPPContext(ctx, question, scopeIDs, tool.Executable, modelPath)
}
