# ECO Product Lab — modern accessible shell baseline

**Date:** 7 September 2026  
**Scope:** product/UX/accessibility control record; no application source change

## Decision

ECO will modernise the current native Windows shell incrementally. It will **not** replace the working application architecture merely to obtain a fashionable UI framework.

The existing evidence vault, workspace/recovery controls, local-tool registry, search, grounding and newly connected llama.cpp/Qwen path remain the product foundation. Visual work must not regress them.

## GitHub-first reference set

### `microsoft/WinUI-Gallery` — MIT — reference

Use as a current Windows/Fluent interaction and layout reference. It is actively maintained in September 2026. ECO may reproduce appropriate design principles and interaction patterns without importing WinUI/XAML as a new application architecture.

### `microsoft/fluentui-system-icons` — MIT — candidate donor

Candidate source for recognisable Windows-style icons. Primary navigation must keep visible text labels; icons are reinforcement, not the only label.

### `microsoft/accessibility-insights-windows` — MIT — acceptance tool

Use for retained Windows accessibility inspection evidence, including FastPass/manual inspection where applicable. Tool output does not replace Narrator/NVDA, keyboard-only or cognitive-walkthrough testing.

## Current ECO shell — facts

The current native shell already has useful foundations:

- one stable left navigation area;
- seven top-level destinations;
- native Windows edit/button controls for the Ask and search inputs;
- keyboard shortcuts for navigation and common actions;
- per-monitor DPI awareness;
- explicit Low Sensory setting;
- cards, whitespace and a restrained teal/neutral visual language.

It also has material accessibility/product defects:

- primary sidebar navigation is custom-drawn hit regions rather than normal focusable controls;
- many in-page buttons are custom-drawn mouse hit regions and are not ordinary Tab stops;
- evidence rows and citation rows are custom-drawn interactive regions;
- the Win32 interop layer currently has no explicit UI Automation / `WM_GETOBJECT` provider path for those app-defined interactive regions;
- font DPI is deliberately capped because earlier layouts clipped at normal Windows scaling, so present scaling is not a finished accessibility solution;
- routine information/errors are often presented through blocking message boxes;
- the Ask screen still describes the old deterministic-only mode even though the application-facing Ask route can now invoke verified local llama.cpp;
- the UI does not yet prove whether Qwen actually ran or deterministic fallback was used.

## Information architecture

Keep the left-navigation model, but make its hierarchy clearer.

### Primary work

1. **Home / Current Position** — where am I, what matters now, what should I do next?
2. **Matters** — choose or organise the workstream.
3. **Evidence** — add, find and inspect sources.
4. **Ask ECO** — ask questions and see the exact AI/source state.

### Checking/history

5. **Review** — items requiring human attention.
6. **Activity** — user-readable history; evaluate renaming the current `Changes` label after compatibility review.

### Secondary/footer

7. **Trust & settings** — security, local runtimes, backup/restore and accessibility preferences.

No navigation rename is authorised until source/tests and user-facing terminology are checked for compatibility. The list above is the Product Lab target, not a claim that source has already changed.

## Orientation contract

Every core page must make these five answers obvious without relying on memory:

1. **Where am I?** — page name and active Matter/context.
2. **What is this for?** — one short plain-language sentence.
3. **What can I do here?** — the primary action(s), visually and semantically obvious.
4. **What happens next?** — current status/next action, not an internal engineering state.
5. **How do I get back?** — stable left navigation and visible Home/Current Position route.

## AI-state contract

The Ask screen must never make a deterministic answer look like proof that the model ran.

Required states:

- **Qwen ready** — verified runtime/model configuration is usable locally.
- **Qwen running** — generation is actively executing.
- **Qwen checked** — Qwen ran and ECO accepted only grounded claims.
- **Source fallback used** — Qwen did not produce an accepted answer; deterministic source engine answered instead.
- **Qwen unavailable** — no usable configured model/runtime.
- **Qwen rejected** — model output failed ECO grounding/validation and was not released.

Routine state belongs in the page, not a modal dialog.

## Cognitive-accessibility rules

- use familiar, concrete words;
- use the same label for the same function everywhere;
- keep important controls in stable places;
- keep primary navigation text visible at ordinary desktop widths;
- do not require the user to remember hidden gestures or icon meanings;
- break multi-step tasks into visible steps;
- use progressive disclosure for technical/audit detail;
- preserve user work when errors occur;
- error messages state what happened, what was preserved, and the safe next action;
- avoid automatic context switches and moving controls;
- avoid unnecessary animation; Low Sensory must remain respected;
- important state must not be conveyed only by colour.

## Windows accessibility engineering order

1. Convert the most important custom-drawn interactive regions to accessible HWND controls where practical, starting with primary navigation and core action buttons.
2. Give those controls stable text/accessibility names, logical Tab order and visible focus.
3. For remaining genuinely custom interactive regions, implement and test the required UI Automation provider boundary rather than assuming screen readers can infer it.
4. Remove the current font-scaling workaround by making layout measurements scale/adapt correctly.
5. Respect Windows text scaling and contrast themes; do not rely only on hard-coded colours.
6. Run Accessibility Insights for Windows, keyboard-only, Narrator and NVDA passes.

## Visual direction

ECO should feel like a current Windows productivity application, not a web dashboard copied into a native window and not an old Win32 utility.

- stable left navigation;
- strong page title + short purpose line;
- active Matter/context visible in the shell;
- one obvious primary action per section where possible;
- restrained rounded surfaces/cards only where they aid grouping;
- deliberate whitespace;
- standard Windows/system typography and responsive text/layout;
- Fluent System Icons only where they improve recognition;
- non-modal InfoBar-like status surfaces for routine progress, warnings and recoverable errors;
- adaptive content layout at narrower widths while preserving text labels whenever possible;
- advanced provenance/security detail available without dominating the normal workflow.

## Acceptance rule

A screenshot that looks modern is **not** acceptance.

The shell is accepted only when a first-time user can complete the synthetic Matter → Evidence → inspect → Ask ECO → source route with clear orientation; keyboard-only operation works; Narrator/NVDA expose meaningful names/roles/state; Windows scaling/text-size/contrast remain usable; and AI status truthfully distinguishes real Qwen execution from fallback.
