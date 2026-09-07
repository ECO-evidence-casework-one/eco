# ECO Product Lab — accessible interaction inventory

**Date:** 7 September 2026  
**Basis:** native Windows source audit, refreshed from current `main` `5623de397fed16185348095b19fe692afe00e5d5`

## Current AI/accessibility boundary

The application-facing Ask route can use verified local llama.cpp/Qwen and the native Ask surface now exposes truthful model/fallback states. The controlled rig hand-off also has visible/resumable Qwen download controls. **Real end-to-end rig generation and Ask acceptance remains a human gate.**

Do not modify that qualified AI baseline while accessibility shell experiments are still being proved.

## Already using standard Windows child controls

These are the strongest starting points for keyboard/screen-reader semantics because they are real HWND controls rather than painted hit regions:

- Ask ECO question input — `EDIT`, Tab-stop.
- Ask ECO submit — `BUTTON`, Tab-stop.
- Ask ECO answer/status — read-only multiline `EDIT`.
- Evidence search input/buttons.
- Standard Windows file/folder dialogs where invoked.

The visible AI-state work deliberately reuses a native read-only `EDIT` rather than adding another custom-painted status badge.

## Custom-drawn interactive regions requiring priority remediation

### Primary left navigation — P0 accessibility dependency

Top-level destinations are painted by ECO and activated through hit rectangles rather than individual focusable HWND navigation controls.

Consequences to prove/fix:

- no assumption of native Button/ListItem role;
- no assumption of per-item Narrator/NVDA name/state;
- mouse hit testing is not keyboard focus;
- selected state must be exposed semantically, not only through paint/colour;
- shortcuts are useful redundancy but do not replace focusable navigation.

Preferred route: convert the seven primary destinations to real focusable native controls while retaining the shell layout, unless a concrete requirement proves a custom UI Automation provider is safer.

### Home / Current Position actions

Painted action rectangles must become focusable/semantic. A user should be able to Tab to consequential actions, hear their name/purpose/state, and activate them using Enter/Space.

### Evidence rows / source actions

Evidence selection/open/reveal actions are painted rows/hit targets. The list must expose item names, selected state, available actions and scrolling position to keyboard and assistive technology.

### Citation/source rows

Supporting-source links from Ask answers are painted interactive regions. These are a high priority because a grounded answer is not operationally useful if the supporting source cannot be reached without a mouse.

### Other page-specific painted controls

Review, Activity/Changes, Home and Trust page custom buttons must be enumerated and either converted or provided with tested UI Automation semantics. No core visible control may remain mouse-only without an explicit release-blocking exception.

## Existing keyboard foundations to preserve

Global shortcuts and search accelerators are useful, but they supplement logical Tab/Shift+Tab navigation rather than replacing it.

The earlier draft Tab-navigation experiment compiled and passed CI, but it remains **unaccepted** because real keyboard behaviour has not been physically exercised against the current Qwen baseline. Do not merge that stale branch into `main`.

## Scaling defect to remove

The app is per-monitor DPI aware, but font creation currently caps effective font DPI because earlier fixed-pixel layouts clipped at common Windows scaling levels. That is a workaround, not finished accessibility.

Required route:

1. scale layout geometry and minimum sizes correctly;
2. allow system text/display scaling without the font cap;
3. test 100%, 150% and 200% display scaling plus Windows text-size enlargement;
4. test the minimum supported display and narrow-window adaptive behaviour.

## Modal/status load

Routine progress, recoverable errors and status are frequently shown through modal message boxes. Keep modal confirmation for genuinely consequential actions; move routine progress/warnings/recovery guidance toward an in-window, assistive-technology-readable status surface.

## First implementation order after real Qwen acceptance

1. **Primary navigation** — focusable/semantic and selected-state aware.
2. **Exact source route from Ask ECO** — keyboard + screen-reader accessible.
3. **Evidence list/actions** — semantic item model and keyboard scrolling.
4. **Home/Current Position actions** — real controls and simpler task language.
5. **Review/Activity/Trust custom actions.**
6. Remove DPI/font workaround once adaptive layout passes.
7. Accessibility Insights + keyboard + Narrator + NVDA + cognitive walkthrough acceptance.

## Acceptance rule

A control is not accepted merely because it has a shortcut, painted focus-looking border, tooltip or mouse hit rectangle. Core interactive elements require a discoverable semantic role/name/state, visible keyboard focus and a working keyboard activation path.
