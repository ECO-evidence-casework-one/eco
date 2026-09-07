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

## `microsoft/terminal` — native UI Automation provider reference

- Licence: MIT.
- Status: **REFERENCE / bounded donor candidate.**
- Evidence: the current `HwndTerminalWndProc` handles `WM_GETOBJECT`, checks `UiaRootObjectId`, and returns the provider through `UiaReturnRawElementProvider`.
- ECO use: concrete production reference for the Win32 window-procedure/provider boundary when a genuinely custom region must expose a UI Automation tree.
- Boundary: do not copy Terminal's wider architecture or dependency graph. Extract only the provider-boundary principles needed by ECO.

## `MicrosoftDocs/sdk-api` — authoritative UI Automation API contract

- Status: **AUTHORITATIVE PLATFORM REFERENCE.**
- Evidence: the GitHub-hosted Windows SDK documentation states that controls respond to `WM_GETOBJECT` with `UiaReturnRawElementProvider`; the original `wParam`/`lParam` should be passed through; provider maps should be released when the window is destroyed.
- ECO use: implementation and test contract for any app-defined provider path.

## `flutter/flutter` Windows accessibility architecture — reference

- Licence: BSD-style 3-clause licence.
- Status: **REFERENCE — do not import Flutter.**
- Evidence: Flutter's Windows accessibility documentation describes the UIA fragment/root model, `IRawElementProviderFragment`, `IRawElementProviderFragmentRoot`, point/focus navigation, patterns, and the `WM_GETOBJECT` → `UiaReturnRawElementProvider` root path used by Narrator/NVDA.
- ECO use: architecture reference for what a complete custom accessibility tree must expose if native controls cannot represent a region.

## Native Windows UI Automation

- Status: **PLATFORM API — preferred semantics path, not a third-party UI framework.**
- ECO use: standard HWND controls should be preferred for ordinary interactive elements where practical. Genuinely custom interactive regions must expose appropriate UI Automation semantics/provider behaviour rather than relying on pixels/hit testing alone.

## Decision after GitHub reconnaissance

For ECO's **primary navigation, ordinary buttons and simple lists**, converting painted hit regions to standard accessible HWND controls is the lower-risk route and remains preferred.

Use a custom UI Automation provider only for regions where replacing the custom rendering would materially damage required evidence/document interaction. If that route is needed, base the provider boundary on the Microsoft SDK contract and production references above, then verify it with Accessibility Insights, Narrator and NVDA.

## Rejected direction for this stage

A wholesale shell rewrite into a browser/WebView or a new GUI framework is not justified by current evidence. It would change too many variables at once and risks regressing the already-qualified encrypted workspace, evidence preservation, recovery and local-Qwen paths.

Modernisation therefore proceeds by bounded native slices: **semantics/focus → orientation/status → layout/scaling → visual refinement/icons → full assistive-technology acceptance**.
