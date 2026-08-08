# Native Win32 Accessibility Spike

This directory is an isolated presentation-architecture experiment for Evidence & Casework One (ECO).

It is **not** the ECO application, not a release candidate and not approved for real evidence.

## Purpose

The original P2 renderer bake-off proved that Gio and Fyne can render the representative synthetic screens, but neither original candidate currently meets ECO's Windows accessibility requirements as-is. Cogent Core was also audited and did not provide a Windows screen-reader bridge.

This spike tests a different architecture:

- Go application logic;
- direct Win32 window creation;
- standard Windows controls for semantic and interactive UI;
- optional custom drawing only for presentation/decorative surfaces;
- no cloud, network, telemetry or browser runtime;
- one Windows GUI executable;
- `CGO_ENABLED=0`;
- no third-party Go module in this proof.

## Current controls

The synthetic workspace creates real standard Windows controls:

- `STATIC` text labels;
- `BUTTON` navigation/actions;
- `EDIT` search input;
- `LISTBOX` evidence list.

The controls are intentionally real HWND-backed controls so Windows accessibility technology can obtain established semantic contracts rather than relying on a canvas renderer to invent them later.

## Current CI proof

Workflow: `.github/workflows/p2-native-win32-accessibility-spike.yml`

Green reference run: `31267931187`

Reference commit: `6bfa0b08fad38b64ab91657e95767fbb972176e2`

Observed build:

- Go 1.23.12 windows/amd64;
- `go vet ./...` PASS;
- `CGO_ENABLED=0` PASS;
- one-file GUI build PASS;
- size 2,305,536 bytes;
- SHA-256 `bd48f60f43306265f6c2fb4b5b5e703283120600a77f3ffce18f61654abf06ae` for that ephemeral CI build;
- no workflow artifact upload.

Direct Oleacc/MSAA inspection of the real child HWNDs passed:

- static text role 41;
- push-button role 43;
- text/edit role 42;
- list role 33;
- button/edit/list focusability;
- button accessible names;
- Review evidence default action.

The hosted Windows Server 2025 managed UIA client still flattens those standard controls to `ControlType.Pane` in the non-interactive runner session. That result is explicitly treated as inconclusive. Physical Windows 11 validation with NVDA and Narrator remains mandatory before production promotion.

## Build locally

From this directory on Windows:

```powershell
go vet ./...
$env:CGO_ENABLED='0'
go build -trimpath -ldflags '-H windowsgui' -o eco-native-a11y-spike.exe .
```

The generated EXE is a development spike only. Do not treat it as an ECO release.

## Accessibility probe

`probe_accessibility.ps1` is CI/test support, not product runtime code. It uses Windows PowerShell and the Windows accessibility interop assemblies to inspect the standard-control HWNDs directly through `IAccessible`.

## Architecture rule

The next visual prototype must keep semantic/interactive elements native or provide an equivalently proven Windows accessibility provider.

Custom drawing is allowed for backgrounds, cards, separators, icons and other presentation-only elements. It must not silently replace edit controls, buttons, lists, trees, tabs, checkboxes, radio buttons, menus or other user interactions with inaccessible painted pixels.

Any owner-drawn native control must be re-run through the accessibility gates after styling.

## Known limitations

- this spike does not implement the whole ECO screen set;
- it does not yet reproduce the locked visual specification;
- it has not yet been tested with NVDA or Narrator on a physical interactive Windows 11 desktop;
- scaling, contrast themes, Windows text-size settings and complete keyboard traversal remain unverified;
- raw Win32 calls need to be isolated behind a maintainable internal wrapper before any production integration;
- the spike contains synthetic information only.

## Next step

Create the isolated native Win32 hybrid visual prototype against the locked synthetic presentation fixture. Capture and review the complete representative screen set before any renderer is integrated into a new ECO application version.

See `P2_PRESENTATION_RENDERER_BAKEOFF_DECISION_2026-08-08.md` for the controlling decision and mandatory gates.
