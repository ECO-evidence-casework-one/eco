# ECO Product Lab — visible local-AI truth state

**Date:** 7 September 2026  
**Scope:** first user-visible AI status slice after PR #139

## Problem

The current Ask screen still contains deterministic-only wording even though the application-facing Ask path can now use a verified local llama.cpp/Qwen configuration. A user must not have to infer from answer style whether the model actually ran.

## This slice

This slice adds a non-mutating local-AI readiness check and exposes the result through ECO's existing native read-only Ask answer control.

States shown to the user:

- **QWEN READY** — runtime/model identities and offline llama.cpp version check passed.
- **QWEN RUNNING** — the current Ask operation has entered the configured local-AI path.
- **QWEN CHECKED** — the persisted question was released through ECO's grounded llama.cpp workflow.
- **SOURCE-ONLY ANSWER** — no usable Qwen configuration was active; deterministic source retrieval answered.
- **SOURCE FALLBACK USED** — Qwen was configured/ready but did not produce an accepted grounded answer.
- **QWEN REJECTED** — a model emission was rejected by ECO grounding before release, then source fallback answered.
- **QWEN UNAVAILABLE FOR THIS QUESTION** — ECO blocked model launch because local resources were critically constrained, then source fallback answered.
- **QWEN CONFIGURATION NEEDS ATTENTION** — configured files failed identity/readiness verification.

## Accessibility choice

The status is deliberately placed in the existing standard Windows read-only `EDIT` control rather than a new custom-painted badge. This gives Windows assistive technology an ordinary text-control surface while the broader UI Automation/nav work proceeds.

This is an interim accessibility improvement, not acceptance of the complete Ask screen. The custom-drawn subtitle/receipt and citation cards still require the wider Issue #7 accessibility work.

## Truth boundary

- `ready` means the configured files passed their hash checks and llama.cpp passed an offline version probe. It does **not** mean a question used the model.
- `QWEN CHECKED` is shown only when the persisted QuestionRecord carries ECO's grounded llama.cpp support classification or its exact question audit record identifies the successful model path.
- fallback/rejection messages are bound to the current question so an older model failure cannot be misreported as the status of a later answer.
- the underlying persisted answer text is not rewritten; the status line is a UI presentation layer.

## Remaining work

- remove the stale deterministic-only custom-drawn Ask subtitle and receipt wording;
- make primary navigation and custom-drawn action/citation/evidence regions properly keyboard/UI Automation accessible;
- move routine AI progress/errors toward non-modal in-window status surfaces;
- perform Narrator/NVDA/Accessibility Insights acceptance;
- prove a real Qwen generation and Ask ECO turn on the user's rig.
