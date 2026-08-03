package main

import (
	"os"
	"strings"
	"testing"
)

func windowsSource(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("main_windows.go")
	if err != nil {
		t.Fatalf("read native Windows source: %v", err)
	}
	return string(b)
}

func functionBody(t *testing.T, src, signature, next string) string {
	t.Helper()
	start := strings.Index(src, signature)
	if start < 0 {
		t.Fatalf("missing %s", signature)
	}
	end := strings.Index(src[start+len(signature):], next)
	if end < 0 {
		t.Fatalf("missing end marker %s", next)
	}
	return src[start : start+len(signature)+end]
}

func TestPaintDoesNotCloneOrLockVault(t *testing.T) {
	src := windowsSource(t)
	body := functionBody(t, src, "func (a *application) paint", "func (a *application) refreshView")
	if strings.Contains(body, "Snapshot()") || strings.Contains(body, "a.vault") {
		t.Fatalf("paint path must not clone or lock the encrypted workspace")
	}
}

func TestBackgroundResultsReturnThroughWindowMessages(t *testing.T) {
	src := windowsSource(t)
	if !strings.Contains(src, "msgAskDone") || !strings.Contains(src, "msgVerifyDone") || !strings.Contains(src, "msgPreviewReady") {
		t.Fatalf("expected message-queue completion routes for background work")
	}
	ask := functionBody(t, src, "func (a *application) askEvidence", "func (a *application) openCitation")
	if strings.Count(ask, "setWindowText") != 1 {
		t.Fatalf("Ask ECO worker must not directly modify the Windows edit control")
	}
	verify := functionBody(t, src, "func (a *application) verifyEvidence", "func (a *application) createBackup")
	if strings.Contains(verify, "messageBox") {
		t.Fatalf("integrity worker must not open a Windows dialog from its background goroutine")
	}
}

func TestDPIAndLiteralTextCorrectionsPresent(t *testing.T) {
	src := windowsSource(t)
	for _, required := range []string{"WM_DPICHANGED", "DT_NOPREFIX", "fontDPI > 108", "Trust & settings"} {
		if !strings.Contains(src, required) {
			t.Fatalf("missing native layout safeguard %q", required)
		}
	}
}

func TestClipboardShortcutDoesNotStealNormalPaste(t *testing.T) {
	src := windowsSource(t)
	body := functionBody(t, src, "func (a *application) handleGlobalShortcut", "func (a *application) handleKey")
	for _, required := range []string{"shiftDown", "procGetFocus", "focus != a.questionEdit", "Ctrl+Shift+V"} {
		if !strings.Contains(body, required) {
			t.Fatalf("clipboard shortcut safeguard missing %q", required)
		}
	}
}

func TestDocumentVisionPreviewControlsPresent(t *testing.T) {
	src := windowsSource(t)
	for _, required := range []string{"BoundedPreviewImage", "SuggestDocumentBounds", "SkewCorrectionDegrees", "case 'C'", "case 'D'", "case 'A'", "case 'Q'", "adaptive reading mode"} {
		if !strings.Contains(src, required) {
			t.Fatalf("missing document-vision preview safeguard %q", required)
		}
	}
}

func TestCoordinateCitationHighlightingPresent(t *testing.T) {
	src := windowsSource(t)
	for _, required := range []string{"pendingCitationRegion", "exact OCR region highlighted", "rotateNormalizedRegion", "drawHighlightRect"} {
		if !strings.Contains(src, required) {
			t.Fatalf("missing coordinate citation feature %q", required)
		}
	}
}

func TestWorkspaceIdentityAndExplicitLifecycleControlsPresent(t *testing.T) {
	src := windowsSource(t)
	for _, required := range []string{"StartCandidate", "Identity:", "Path:", "Opened as:", "Compatibility:", "Create new workspace", "Open / migrate", "Workspace recovered", "Reset selected", "Reset only this selected workspace"} {
		if !strings.Contains(src, required) {
			t.Fatalf("missing explicit workspace lifecycle wording or control %q", required)
		}
	}
	startup := functionBody(t, src, "func main()", "func registerClasses")
	if strings.Contains(startup, "OpenVault(") || strings.Contains(startup, `"V25N2"`) {
		t.Fatal("startup must use candidate-specific application state instead of silently reopening the previous fixed vault")
	}
}

func TestNativeSourceContainsNoBrowserOrLocalhostRuntime(t *testing.T) {
	src := windowsSource(t)
	for _, forbidden := range []string{"\"net/http\"", "ListenAndServe", "brave.exe", "msedge.exe", "chrome.exe"} {
		if strings.Contains(strings.ToLower(src), strings.ToLower(forbidden)) {
			t.Fatalf("native Windows source contains forbidden browser/server runtime %q", forbidden)
		}
	}
}
