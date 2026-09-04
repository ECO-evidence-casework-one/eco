# ECO RC2 — Product Acceptance Matrix

**Date:** 6 August 2026  
**Scope:** Product Lab review of the next V38-based private synthetic candidate

## P0 — must pass

| ID | Area | Pass condition |
|---|---|---|
| RC2-P0-01 | Clean launch | No earlier candidate state is silently loaded. |
| RC2-P0-02 | V38 continuity | Candidate is recognisably derived from V38 and does not reintroduce the V40 shell. |
| RC2-P0-03 | Matter identity | Active Matter is always visible and conversations remain Matter-scoped. |
| RC2-P0-04 | Instruction following | Exact-text and word-count prompts are obeyed or deterministically rejected/corrected. |
| RC2-P0-05 | Evidence preservation | Every selected occurrence is accounted for and the preserved source is hashed before downstream use. |
| RC2-P0-06 | Source separation | Original, extraction, OCR, correction, generated suggestion and user note remain distinguishable. |
| RC2-P0-07 | Source-backed answer | Evidence claims identify supporting Matter sources and do not invent missing facts. |
| RC2-P0-08 | Exact source route | Important conclusions open the exact or nearest clearly labelled supporting source location. |
| RC2-P0-09 | Stop/rejection isolation | Stopped, rejected or unfinished output cannot update Matter state, tasks, memory or Current Position. |
| RC2-P0-10 | Persistence | Closing and reopening restores the correct Matter and conversation without cross-build contamination. |
| RC2-P0-11 | AI unavailable | Current Position, Evidence and prior work remain usable when local AI is unavailable. |
| RC2-P0-12 | Synthetic-only status | Candidate remains clearly private, unsigned and synthetic-test only. |

## P1 — required before independent inspection readiness

| ID | Area | Pass condition |
|---|---|---|
| RC2-P1-01 | Current Position | Current Position is the understandable Matter home. |
| RC2-P1-02 | Evidence journey | PDF, scanned PDF, image and DOCX test items complete the intended import/read/OCR journey. |
| RC2-P1-03 | Different accounts | Materially different source statements are shown without automatically choosing one. |
| RC2-P1-04 | Missing expected evidence | ECO explains the source basis for expecting an absent item. |
| RC2-P1-05 | OCR uncertainty | Unclear OCR remains uncertain until the original is reviewed or corrected. |
| RC2-P1-06 | Stale output | A changed source or correction marks dependent earlier output for review. |
| RC2-P1-07 | Streaming lifecycle | Visible generated text is clearly unfinished until verification completes. |
| RC2-P1-08 | Technical detail | Timings, receipts and internal route detail are secondary and available behind Details. |
| RC2-P1-09 | Current Work | Displays a real user goal or a calm empty state, never an internal AI instruction. |
| RC2-P1-10 | Responsiveness | Window remains usable during AI, import and OCR work. |
| RC2-P1-11 | Keyboard | Core journey can be completed without a mouse and with visible focus. |
| RC2-P1-12 | Scaling | Core journey remains usable at 100%, 150%, 200% and 225% where supported. |
| RC2-P1-13 | Screen reader | Principal controls, states, progress and errors are meaningfully exposed to Narrator/NVDA. |
| RC2-P1-14 | What’s New | Claims only changes actually visible and implemented in the candidate. |

## Fixed RC1 regression prompts

1. `In one sentence, tell me what ECO can help me do. Do not invent facts.`
2. `What two actions did you say ECO can help with? Answer in four words or fewer.`
3. `Reply with exactly these three words and nothing else: Organize and explain`
4. `Your previous answer broke my instruction. Explain in one sentence what went wrong.`
5. `What deadline do I have?`
6. `Where do I add evidence?`

Expected exact answer for prompt 3:

> Organize and explain

With no supporting Matter source, prompt 5 must not invent a deadline.

## Synthetic Matter journey

Use only the controlled `Community Hall Boiler Repair` fixture.

Required questions:

- What date did the engineer attend? Show the source.
- Which completion dates differ? Do not choose between them.
- What document is missing, and why does ECO expect it?
- What is the boiler serial number? Do not guess if it is unclear.
- Is the serial inside the warranty schedule range?
- What does that file say?

The ambiguous prompt must request clarification when more than one file is plausible.

## Required evidence return

- exact build and source identities;
- executable size and SHA-256;
- implemented, deferred and blocked items;
- raw test results;
- same-dimension RC1/RC2 screenshots;
- Acer startup, first-text, total-time, memory and CPU evidence;
- clean launch, import, OCR, source answer, exact source, Stop and close/reopen evidence;
- keyboard, scaling and screen-reader evidence.

## Decision

- [ ] READY FOR INDEPENDENT INSPECTION
- [ ] RETURN TO DEVELOPMENT
- [ ] EVIDENCE JOURNEY INCOMPLETE
- [ ] AI BEHAVIOUR INCOMPLETE
- [ ] ACCESSIBILITY INCOMPLETE
- [ ] INTAKE INCOMPLETE

No box may be selected without the corresponding evidence.
