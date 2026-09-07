# ECO Product Lab — GitHub-first Windows UI Automation reconnaissance

**Date:** 7 September 2026  
**Application baseline:** `5623de397fed16185348095b19fe692afe00e5d5`  
**Purpose:** avoid duplicating accessible-Win32/provider engineering that already exists in free/open-source code.

## Mission boundary

This record does **not** authorise an application-source merge before the real Qwen rig acceptance gate. It identifies donor/reference material so the first accessibility implementation slice can start from proven work after that gate.

## 1. Microsoft Terminal — strong production reference

**Repository:** `microsoft/terminal`  
**Licence:** MIT  
**Decision:** **REFERENCE / bounded donor candidate**

Current `HwndTerminalWndProc` contains the exact Win32 root-provider handshake relevant to ECO:

- receives `WM_GETOBJECT`;
- checks `UiaRootObjectId`;
- returns the provider with `UiaReturnRawElementProvider(hwnd, wParam, lParam, provider)`.

Use this as a production-quality reference for the window-procedure boundary. Do not import Terminal's wider framework/dependency architecture.

## 2. Microsoft Windows SDK API documentation — authoritative contract

**Repository:** `MicrosoftDocs/sdk-api`  
**Decision:** **AUTHORITATIVE PLATFORM CONTRACT**

The GitHub-hosted SDK documentation establishes important implementation details:

- `UiaReturnRawElementProvider` is the window/control response path for `WM_GETOBJECT`;
- the `wParam`/`lParam` values should be passed through rather than pre-filtered in a way that can break MSAA clients;
- windows that have returned providers should notify UIA on destruction so provider-map references can be released.

ECO tests must encode these lifecycle/boundary rules if an app-defined provider is implemented.

## 3. Flutter Windows accessibility architecture — strong conceptual reference

**Repository:** `flutter/flutter`  
**Licence:** BSD-style 3-clause  
**Decision:** **REFERENCE ONLY — do not import Flutter**

The Windows accessibility documentation describes a complete custom accessibility tree model:

- `IRawElementProviderFragment` per accessibility element;
- a window-level `IRawElementProviderFragmentRoot`;
- screen bounds and hit testing;
- current-focus lookup;
- properties such as name/role;
- optional UIA patterns such as text/toggle support;
- `WM_GETOBJECT` + `UiaRootObjectId` → `UiaReturnRawElementProvider`.

This establishes what ECO would need to expose for any genuinely custom rendered region.

## 4. `doug/gophics` — high-value pure-Go donor candidate

**Licence:** Apache-2.0  
**Decision:** **QUALIFY AS DONOR CANDIDATE; do not copy wholesale yet**

The Windows UIA implementation is unusually relevant to ECO because it is pure Go and already uses the same broad Win32/syscall style.

Observed engineering includes:

- pure-Go COM provider construction without CGo;
- real multi-interface COM layout and `QueryInterface` handling;
- explicit reference counting and live-object retention so UIA-held pointers remain valid to the Go GC;
- `IRawElementProviderSimple`, `IRawElementProviderFragment`, fragment-root and invoke/toggle interfaces;
- proper UIA VARIANT/BSTR/SAFEARRAY concerns;
- provider bounds in screen coordinates;
- `UiaHostProviderFromHwnd` / host-provider fallback concepts;
- event/disconnect functions;
- debug logging for the otherwise invisible provider handshake.

This can save substantial duplicated engineering if qualified carefully.

### Required adoption checks

Before taking any code into ECO:

1. identify the minimum self-contained provider/lifetime pieces actually required;
2. inspect all referenced files and dependencies, not only `uia_windows.go`;
3. verify Go/Windows architecture assumptions against ECO's current syscall/memory rules;
4. preserve Apache-2.0 attribution/notice requirements for any copied substantial portions;
5. build focused COM lifetime/refcount tests;
6. run provider behaviour through Accessibility Insights + Narrator + NVDA on real Windows;
7. red-team GC/lifetime/reentrancy/window-destroy paths;
8. prefer native HWND conversion where a custom provider is unnecessary.

## 5. `odvcencio/FluffyUI` — useful mapping reference, reject direct adoption now

**Licence:** MIT  
**Decision:** **REFERENCE ONLY / NOT QUALIFIED FOR DIRECT ADOPTION**

Useful material exists for:

- UIA DLL/procedure names;
- control/property/event identifiers;
- Go vtable/callback shape;
- an accessible-tree abstraction.

However, inspected portions explicitly describe property/event code as simplified; shown provider methods include minimal `QueryInterface`/reference-count behaviour and incomplete pattern/property handling. That is not strong enough evidence for direct adoption into an evidence application.

ECO may use it as a cross-check for symbols/structure, not as the provider implementation baseline.

## Engineering decision

### Prefer native controls for ordinary interaction

For ECO's seven primary navigation items, ordinary buttons, search controls and simple action rows, **real focusable HWND controls remain the preferred solution**. They reduce custom COM code, improve default keyboard behaviour and provide Windows accessibility proxies.

### Reserve a custom UIA provider for genuinely custom surfaces

A provider is justified only where the interaction genuinely depends on custom rendering — for example document/image/evidence regions whose semantics cannot be represented safely by ordinary controls.

If needed, the implementation order should be:

1. Microsoft SDK contract;
2. Microsoft Terminal production boundary;
3. qualified minimal pure-Go donor concepts from `doug/gophics`;
4. ECO-specific tests/lifetime safeguards;
5. real assistive-technology acceptance.

## Current conclusion

GitHub already contains enough free/open-source/reference material that ECO should **not invent the Windows UI Automation COM/provider layer from scratch**. But the evidence also supports **not building a custom provider where native controls are sufficient**.
