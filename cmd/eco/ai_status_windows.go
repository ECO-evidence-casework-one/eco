//go:build windows

package main

import (
	"strings"
	"time"

	"github.com/ECO-evidence-casework-one/eco/internal/eco"
)

// The current native shell already exposes the Ask answer as a standard
// read-only Windows EDIT control. Until the larger accessibility shell work
// replaces custom-drawn status surfaces, use that existing accessible control
// to make model/fallback provenance explicit instead of adding another painted
// badge that Narrator/NVDA may not discover.
func init() {
	go monitorLocalAIStatus()
}

func monitorLocalAIStatus() {
	var a *application
	for i := 0; i < 600; i++ {
		if app != nil && app.vault != nil && app.answerEdit != 0 {
			a = app
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if a == nil {
		return
	}

	setWindowText(a.answerEdit, "Checking local Qwen configuration…")
	status := a.vault.LocalAIStatus()
	ready := status.Ready
	setWindowText(a.answerEdit, initialAIStatusText(status))

	lastDecoratedReceipt := ""
	runningShown := false
	ticker := time.NewTicker(120 * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		if app == nil || app != a || a.hwnd == 0 || a.answerEdit == 0 {
			return
		}

		text := strings.TrimSpace(getWindowText(a.answerEdit))
		if strings.HasPrefix(text, "Searching local source passages") {
			if ready {
				setWindowText(a.answerEdit, "QWEN RUNNING — reasoning locally from ECO's verified source context…")
			} else {
				setWindowText(a.answerEdit, "SOURCE-ONLY SEARCH — Qwen is not ready; ECO is searching verified local source passages…")
			}
			runningShown = true
		}

		a.mu.Lock()
		rec := a.lastQuestion
		a.mu.Unlock()
		if rec.ReceiptID == "" || rec.ReceiptID == lastDecoratedReceipt || strings.TrimSpace(rec.Answer) == "" {
			continue
		}

		// msgAskDone writes rec.Answer into the edit control. Wait for that exact
		// transition so this monitor cannot race it and lose the provenance line.
		if strings.TrimSpace(getWindowText(a.answerEdit)) != strings.TrimSpace(rec.Answer) {
			continue
		}

		prefix := questionAIStatusText(a, status, rec)
		setWindowText(a.answerEdit, prefix+"\r\n\r\n"+rec.Answer)
		lastDecoratedReceipt = rec.ReceiptID
		runningShown = false
		_ = runningShown
	}
}

func initialAIStatusText(status eco.LocalAIStatus) string {
	switch status.State {
	case "ready":
		model := status.ModelName
		if model == "" {
			model = "the configured Qwen model"
		}
		return "QWEN READY — " + model + " and the local llama.cpp runtime passed ECO's hash and offline readiness checks.\r\n\r\nAsk a question about readable evidence. ECO will still independently verify model-selected claims before releasing source wording."
	case "invalid":
		return "QWEN CONFIGURATION NEEDS ATTENTION — " + status.Detail + "\r\n\r\nAsk ECO still works: it will use the deterministic source-backed engine rather than pretending the model ran."
	default:
		return "QWEN UNAVAILABLE — no verified local model is active for this preview.\r\n\r\nAsk ECO still works from deterministic, verified local source passages."
	}
}

func questionAIStatusText(a *application, startup eco.LocalAIStatus, rec eco.QuestionRecord) string {
	if eco.QuestionUsedLocalAI(rec) {
		return "QWEN CHECKED — the local model ran. ECO released only claims that passed deterministic source grounding."
	}
	if !startup.Ready {
		return "SOURCE-ONLY ANSWER — Qwen was not ready, so this answer came from ECO's deterministic source-backed engine."
	}

	// A ready model can still be blocked, rejected or fail its grounding contract.
	// The workspace audit records those reasons. Keep the ordinary UI concise but
	// distinguish a rejected model output from a generic fallback where possible.
	ws := a.vault.Snapshot()
	for i, change := range ws.Changes {
		if i >= 12 {
			break
		}
		switch change.Type {
		case "local-ai-grounding-rejected":
			return "QWEN REJECTED — the model ran, but ECO did not accept its grounded output. SOURCE FALLBACK USED."
		case "local-ai-resource-blocked":
			return "QWEN UNAVAILABLE FOR THIS QUESTION — ECO blocked model launch because local resources were critically constrained. SOURCE FALLBACK USED."
		case "configured-local-ai-fallback":
			return "SOURCE FALLBACK USED — Qwen was configured, but it did not produce an accepted grounded answer."
		case "grounded-local-ai-question":
			if id, ok := change.Details["question_id"].(string); ok && id == rec.ID {
				return "QWEN CHECKED — the local model ran. ECO released only claims that passed deterministic source grounding."
			}
		}
	}
	return "SOURCE FALLBACK USED — Qwen was ready at startup, but this answer was released by ECO's deterministic source-backed engine."
}
