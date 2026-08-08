# P2 Presentation Renderer Bake-off Decision

**Date:** 2026-08-08  
**Branch:** `dev/p2-presentation-bakeoff-20260808`  
**Status:** Decision record for isolated development only. Not a release approval, not a merge approval and not authority to alter the current user-test application.

## Decision

None of the original three renderer candidates is acceptable **as-is** for ECO's presentation layer because Windows accessibility is a mandatory product requirement, not a later enhancement.

A fourth architecture — a **pure-Go native Win32 hybrid** using real standard Windows controls for semantic and interactive elements — is promoted to **lead architecture for the next isolated P2 presentation spike**.

This is **not** a final production selection. It must still pass physical Windows screen-reader, keyboard, scaling, high-contrast, visual-fidelity and maintainability gates before it may be proposed for the next application version.

The current application source and user-test candidate remain untouched by this decision.

## Why accessibility is a hard gate

ECO is intended to be usable by people who may be blind, partially sighted, neurodivergent, elderly, cognitively overloaded or otherwise reliant on keyboard navigation, clear focus behaviour, screen readers, large text and operating-system accessibility features.

A renderer that can draw attractive pixels but cannot expose reliable semantics and interaction contracts to Windows accessibility technology is not acceptable.

## Bake-off evidence

| Candidate | Compile / render | Windows build | Windows accessibility | Decision |
|---|---|---|---|---|
| Gio v0.10.1 | PASS | PASS, including no-CGo Windows cross-build | FAIL as-is: Windows source audit found no UI Automation or MSAA bridge | Do not promote |
| Fyne v2.8.0 | PASS | PASS; accessibility-tag build also compiles | FAIL as-is for ECO: current accessible roles are too limited for the required edit-heavy workflow; `Entry` is not exposed through the current accessibility interface and the Windows provider supplies no UIA interaction patterns | Do not promote |
| Cogent Core v0.3.38 | Reserve only | Not needed for visual bake-off after accessibility audit | FAIL as-is: source audit across 1,219 files found no Windows screen-reader bridge; upstream accessibility work remains unresolved | Do not promote |
| Native Win32 hybrid spike | Architecture spike, not visual parity candidate | PASS: one-file amd64 GUI EXE, `CGO_ENABLED=0` | PASS for direct MSAA contracts on standard controls; hosted Windows Server UIA mapping remains inconclusive | **Lead architecture for next isolated spike only** |

## Native Win32 proof

Green workflow:

- workflow: `P2 native Win32 accessibility spike`
- run: `31267931187`
- tested commit: `6bfa0b08fad38b64ab91657e95767fbb972176e2`
- runner: Microsoft Windows Server 2025, build 10.0.26100
- Go: 1.23.12 windows/amd64

Build proof:

- standard-library-only Go module: PASS
- unexpected network-capable API / URL scan: PASS
- `go vet ./...`: PASS
- `CGO_ENABLED=0`: PASS
- Windows GUI build: PASS
- executable size: **2,305,536 bytes**
- SHA-256 for this CI build: `bd48f60f43306265f6c2fb4b5b5e703283120600a77f3ffce18f61654abf06ae`
- public runnable artifact upload: ABSENT

The executable was created only inside the ephemeral CI runner and was not published or committed.

### Direct Microsoft Active Accessibility proof

The CI probe queried the actual standard control HWNDs through `oleacc.dll` / `IAccessible` rather than inferring accessibility from visual output.

Observed contracts:

| Control | Accessible name | MSAA role | Focusable | Default action |
|---|---|---:|---|---|
| `STATIC` | Matter Workspace | 41 — static text | not required | n/a |
| `BUTTON` | Workspace | 43 — push button | yes | Press |
| `BUTTON` | Documents | 43 — push button | yes | Press |
| `BUTTON` | Review evidence | 43 — push button | yes | Press |
| `BUTTON` | Ask ECO | 43 — push button | yes | Press |
| `EDIT` | Search this Matter | 42 — text/edit | yes | n/a |
| `LISTBOX` | standard list object | 33 — list | yes | n/a |

The probe also invoked the Review evidence button through the accessibility default-action contract successfully.

Green gates recorded by CI:

- `native_win32_msaa_roles=PASS_STATIC_BUTTON_EDIT_LIST`
- `native_win32_msaa_names=PASS_BUTTON_NAMES`
- `native_win32_msaa_focusability=PASS_BUTTON_EDIT_LIST`
- `native_win32_msaa_default_action=PASS_REVIEW_BUTTON`

## UI Automation limitation observed in hosted CI

The Windows Server 2025 GitHub-hosted runner exposes the real standard-control HWNDs, class names and text, but the managed UI Automation client used in that non-interactive session flattens the controls to `ControlType.Pane` rather than applying the normal standard-control proxy mapping.

This is recorded as:

`native_win32_hosted_uia=INCONCLUSIVE_STANDARD_CONTROLS_FLATTENED_TO_PANES`

It is **not** counted as a UIA pass and it is **not** counted as evidence that the controls lack accessibility. Direct MSAA contracts passed on those same HWNDs.

Production promotion therefore requires an interactive physical-Windows validation using the actual accessibility tools a user would rely on.

## Mandatory next gates

The native Win32 hybrid may advance only through the following sequence.

### Gate N1 — physical Windows accessibility

Run on a normal interactive Windows 11 machine, preferably the target low-spec Acer as well as at least one independent Windows machine.

Mandatory checks:

- NVDA reads window title, page title, navigation, buttons, edit controls, document lists, status changes and error messages correctly.
- Windows Narrator performs the same core traversal and interaction.
- keyboard-only operation reaches every interactive function in a logical order.
- focus is always visible and never becomes trapped or lost.
- text entry can be read, edited, selected and corrected with a screen reader.
- list selection, list navigation and button activation work without a mouse.
- dynamic status changes are announced where appropriate without becoming noisy.
- no decorative element is exposed as misleading interactive content.

Failure of a core control is a stop condition.

### Gate N2 — Windows display and accessibility settings

Mandatory checks at minimum:

- 100%, 125%, 150% and 200% display scaling;
- Windows text-size enlargement;
- Windows high-contrast / contrast themes;
- keyboard focus indicators;
- light/dark or ECO-supported appearance behaviour;
- window resizing down to the supported minimum size;
- no clipped controls, unreadable text or inaccessible horizontal-only navigation.

### Gate N3 — visual-fidelity prototype

The current native spike proves architecture and accessibility only. It does **not** establish final visual quality.

Before the next real ECO version is built:

1. reproduce every locked representative ECO screen in the native hybrid;
2. capture the resulting screens;
3. compare them against the approved visual specification;
4. fix discrepancies before application integration;
5. show the complete screen set before authorising the application build.

No new ECO version should surprise the tester visually.

### Gate N4 — hybrid rendering rule

To preserve accessibility while allowing a modern interface:

- semantic and interactive controls remain real native controls or have a proven equivalent native accessibility provider;
- decorative backgrounds, cards, separators, icons and non-semantic visual effects may be custom painted;
- custom drawing must never silently replace the accessible edit, button, list, tree, tab, checkbox, radio, menu or other interaction contract;
- owner/custom drawing is permitted only when the underlying native control semantics remain intact and are re-tested;
- keyboard focus and accessibility state are part of the component contract, not optional styling details.

### Gate N5 — maintainability and resource use

Before integration, measure and record:

- cold start;
- idle memory;
- active memory during representative navigation;
- CPU use while idle and while resizing/scrolling;
- executable size;
- resize responsiveness;
- rendering defects and flicker;
- code complexity of the Win32 wrapper layer;
- Windows-version compatibility boundaries.

A small, explicit internal wrapper should be preferred over scattering raw Win32 calls throughout ECO.

## Dependency and platform position

The architecture spike uses only the Go standard library plus Windows system DLLs already provided by the target operating system. It adds no cloud service, browser runtime, telemetry service, proprietary ECO dependency or third-party Go module.

Windows itself remains the unavoidable proprietary target platform. That is distinct from introducing proprietary components into ECO's own bundled stack.

If a later implementation uses an additional Go package such as `golang.org/x/sys/windows`, it must be independently licence-checked, pinned and justified before adoption; this decision does not pre-approve it.

## What this decision does not authorise

It does not authorise:

- merging this branch to `main`;
- replacing the current ECO renderer;
- creating a user release;
- claiming full screen-reader compliance;
- claiming visual parity with the locked ECO design;
- publishing the spike EXE;
- introducing real evidence or personal data into the spike;
- removing the preserved Gio/Fyne/Cogent evidence.

## Next development action

Build the **native Win32 hybrid visual/accessibility prototype** as an isolated P2 continuation, using the existing locked synthetic presentation fixture and the approved ECO visual specification. The prototype must keep native semantic controls while bringing the appearance up to the locked design.

Only after the complete visual set and accessibility behaviour are demonstrated should integration into the next application version be considered.
