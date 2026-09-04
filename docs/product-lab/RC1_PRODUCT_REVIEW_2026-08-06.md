# ECO RC1 — Product Lab Review

**Candidate:** `ECO-RECOVERY-20260805-RC1`  
**Date:** 6 August 2026  
**Decision:** RETURN TO DEVELOPMENT — genuine recovery achieved, product behaviour incomplete

## What RC1 proved

- The recovered application is recognisably V38-derived.
- The rejected V40 replacement shell did not return.
- Qwen2.5 1.5B ran locally on the Acer baseline.
- Local AI startup, readiness and ordinary conversation were visible.
- Conversation continuity existed across turns.
- The candidate remained visibly private, offline, unsigned and synthetic-test only.

## P0 findings

### Explicit instruction following failed

Prompt:

> What two actions did you say ECO can help with? Answer in four words or fewer.

RC1 returned a longer answer.

Prompt:

> Reply with exactly these three words and nothing else: Organize and explain

Expected:

> Organize and explain

RC1 added further words.

The accepted-answer path did not reject or correct either breach.

### Required correction

- The newest explicit user instruction must outrank conversational continuation.
- Deterministic checks must enforce exact text and bounded word-count requirements where the request is unambiguous.
- A non-compliant model reply must be corrected or rejected before it becomes an accepted answer.
- Stopped, rejected or incomplete text must not update Current Position, tasks, memory or Matter state.

### Complete evidence journey not yet proved

RC1 had not yet proved the complete synthetic journey:

1. clean launch;
2. create or open a named Matter;
3. import synthetic PDF, scanned PDF, image and DOCX evidence;
4. preserve and hash selected evidence;
5. extract text and run OCR;
6. distinguish original, extraction/OCR, correction and note;
7. generate Current Position;
8. answer from Matter sources;
9. open the supporting source location;
10. Stop or reject safely;
11. close and reopen without losing or contaminating state.

## P1 findings

- First visible answer text took roughly 20–27 seconds in the supplied tests and arrived near completion rather than as useful progressive output.
- Technical timings and inference information were too prominent in ordinary answers.
- The active Matter was not sufficiently prominent in the Conversation view.
- Conversations still behaved as the effective home rather than Current Position.
- Current Work displayed internal-style wording rather than a concrete user goal.
- Several controls remained visible when they were not relevant to the current state.
- The full 225% scaling, screen-reader and keyboard journey remained incomplete.

## Preservation requirements for RC2

RC2 must preserve:

- V38 visual continuity;
- working navigation and scrolling;
- conversation tabs and local continuity;
- the warm local Qwen worker path;
- evidence/OCR foundations;
- offline and one-file candidate packaging;
- truthful synthetic-test and unsigned status.

It must not reintroduce the V40 replacement shell.

## Product Lab acceptance route

RC2 should be assessed as one of:

- READY FOR INDEPENDENT INSPECTION;
- RETURN TO DEVELOPMENT;
- EVIDENCE JOURNEY INCOMPLETE;
- AI BEHAVIOUR INCOMPLETE;
- ACCESSIBILITY INCOMPLETE;
- INTAKE INCOMPLETE.

This review changes no release or real-evidence gate.
