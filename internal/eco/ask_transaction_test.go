package eco

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func disableConfiguredLocalAI(t *testing.T) {
	t.Helper()
	for _, name := range []string{envLlamaCPPExecutable, envLlamaCPPExecutableSHA256, envLlamaCPPModel, envLlamaCPPModelSHA256} {
		t.Setenv(name, "")
	}
}

func importSyntheticText(t *testing.T, v *Vault, name, text string) EvidenceItem {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	item, _, err := v.ImportFile(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func TestWorkspaceOnlyAskPerformsNoSourceVerification(t *testing.T) {
	disableConfiguredLocalAI(t)
	t.Setenv(envLlamaCPPExecutable, filepath.Join(t.TempDir(), "incomplete-llama-cli.exe"))
	v, err := openTestVault(filepath.Join(t.TempDir(), "vault"))
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	hash := strings.Repeat("a", 64)
	v.mu.Lock()
	v.Workspace.Evidence = []EvidenceItem{
		{ID: "E-READABLE", ObjectFile: "E-READABLE.ecoobj", SHA256: hash, Preservation: preservationCommitted, SourceVerified: true, Readable: true},
		{ID: "E-IMAGE", ObjectFile: "E-IMAGE.ecoobj", SHA256: hash, Preservation: preservationCommitted, SourceVerified: true, Image: &ImageAssessment{}, Warnings: []string{"synthetic warning"}},
		{ID: "E-UNVERIFIED", ObjectFile: "E-UNVERIFIED.ecoobj", SHA256: hash, Preservation: preservationCommitted},
		{ID: "E-QUARANTINED", ObjectFile: "E-QUARANTINED.ecoobj", SHA256: hash, Preservation: preservationCommitted, SourceVerified: true, Status: "Quarantined"},
	}
	v.Workspace.Matters = []Matter{{ID: "M-ACTIVE", Status: "ACTIVE"}, {ID: "M-CLOSED", Status: "Closed"}}
	if err := v.saveUnlocked(); err != nil {
		v.mu.Unlock()
		t.Fatal(err)
	}
	v.mu.Unlock()

	verified := 0
	snapshots := 0
	v.sourceVerificationBoundary = func(string) { verified++ }
	v.snapshotBoundary = func() { snapshots++ }
	questions := []struct {
		text   string
		intent string
	}{
		{"What is the current workspace status?", "status"},
		{"How do I use ECO and what can you do?", "help"},
		{"What is the evidence integrity?", "integrity"},
	}
	for i, question := range questions {
		v.mu.Lock()
		wantRevision := v.Workspace.Revision
		v.mu.Unlock()
		record := v.Ask(question.text, nil)
		if record.Intent != question.intent {
			t.Fatalf("%q classified as %q, want %q", question.text, record.Intent, question.intent)
		}
		if record.EvidenceConsidered != 0 || record.SourceVerificationBytes != 0 || len(record.VerifiedEvidenceIDs) != 0 {
			t.Fatalf("workspace-only Ask reported source verification: %+v", record)
		}
		if record.WorkspaceRevision != wantRevision {
			t.Fatalf("question %d recorded workspace revision %d, want %d", i, record.WorkspaceRevision, wantRevision)
		}
		if question.intent == "status" && !strings.Contains(record.Answer, "4 preserved evidence items, 1 readable items, 1 assessed images, 1 active matters and 3 items needing attention") {
			t.Fatalf("status answer did not use the lightweight summary correctly: %q", record.Answer)
		}
	}
	if verified != 0 {
		t.Fatalf("workspace-only Ask verified %d preserved objects, want zero", verified)
	}
	if snapshots != 0 {
		t.Fatalf("workspace-only Ask took %d full workspace snapshots, want zero", snapshots)
	}
	v.snapshotBoundary = nil
	ws := v.Snapshot()
	if got := len(ws.Questions); got != len(questions) {
		t.Fatalf("persisted %d workspace-only questions, want %d", got, len(questions))
	}
	audits := 0
	for _, change := range ws.Changes {
		if change.Type == "question-asked" {
			audits++
		}
		if change.Type == "configured-local-ai-fallback" {
			t.Fatal("workspace-only Ask unexpectedly entered the configured local-AI path")
		}
	}
	if audits != len(questions) {
		t.Fatalf("persisted %d question audits, want %d", audits, len(questions))
	}
}

func TestWorkspaceAskSummaryCannotCarryNestedPayloads(t *testing.T) {
	typeOfSummary := reflect.TypeOf(workspaceAskSummary{})
	for i := 0; i < typeOfSummary.NumField(); i++ {
		kind := typeOfSummary.Field(i).Type.Kind()
		if kind != reflect.Int && kind != reflect.Uint64 {
			t.Fatalf("workspace summary field %q can retain non-scalar payload type %s", typeOfSummary.Field(i).Name, kind)
		}
	}
	hash := strings.Repeat("b", 64)
	v := &Vault{Workspace: newWorkspace()}
	v.Workspace.Revision = 42
	v.Workspace.Evidence = []EvidenceItem{{
		ID:             "E-LARGE-NESTED",
		ObjectFile:     "E-LARGE-NESTED.ecoobj",
		SHA256:         hash,
		Preservation:   preservationCommitted,
		SourceVerified: true,
		Readable:       true,
		Segments:       make([]SourceSegment, 100000),
		OCR:            &OCRReceipt{Words: make([]OCRWord, 100000), Lines: make([]OCRLine, 10000)},
	}}
	snapshots := 0
	v.snapshotBoundary = func() { snapshots++ }
	summary := v.askWorkspaceSummary("status")
	if summary.revision != 42 || summary.evidence != 1 || summary.readable != 1 {
		t.Fatalf("unexpected scalar summary: %+v", summary)
	}
	if snapshots != 0 {
		t.Fatalf("lightweight summary invoked full Snapshot %d times", snapshots)
	}
}

func TestAskVerifiesOnlyBoundedEligibleSources(t *testing.T) {
	disableConfiguredLocalAI(t)
	v, err := openTestVault(filepath.Join(t.TempDir(), "vault"))
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	const total = maxAskVerificationItems + 8
	for i := 0; i < total; i++ {
		importSyntheticText(t, v, fmt.Sprintf("hearing-%02d.txt", i), fmt.Sprintf("Synthetic omega hearing source %d records an ordinary bounded verification fact.", i))
	}

	verified := make([]string, 0, maxAskVerificationItems)
	v.sourceVerificationBoundary = func(id string) { verified = append(verified, id) }
	record := v.Ask("What does the synthetic omega hearing evidence record?", nil)
	if len(verified) != maxAskVerificationItems {
		t.Fatalf("verified %d objects, want bounded maximum %d (all evidence=%d)", len(verified), maxAskVerificationItems, total)
	}
	if record.EvidenceConsidered != maxAskVerificationItems || len(record.VerifiedEvidenceIDs) != maxAskVerificationItems || !record.SourceVerificationLimit {
		t.Fatalf("bounded verification receipt is incomplete: %+v", record)
	}
	if len(record.Citations) == 0 {
		t.Fatalf("expected a source-backed answer from verified candidates: %+v", record)
	}
	verifiedSet := make(map[string]bool, len(verified))
	for _, id := range verified {
		verifiedSet[id] = true
	}
	for _, citation := range record.Citations {
		if !verifiedSet[citation.EvidenceID] {
			t.Fatalf("citation used an unverified source: %+v; verified=%v", citation, verified)
		}
	}
}

func TestWorkspaceWordDoesNotBypassEvidenceVerification(t *testing.T) {
	disableConfiguredLocalAI(t)
	v, err := openTestVault(filepath.Join(t.TempDir(), "vault"))
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	item := importSyntheticText(t, v, "comparison.txt", "Synthetic workspace evidence records the alpha and beta comparison.")
	verified := ""
	v.sourceVerificationBoundary = func(id string) { verified = id }
	record := v.Ask("Compare the evidence in this workspace", nil)
	if record.Intent != "compare" || verified != item.ID || len(record.Citations) == 0 {
		t.Fatalf("evidence question bypassed source verification: verified=%q record=%+v", verified, record)
	}
}

func TestAskExcludesCorruptCandidateAndUsesVerifiedFallback(t *testing.T) {
	disableConfiguredLocalAI(t)
	v, err := openTestVault(filepath.Join(t.TempDir(), "vault"))
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	corrupt := importSyntheticText(t, v, "corrupt.txt", strings.Repeat("omega hearing ", 20)+"is on 10 October 2026.")
	valid := importSyntheticText(t, v, "valid.txt", "The synthetic omega hearing is on 22 November 2026 and this fallback is valid.")
	objectPath := filepath.Join(v.Objects, corrupt.ObjectFile)
	if err := os.Chmod(objectPath, 0600); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(objectPath)
	if err != nil {
		t.Fatal(err)
	}
	raw[len(raw)-1] ^= 0x5a
	if err := os.WriteFile(objectPath, raw, 0600); err != nil {
		t.Fatal(err)
	}

	record := v.Ask("When is the synthetic omega hearing?", nil)
	if record.SourceVerificationFailures != 1 || !strings.Contains(record.Answer, "22 November 2026") {
		t.Fatalf("corrupt top candidate did not fail closed with verified fallback: %+v", record)
	}
	for _, citation := range record.Citations {
		if citation.EvidenceID == corrupt.ID {
			t.Fatalf("corrupt evidence supported the answer: %+v", citation)
		}
		if citation.EvidenceID != valid.ID {
			t.Fatalf("unexpected fallback citation: %+v", citation)
		}
	}
	state := v.Snapshot()
	for _, item := range state.Evidence {
		if item.ID == corrupt.ID && (item.SourceVerified || !strings.Contains(item.Status, "blocked")) {
			t.Fatalf("corrupt candidate was not blocked: %+v", item)
		}
	}
}

type askRestoreFixture struct {
	active     *Vault
	backupPath string
}

func newAskRestoreFixture(t *testing.T) askRestoreFixture {
	t.Helper()
	dir := t.TempDir()
	active, err := openTestVault(filepath.Join(dir, "active"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = active.Close() })
	importSyntheticText(t, active, "active.txt", "The synthetic transaction marker says ACTIVE-OLD-WORKSPACE.")

	source, err := openTestVault(filepath.Join(dir, "source"))
	if err != nil {
		t.Fatal(err)
	}
	importSyntheticText(t, source, "restored.txt", "The synthetic transaction marker says RESTORED-NEW-WORKSPACE.")
	backupPath := filepath.Join(dir, "source.ecobackup")
	if _, err := source.CreatePortableBackup(backupPath, "synthetic restore passphrase", nil); err != nil {
		_ = source.Close()
		t.Fatal(err)
	}
	if err := source.Close(); err != nil {
		t.Fatal(err)
	}
	return askRestoreFixture{active: active, backupPath: backupPath}
}

func TestAskCompletesBeforeRestoreActivation(t *testing.T) {
	disableConfiguredLocalAI(t)
	fixture := newAskRestoreFixture(t)
	verificationStarted := make(chan struct{})
	releaseVerification := make(chan struct{})
	var once sync.Once
	fixture.active.sourceVerificationBoundary = func(string) {
		once.Do(func() { close(verificationStarted) })
		<-releaseVerification
	}
	records := make(chan QuestionRecord, 1)
	go func() { records <- fixture.active.Ask("What does the synthetic transaction marker say?", nil) }()
	waitForSignal(t, verificationStarted, "Ask source verification")

	phases := make(chan string, 8)
	fixture.active.restoreBoundary = func(phase string) { phases <- phase }
	restoreDone := make(chan error, 1)
	go func() {
		_, err := fixture.active.RestorePortableBackup(fixture.backupPath, "synthetic restore passphrase", nil)
		restoreDone <- err
	}()
	waitForPhase(t, phases, "stage_verified")
	assertNoPhase(t, phases, "original_renamed")
	close(releaseVerification)
	record := waitForRecord(t, records)
	if !strings.Contains(record.Answer, "ACTIVE-OLD-WORKSPACE") {
		t.Fatalf("Ask did not remain wholly on the old workspace: %+v", record)
	}
	if err := waitForError(t, restoreDone); err != nil {
		t.Fatal(err)
	}
	ws := fixture.active.Snapshot()
	if len(ws.Evidence) != 1 || ws.Evidence[0].SafeName != "restored.txt" || len(ws.Questions) != 0 {
		t.Fatalf("restore activation mixed the completed old Ask into new state: evidence=%+v questions=%+v", ws.Evidence, ws.Questions)
	}
}

func TestAskWaitsForRestoreActivation(t *testing.T) {
	disableConfiguredLocalAI(t)
	fixture := newAskRestoreFixture(t)
	activationStarted := make(chan struct{})
	releaseActivation := make(chan struct{})
	var once sync.Once
	fixture.active.restoreBoundary = func(phase string) {
		if phase == "original_renamed" {
			once.Do(func() { close(activationStarted) })
			<-releaseActivation
		}
	}
	restoreDone := make(chan error, 1)
	go func() {
		_, err := fixture.active.RestorePortableBackup(fixture.backupPath, "synthetic restore passphrase", nil)
		restoreDone <- err
	}()
	waitForSignal(t, activationStarted, "restore activation")

	verificationStarted := make(chan struct{}, 1)
	fixture.active.sourceVerificationBoundary = func(string) { verificationStarted <- struct{}{} }
	records := make(chan QuestionRecord, 1)
	askStarted := make(chan struct{})
	go func() {
		close(askStarted)
		records <- fixture.active.Ask("What does the synthetic transaction marker say?", nil)
	}()
	waitForSignal(t, askStarted, "Ask goroutine")
	select {
	case <-verificationStarted:
		t.Fatal("Ask began source verification while restore activation held the operation lock")
	case <-records:
		t.Fatal("Ask completed while restore activation held the operation lock")
	case <-time.After(100 * time.Millisecond):
	}
	close(releaseActivation)
	if err := waitForError(t, restoreDone); err != nil {
		t.Fatal(err)
	}
	record := waitForRecord(t, records)
	if !strings.Contains(record.Answer, "RESTORED-NEW-WORKSPACE") || strings.Contains(record.Answer, "ACTIVE-OLD-WORKSPACE") {
		t.Fatalf("Ask did not bind wholly to restored state: %+v", record)
	}
	ws := fixture.active.Snapshot()
	if len(ws.Questions) != 1 || ws.Questions[0].ID != record.ID || len(record.Citations) == 0 || record.Citations[0].EvidenceID != ws.Evidence[0].ID {
		t.Fatalf("restored Ask record is not coherently bound to the active workspace: record=%+v workspace=%+v", record, ws)
	}
}

func TestGroundedLocalAICompletesBeforeRestoreActivation(t *testing.T) {
	disableConfiguredLocalAI(t)
	fixture := newAskRestoreFixture(t)
	modelStarted := make(chan struct{})
	releaseModel := make(chan struct{})
	fake := func(_ context.Context, _, _ string, grounding GroundingContext) (LlamaCPPModelResult, error) {
		close(modelStarted)
		<-releaseModel
		record := grounding.Records[0]
		return LlamaCPPModelResult{
			Emission: GroundingEmission{
				Answer: "Synthetic draft",
				Claims: []GroundingClaim{{Kind: "presence", EvidenceID: record.EvidenceID, SegmentID: record.SegmentID}},
			},
			EngineVersion: "synthetic llama.cpp",
			ModelName:     "synthetic-qwen.gguf",
			ModelSHA256:   strings.Repeat("b", 64),
		}, nil
	}
	results := make(chan LlamaCPPAnswerResult, 1)
	modelErrors := make(chan error, 1)
	go func() {
		result, err := fixture.active.askWithLlamaCPPRunner(context.Background(), "What does the synthetic transaction marker say?", nil, "ignored", "ignored", fake)
		results <- result
		modelErrors <- err
	}()
	waitForSignal(t, modelStarted, "synthetic local model")

	phases := make(chan string, 8)
	fixture.active.restoreBoundary = func(phase string) { phases <- phase }
	restoreDone := make(chan error, 1)
	go func() {
		_, err := fixture.active.RestorePortableBackup(fixture.backupPath, "synthetic restore passphrase", nil)
		restoreDone <- err
	}()
	waitForPhase(t, phases, "stage_verified")
	assertNoPhase(t, phases, "original_renamed")
	close(releaseModel)
	if err := waitForError(t, modelErrors); err != nil {
		t.Fatal(err)
	}
	result := waitForLlamaResult(t, results)
	if !QuestionUsedLocalAI(result.Question) || !strings.Contains(result.Question.Answer, "ACTIVE-OLD-WORKSPACE") {
		t.Fatalf("local-model answer did not remain wholly on old verified state: %+v", result)
	}
	if result.Question.SourceVerificationBytes > maxAskVerificationBytes {
		t.Fatalf("two-pass local-model verification exceeded byte budget: %+v", result.Question)
	}
	if err := waitForError(t, restoreDone); err != nil {
		t.Fatal(err)
	}
	if ws := fixture.active.Snapshot(); len(ws.Questions) != 0 || ws.Evidence[0].SafeName != "restored.txt" {
		t.Fatalf("restore mixed old local-model output into restored workspace: %+v", ws)
	}
}

func waitForSignal(t *testing.T, signal <-chan struct{}, name string) {
	t.Helper()
	select {
	case <-signal:
	case <-time.After(30 * time.Second):
		t.Fatalf("timed out waiting for %s", name)
	}
}

func waitForPhase(t *testing.T, phases <-chan string, want string) {
	t.Helper()
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	for {
		select {
		case phase := <-phases:
			if phase == want {
				return
			}
		case <-timer.C:
			t.Fatalf("timed out waiting for restore phase %q", want)
		}
	}
}

func assertNoPhase(t *testing.T, phases <-chan string, forbidden string) {
	t.Helper()
	select {
	case phase := <-phases:
		if phase == forbidden {
			t.Fatalf("restore crossed %q while Ask transaction was active", forbidden)
		}
	case <-time.After(100 * time.Millisecond):
	}
}

func waitForRecord(t *testing.T, records <-chan QuestionRecord) QuestionRecord {
	t.Helper()
	select {
	case record := <-records:
		return record
	case <-time.After(30 * time.Second):
		t.Fatal("timed out waiting for Ask")
		return QuestionRecord{}
	}
}

func waitForLlamaResult(t *testing.T, results <-chan LlamaCPPAnswerResult) LlamaCPPAnswerResult {
	t.Helper()
	select {
	case result := <-results:
		return result
	case <-time.After(30 * time.Second):
		t.Fatal("timed out waiting for local-model Ask")
		return LlamaCPPAnswerResult{}
	}
}

func waitForError(t *testing.T, result <-chan error) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(30 * time.Second):
		t.Fatal("timed out waiting for restore")
		return nil
	}
}

func TestAskVerificationByteLimitExcludesOversizedCandidate(t *testing.T) {
	disableConfiguredLocalAI(t)
	item := EvidenceItem{
		ID:             "E-SYNTHETIC-LARGE",
		SafeName:       "large.txt",
		ObjectFile:     "E-SYNTHETIC-LARGE.ecoobj",
		SHA256:         strings.Repeat("a", 64),
		Size:           maxAskVerificationBytes + 1,
		Preservation:   preservationCommitted,
		SourceVerified: true,
		Segments: []SourceSegment{{
			ID:           "SEG-SYNTHETIC-LARGE",
			Text:         "Synthetic oversized omega evidence candidate.",
			SourceObject: "E-SYNTHETIC-LARGE.ecoobj",
			SourceSHA256: strings.Repeat("a", 64),
		}},
	}
	v, err := openTestVault(filepath.Join(t.TempDir(), "vault"))
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	v.mu.Lock()
	v.Workspace.Evidence = []EvidenceItem{item}
	if err := v.saveUnlocked(); err != nil {
		v.mu.Unlock()
		t.Fatal(err)
	}
	v.mu.Unlock()
	verified := false
	v.sourceVerificationBoundary = func(string) { verified = true }
	record := v.Ask("What does the oversized omega evidence say?", nil)
	if verified || !record.SourceVerificationLimit || record.EvidenceConsidered != 0 || len(record.Citations) != 0 || record.SourceVerificationBytes != 0 {
		t.Fatalf("oversized source crossed the Ask byte budget: verified=%v record=%+v", verified, record)
	}
	if !bytes.Contains([]byte(record.Answer), []byte("64 MiB")) {
		t.Fatalf("bounded outcome was not explained: %q", record.Answer)
	}
}

func TestAskPersistenceFailureRollsBackQuestionAndAudit(t *testing.T) {
	disableConfiguredLocalAI(t)
	v, err := openTestVault(filepath.Join(t.TempDir(), "vault"))
	if err != nil {
		t.Fatal(err)
	}
	defer v.Close()
	before := v.Snapshot()
	if err := os.Mkdir(filepath.Join(v.Root, "workspace.ecodb.tmp"), 0700); err != nil {
		t.Fatal(err)
	}
	record := v.Ask("What is the current workspace status?", nil)
	if record.ID != "" || record.ReceiptID != "" || len(record.Citations) != 0 || record.Support != "Question was not committed" {
		t.Fatalf("failed persistence returned a committed-looking answer: %+v", record)
	}
	after := v.Snapshot()
	if len(after.Questions) != len(before.Questions) || len(after.Changes) != len(before.Changes) || after.Revision != before.Revision {
		t.Fatalf("failed Ask retained in-memory transaction state: before=%+v after=%+v", before, after)
	}
}
