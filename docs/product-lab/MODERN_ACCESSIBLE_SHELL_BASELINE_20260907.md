# ECO Product Lab — modern accessible shell baseline

**Date:** 7 September 2026  
**Application baseline:** `5623de397fed16185348095b19fe692afe00e5d5`  
**Scope:** product/UX/accessibility control record; no application source change

## Decision

ECO will modernise the current native Windows shell incrementally. It will **not** replace the working application architecture merely to obtain a fashionable UI framework.

The encrypted evidence vault, workspace/recovery controls, local-tool registry, search, grounding, visible Qwen/fallback state and controlled rig hand-off remain the product foundation. Visual work must not regress them.

## GitHub-first reference set

### `microsoft/WinUI-Gallery` — MIT — reference
Use as a current Windows/Fluent interaction/layout reference. Appropriate principles may be reproduced without importing WinUI/XAML as a new application architecture.

### `microsoft/fluentui-system-icons` — MIT — candidate donor
Candidate source for recognisable Windows-style icons. Primary navigation keeps visible text labels; icons reinforce meaning rather than replace labels.

### `microsoft/accessibility-insights-windows` — MIT — acceptance tool
Use for retained Windows accessibility inspection evidence. Tool output does not replace Narrator/NVDA, keyboard-only or cognitive-walkthrough testing.

## Current shell — verified direction

Useful foundations to preserve:

- one stable left navigation area;
- seven top-level destinations;
- native Windows edit/button controls for Ask and search inputs;
- keyboard shortcuts for navigation/common actions;
- per-monitor DPI awareness;
- Low Sensory setting;
- restrained card/whitespace visual language;
- visible native Qwen/fallback state in Ask ECO;
- a controlled local-AI rig route with hash-pinned runtime/model identities and resumable visible model transfer.

Material gaps still blocking accessibility acceptance:

- primary sidebar navigation remains custom-drawn hit regions rather than normal focusable controls;
- many in-page actions remain custom-drawn mouse hit regions;
- evidence rows and citation/source rows remain custom-drawn interactive regions;
- no complete verified UI Automation provider path exists for app-defined interactive regions;
- font DPI remains capped as a workaround for fixed-layout clipping;
- routine information/errors still rely heavily on blocking message boxes;
- real keyboard-only, Narrator, NVDA, scaling, contrast and cognitive-walkthrough evidence has not yet been retained.

## Information architecture

Keep the left-navigation model and shallow hierarchy.

### Primary work
1. **Home / Current Position** — where am I, what matters now, what should I do next?
2. **Matters** — choose or organise the workstream.
3. **Evidence** — add, find and inspect sources.
4. **Ask ECO** — ask questions and see exact AI/source state.

### Checking/history
5. **Review** — items requiring human attention.
6. **Activity** — user-readable history; any rename from current terminology requires compatibility review.

### Secondary/footer
7. **Trust & settings** — security, local runtimes, backup/restore and accessibility preferences.

No navigation rename is authorised merely by this Product Lab record.

## Orientation contract

Every core page must make five answers obvious without relying on memory:

1. **Where am I?** — page name and active Matter/context.
2. **What is this for?** — one short plain-language sentence.
3. **What can I do here?** — primary action(s), visually and semantically obvious.
4. **What happens next?** — current status/next action, not internal engineering state.
5. **How do I get back?** — stable navigation and a visible Home/Current Position route.

## AI-state contract

The Ask screen must never make deterministic retrieval look like proof that Qwen ran.

Required states already represented in the current product path include:

- **Qwen ready** — verified runtime/model configuration is usable locally.
- **Qwen running** — generation is executing.
- **Qwen checked** — Qwen ran and ECO accepted the grounded result.
- **Source fallback used** — deterministic source engine answered instead.
- **Qwen unavailable** — no usable configured model/runtime.
- **Qwen rejected** — model output failed validation/grounding and was not released.

Real rig generation + end-to-end Ask acceptance remains outstanding and is the freeze point before major shell source changes.

## Cognitive-accessibility rules

- use familiar, concrete words;
- use the same label for the same function everywhere;
- keep important controls in stable places;
- keep primary navigation text visible at ordinary desktop widths;
- do not require hidden gestures or memorised icon meanings;
- break multi-step tasks into visible steps;
- use progressive disclosure for technical/audit detail;
- preserve work when errors occur;
- error messages state what happened, what was preserved, and the safe next action;
- avoid automatic context switches and moving controls;
- avoid unnecessary animation and respect Low Sensory;
- do not convey important state by colour alone.

## Windows accessibility engineering order

After the real-Qwen rig baseline is physically accepted:

1. convert the most important custom-drawn interactive regions to accessible HWND controls where practical, starting with primary navigation;
2. give controls stable text/accessibility names, logical Tab order and visible focus;
3. for remaining genuinely custom interactive regions, implement/test the UI Automation provider boundary;
4. make layout measurements scale/adapt correctly and remove the current font-scaling cap;
5. respect Windows text scaling and contrast themes;
6. run Accessibility Insights, keyboard-only, Narrator and NVDA passes;
7. retain acceptance evidence against the complete synthetic Matter → Evidence → Ask → source journey.

## Visual direction

ECO should feel like a current Windows productivity application, not a browser dashboard copied into a native window and not an old utility.

- stable left navigation;
- strong page title + short purpose line;
- active Matter/context visible;
- one obvious primary action per section where possible;
- restrained rounded surfaces/cards only where grouping benefits;
- deliberate whitespace;
- system-native typography and responsive layout;
- Fluent System Icons only where they improve recognition;
- non-modal InfoBar-like status for routine progress/warnings/recoverable errors;
- adaptive hierarchy at narrower widths;
- advanced provenance/security detail available without dominating normal use.

## Acceptance rule

A modern screenshot is **not** acceptance.

The shell is accepted only when a first-time user can complete the synthetic Matter → Evidence → inspect → Ask ECO → source route with clear orientation; keyboard-only operation works; Narrator/NVDA expose meaningful names/roles/state; Windows scaling/text-size/contrast remain usable; and AI status truthfully distinguishes real Qwen execution from fallback.
