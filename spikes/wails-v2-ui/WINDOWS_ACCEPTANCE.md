# ECO Wails v2 Windows acceptance plan

Status: controlled acceptance gate for the non-production Wails v2 UI spike.
Date: 2026-09-08

## Purpose
Prove that the isolated Wails v2.14 + React/TypeScript spike behaves acceptably on real Windows before any production Win32 screen is retired or replaced.

This document is an acceptance plan, not a conformance claim.

## Scope
The current spike proves only:
- application launch;
- Wails binding to `internal/app`;
- open workspace;
- create workspace;
- close workspace;
- snapshot summary;
- create matter;
- shutdown closes the active workspace.

It does not yet prove evidence import, document viewing, search, AI, backup/restore, updater behaviour, or final visual design.

## Acceptance stack
Use the already-qualified shared Windows stack:
- FlaUI — deterministic Windows UI driving and focus/control journeys;
- Axe.Windows — automated UI Automation accessibility scanning;
- Accessibility Insights for Windows — external human inspection with telemetry disabled;
- NVDA stable release — real screen-reader acceptance;
- manual keyboard, scaling, high-contrast and visual checks.

Automated scanning is one layer only. Passing scans must never be described as WCAG/EN 301 549/Section 508 conformance.

## Synthetic-data rule
Use only synthetic workspaces and synthetic matters. No private or real evidence is required for this acceptance gate.

## Required journeys

### A. Launch and initial state
- App launches without browser chrome or unexpected network/account prompts.
- Window title identifies the spike clearly as non-production.
- Initial focus is visible and predictable.
- All actionable controls expose meaningful accessible names/roles.

### B. Keyboard-only operation
- Complete all current spike operations without a mouse.
- Tab/Shift+Tab order is logical.
- Enter/Space activate the expected controls.
- Focus is not lost after success, validation error, workspace open/create or close.
- No destructive action is the accidental default.

### C. Workspace lifecycle
- Open existing never creates a missing workspace.
- Create workspace refuses an existing route.
- Close releases ownership.
- Shutdown while a workspace is open closes the service cleanly.
- Reopen after clean shutdown succeeds.

### D. Screen-reader acceptance
Using Narrator and NVDA separately:
- application and main region are identifiable;
- fields have meaningful labels;
- buttons announce names and states;
- errors are announced without visual-only dependence;
- status changes are discoverable;
- matter list/summary information has usable reading order.

### E. Scaling and layout
Exercise at minimum:
- 100%;
- 125%;
- 150%;
- 175%;
- 200% display/text scaling where supported.

At every scale:
- no clipped action labels;
- no overlapping controls;
- no hidden safety/error text;
- keyboard focus remains visible;
- minimum window remains usable;
- resizing does not cause an unrecoverable layout.

### F. High contrast and colour independence
- Windows high-contrast mode remains readable.
- Focus indicator remains visible.
- Success/error/disabled states are not communicated by colour alone.
- Text contrast remains usable in supported themes.

### G. WebView2/runtime behaviour
- Existing WebView2 runtime path launches correctly.
- Missing/broken runtime failure is understandable and does not damage workspace state.
- No unexpected network dependency is required for normal local operation after dependencies are installed.
- Browser developer context/menu is not exposed in production-style builds unless deliberately enabled.

### H. Shutdown/recovery
- Window close invokes Wails `OnShutdown` and `Service.CloseWorkspace()`.
- Reopen does not report stale ownership after a clean exit.
- Forced termination is treated separately and must not be confused with clean shutdown acceptance.

## Automated evidence to capture
For each automated run record:
- exact ECO commit;
- exact spike build hash;
- Windows version;
- display scale;
- FlaUI scenario identifier and PASS/FAIL;
- Axe.Windows error/warning counts;
- local report SHA-256 where a report is produced;
- timestamp;
- synthetic fixture identity.

## Human evidence to capture
For each manual acceptance session record:
- tester;
- exact ECO commit/build;
- Windows version;
- display/text scale;
- theme/high-contrast state;
- screen reader and version;
- journey exercised;
- PASS/FAIL;
- concise observations;
- known limitation reference if not fully passing.

## Promotion gate
The spike may support production Startup/Workspace Chooser implementation only when:
1. dedicated Wails CI is green;
2. normal ECO full CI is green;
3. keyboard-only core journey passes;
4. Narrator core journey passes;
5. NVDA core journey passes;
6. 100/125/150/175/200% scaling checks pass or have explicitly accepted bounded defects;
7. high-contrast check passes;
8. shutdown/reopen ownership check passes;
9. no critical accessibility or workspace-safety defect remains open.

No Win32 screen is retired by passing this gate. Production replacement requires separate screen-level parity, regression and accessibility acceptance.
