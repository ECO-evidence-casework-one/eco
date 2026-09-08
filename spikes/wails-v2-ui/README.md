# ECO Wails v2 UI architecture spike

**NON-PRODUCTION / DO NOT SHIP**

Purpose: prove that Wails v2.14.0 + React/TypeScript can call ECO's UI-neutral `internal/app` service without exposing or bypassing the existing vault, ownership, recovery, CAS, encryption or preservation controls.

This spike proves only:
- application startup and shutdown;
- explicit open existing workspace;
- explicit create new workspace;
- read-only workspace summary;
- create matter;
- explicit close;
- transport-only Wails bindings.

It deliberately does **not** implement evidence import, document preview, search, AI, backup/restore, updates, or the production visual design.

The existing `cmd/eco` Win32 executable remains the production route. Do not retire or alter it based solely on this spike.

## Acceptance before any wider UI work
1. Nested Go module builds independently.
2. React/TypeScript frontend builds.
3. Windows Wails binary builds using v2.14.0.
4. Opening/creating routes preserves `internal/app` safety semantics.
5. Wails shutdown closes the active workspace through `Service.CloseWorkspace()`.
6. Manual Windows acceptance still required for WebView2, keyboard navigation, focus, Narrator/NVDA and DPI/scaling.
