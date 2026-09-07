# ECO FOSS register — modern Windows/accessibility donors

**Date:** 7 September 2026  
**Application baseline:** `5623de397fed16185348095b19fe692afe00e5d5`  
**Decision class:** Product Lab / donor control; no automatic dependency adoption

## `microsoft/WinUI-Gallery`

- Licence: MIT.
- Status: **REFERENCE — do not import as an application framework in the current slice.**
- Reason: current Windows/Fluent interaction examples are useful for navigation hierarchy, page headers, InfoBar-style status, settings placement, spacing and adaptive layout.
- ECO use: benchmark the current native Win32 shell against these interaction/layout patterns without replacing the evidence/vault/recovery/local-AI architecture with XAML/WinUI.

## `microsoft/fluentui-system-icons`

- Licence: MIT.
- Status: **CANDIDATE DONOR.**
- Reason: recognisable current Windows icon set available as source assets.
- ECO use: only where an icon materially improves recognition. Primary destinations retain visible text labels; icon-only primary navigation is rejected for cognitive-accessibility reasons.
- Adoption gate: choose a bounded subset, preserve upstream licence/provenance, package only required assets, and verify contrast/scaling.

## `microsoft/accessibility-insights-windows`

- Licence: MIT.
- Status: **ADOPT AS DEVELOPMENT/ACCEPTANCE TOOL, not runtime dependency.**
- Reason: Windows accessibility inspection and FastPass/manual testing support.
- ECO use: retain accessibility acceptance evidence alongside keyboard-only, Narrator and NVDA testing.
- Boundary: automated inspection alone does not prove cognitive accessibility, complete screen-reader usability or legal conformance.

## Native Windows UI Automation

- Status: **PLATFORM API — preferred semantics path, not a third-party UI framework.**
- ECO use: standard HWND controls should be preferred for ordinary interactive elements where practical. Genuinely custom interactive regions must expose appropriate UI Automation semantics/provider behaviour rather than relying on pixels/hit testing alone.

## Rejected direction for this stage

A wholesale shell rewrite into a browser/WebView or a new GUI framework is not justified by current evidence. It would change too many variables at once and risks regressing the already-qualified encrypted workspace, evidence preservation, recovery and local-Qwen paths.

Modernisation therefore proceeds by bounded native slices: **semantics/focus → orientation/status → layout/scaling → visual refinement/icons → full assistive-technology acceptance**.
