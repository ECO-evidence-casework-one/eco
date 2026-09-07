# ECO FOSS register — modern Windows/accessibility donors

**Date:** 7 September 2026  
**Decision class:** Product Lab / donor control; no automatic dependency adoption

## `microsoft/WinUI-Gallery`

- Licence: MIT.
- Status: **REFERENCE — do not import as an application framework in the current slice.**
- Reason: current Microsoft/Windows interaction examples are useful for NavigationView hierarchy, page headers, InfoBar-style status, settings placement, spacing and adaptive layout.
- ECO use: benchmark current native Win32 shell behaviour/appearance against the patterns without replacing the evidence/vault/local-AI architecture with XAML/WinUI.

## `microsoft/fluentui-system-icons`

- Licence: MIT.
- Status: **CANDIDATE DONOR.**
- Reason: current Windows-recognisable icon set, available as source assets.
- ECO use: only where an icon improves recognition. Primary destinations retain visible text labels; icon-only primary navigation is rejected for cognitive-accessibility reasons.
- Adoption gate: choose a bounded subset, preserve upstream licence/provenance, convert/package only the necessary assets, and verify contrast/scaling.

## `microsoft/accessibility-insights-windows`

- Licence: MIT.
- Status: **ADOPT AS DEVELOPMENT/ACCEPTANCE TOOL, not runtime dependency.**
- Reason: Windows accessibility inspection and FastPass/manual testing support.
- ECO use: retained accessibility acceptance evidence alongside keyboard-only, Narrator and NVDA tests.
- Boundary: passing an automated inspection does not by itself prove cognitive accessibility, complete screen-reader usability or legal conformance.

## Native Windows UI Automation

- Status: **PLATFORM API — preferred semantics path, not a third-party UI framework.**
- ECO use: standard HWND controls should be preferred for ordinary interactive elements where practical. Genuinely custom interactive regions must expose appropriate UI Automation semantics/provider behaviour rather than relying on pixels/hit testing alone.

## Rejected direction for this stage

A wholesale shell rewrite into a browser/WebView or a new GUI framework is not justified by the current evidence. It would change too many variables at once and risks regressing the already-merged encrypted workspace, evidence preservation, recovery and local-AI paths.

Modernisation will therefore proceed by bounded native slices: semantics/focus → orientation/status → layout/scaling → visual refinement/icons → full assistive-technology acceptance.
