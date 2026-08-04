# V40 first usable Matter journey

**Status:** controlling visual and interaction contract for the next native application milestone.  
**Prototype:** `design/v40_first_usable_matter_journey.html`  
**Issue:** #69  
**Date:** 4 August 2026

## Purpose

V40 must convert ECO from a collection of technical foundations into a visibly coherent local casework application.

The interactive HTML prototype is a review instrument only. It is not the Windows application, does not persist information, does not analyse evidence and must never be described as product implementation or release evidence.

The native Windows application must reproduce the essential hierarchy, language and behaviour through native controls and the existing offline architecture.

## Required core journey

1. Start ECO.
2. See a calm Home screen with the current workspace and three obvious actions:
   - Create a Matter;
   - Add evidence;
   - Reopen workspace.
3. Create a Matter using a short guided form:
   - Matter name;
   - optional reference;
   - objective.
4. Enter the Matter workspace.
5. See these sections without navigating through unrelated screens:
   - Current position;
   - What the evidence says;
   - Next actions;
   - Evidence in this Matter.
6. Add synthetic evidence and see truthful stages:
   - checking file type;
   - hashing locally;
   - preserving encrypted copy;
   - verifying preserved bytes;
   - updating the Matter.
7. Cancel an intake without leaving a false completed record.
8. Ask a question and receive either:
   - a verified source-linked response through the controlled M1.18 adapter; or
   - a clear unavailable/fallback state.
9. Close ECO.
10. Reopen the same workspace and Matter with the same identity and no stale-writer overwrite.
11. Open What’s New and review every material visible, functional, safety and limitation change.

## Visual hierarchy

### Home

Home must answer three questions immediately:

- Which workspace is open?
- What needs attention?
- What can I do next?

It must contain:

- local/offline status;
- current workspace identity and disposition;
- Create a Matter as the primary action;
- recent Matters;
- concise metrics;
- recent important changes.

It must not lead with implementation terminology, model names, schemas, hashes or engineering milestones.

### Matter workspace

The Matter workspace is the primary product surface.

The overview must visually separate:

1. **Document says** — wording supported by a preserved source;
2. **App status/suggestion** — processing state or ECO proposal;
3. **You confirmed** — user-confirmed wording or decision;
4. **Your note** — user-authored information not asserted as evidence.

The first view must not be an empty generic dashboard. It must show current position, evidence status and next actions.

### Evidence

Every evidence row must show:

- safe display name;
- type;
- preservation state;
- source verification state;
- reading/OCR state;
- linked Matter count or selected Matter;
- any required review.

No item may say Ready when only preservation has completed but reading or verification is still pending.

### Ask ECO

Ask ECO must not look or behave like an unrestricted generic chatbot.

The response layout must identify:

- Current position;
- Suggested next action;
- named sources;
- support/uncertainty;
- any excluded or failed sources.

Raw generated text must not appear as accepted output before deterministic verification.

### What’s New

What’s New must be available after first launch and later from the top bar.

It must include:

- visible changes;
- functional changes;
- corrected defects;
- accessibility changes;
- performance changes;
- security and privacy changes;
- known limitations;
- exact build identity.

It must distinguish implemented changes from planned work.

## Accessibility requirements

The core journey must support:

- keyboard-only operation;
- visible focus;
- logical tab order;
- 125%, 150% and 200% Windows scaling;
- low-sensory presentation;
- reduced motion;
- plain-language labels;
- no information conveyed by colour alone;
- status updates exposed to assistive technology;
- cancellation reachable without pointer input.

The design prototype is not accessibility evidence. Native validation is required.

## Responsiveness and processing

Long-running work must not block the Windows message thread.

Every operation over approximately 300 milliseconds must provide an honest state. Evidence intake must show its current stage, not a generic spinner.

Cancellation must be cooperative and must result in one of these explicit states:

- cancelled before preservation;
- preserved but not indexed;
- verified and completed;
- failed safely with rollback or recoverable preservation record.

## Public pre-alpha boundary

The V40 source release may be public when issue #69 source gates pass.

A binary must remain private unless its separate gates pass.

The first public version must be labelled **pre-alpha** and must state:

- fully offline architecture;
- free/open-source project status;
- synthetic qualification status;
- unsigned status where applicable;
- no approval for irreplaceable evidence;
- incomplete OCR/model/accessibility claims;
- exact build-from-source procedure.

## Non-goals for the first V40 public pre-alpha

The following are not required for the minimum public source release:

- every document format;
- real generative model integration beyond one controlled synthetic adapter path;
- final installer;
- code signing;
- cloud sync;
- accounts;
- collaboration;
- real-evidence qualification;
- medical, legal, forensic or authority decision functions.

## Implementation order

1. Complete workspace ownership/integrity P0 repair.
2. Implement Home and What’s New native surfaces.
3. Implement guided Matter creation.
4. Implement Matter overview and evidence linking.
5. Implement responsive synthetic intake with cancellation.
6. Connect one controlled M1.18 synthetic adapter path.
7. Qualify close/reopen and stale-writer protection.
8. Perform accessibility and Acer baseline tests.
9. Freeze and review the public pre-alpha source release.
