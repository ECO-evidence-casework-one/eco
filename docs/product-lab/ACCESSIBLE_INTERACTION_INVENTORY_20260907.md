# ECO Product Lab — accessible interaction inventory

**Date:** 7 September 2026  
**Basis:** current native Windows source audit at merged main after PR #141

## Already using standard Windows child controls

These are the strongest starting points for keyboard/screen-reader semantics because they are real HWND controls rather than painted hit regions:

- Ask ECO question input — `EDIT`, Tab-stop.
- Ask ECO submit — `BUTTON`, Tab-stop.
- Ask ECO answer — read-only multiline `EDIT`.
- Evidence search input/button.
- Matters search input/button.
- Current file/matter dialogs that use standard Windows dialogs/controls where invoked.

PR #141 deliberately reuses the Ask read-only `EDIT` for visible Qwen/fallback state rather than adding another custom badge.

## Custom-drawn interactive regions requiring priority remediation

### Primary left navigation — P0 accessibility dependency

Current top-level destinations are painted by ECO and activated through hit rectangles rather than individual focusable HWND navigation controls.

Consequences to prove/fix:

- no assumption of native Button/ListItem role;
- no assumption of per-item Narrator/NVDA name/state;
- mouse hit testing is not equivalent to keyboard focus;
- selected state must be exposed semantically, not only by colour/paint;
- Alt+number shortcuts are useful redundancy but do not replace focusable navigation.

Preferred first engineering route: convert the seven primary destinations to real focusable native controls while retaining the current visual shell, or implement an equally testable UI Automation provider. Prefer native controls unless a concrete visual/behaviour requirement prevents it.

### Home/current-position action cards

Painted action rectangles are convenient visually but ordinary users should be able to Tab to every consequential action, hear its name/purpose/state and activate it using Enter/Space.

### Evidence rows / source actions

Evidence selection/open/reveal actions are painted rows/hit targets. The evidence list must expose item names, selected state, available actions and scrolling position to keyboard and assistive technology.

### Citation/source rows

Source links from Ask answers are painted interactive regions. These are particularly important because a grounded answer is not useful if a keyboard/screen-reader user cannot reach the supporting source.

### Other page-specific painted controls

Review, Changes/Activity, Home and Trust page custom buttons must be enumerated before release and either converted or given UI Automation semantics. No visible control may remain mouse-only without an explicit release-blocking exception.

## Existing keyboard foundations to preserve

Current global shortcuts include page navigation and common operations such as Alt+number destinations, Ctrl+F search and other app shortcuts. These are useful accelerators and should remain, but they are a supplement to logical Tab/Shift+Tab navigation rather than the accessibility mechanism itself.

## Scaling defect to remove

The native app is per-monitor DPI aware, but current font creation caps effective font DPI because earlier layouts clipped at common Windows scaling levels. That is a known workaround, not finished accessibility.

Required route:

1. scale layout geometry and minimum sizes correctly;
2. allow system text/display scaling without the font cap;
3. test at ordinary 100/150/200% display scaling plus Windows text-size enlargement;
4. test the minimum supported display and narrow-window adaptive behaviour.

## Modal/status load

Routine progress, recoverable errors and status are frequently shown through modal message boxes. Keep modal confirmation for genuinely consequential actions; move ordinary progress/warnings/recovery guidance toward an in-window, assistive-technology-readable status surface.

## First implementation order

1. **Primary navigation** — focusable/semantic and selected-state aware.
2. **Exact source route from Ask ECO** — keyboard + screen-reader accessible.
3. **Evidence list/actions** — semantic item model and keyboard scrolling.
4. **Home/Current Position actions** — real controls and simpler task language.
5. **Review/Activity/Trust custom actions.**
6. Remove DPI/font workaround once adaptive layout passes.
7. Accessibility Insights + keyboard + Narrator + NVDA + cognitive walkthrough acceptance.

## Acceptance rule

A control is not accepted merely because it has a keyboard shortcut, painted focus-looking border, tooltip or mouse hit rectangle. Core interactive elements must have a discoverable semantic role/name/state, visible keyboard focus and an operable keyboard activation path.
