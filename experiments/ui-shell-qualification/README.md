# ECO UI shell qualification — Fyne / Wails / Avalonia

Status: isolated synthetic qualification only. This branch is **not** ECO product source and must not be merged into the historical `main` as a product basis.

## Why this lane exists

The current authoritative ECO source is the sealed 3 September 2026 Slice 10 GitHub-first P0-hardening checkpoint outside public `main`. This lane exists only to answer the presentation-framework question using current GitHub upstreams before any shell migration is attempted.

## GitHub-first candidates

- **Fyne v2.8.1** — BSD-3-Clause. Current stable tag commit `3dc06f47137aa3709807f1bb78dacc5daf68d551`. Fully open desktop toolkit, but its current Windows UIA provider exposes roles/names/focus without UIA control-pattern providers. It therefore requires an empirical ValuePattern/text-input gate before ECO can rely on it.
- **Wails v2.14.0** — MIT. Already proven to compile the React/TanStack/Radix/XYFlow/vis-timeline casework stack on Windows. Windows uses Microsoft Edge WebView2 Runtime, so WebView2 remains an external runtime/FOSS-boundary dependency even if the UIA behaviour passes.
- **Avalonia v12.1.2** — MIT. Current release published 2 September 2026. Its automation-peer model includes `TextBoxAutomationPeer : IValueProvider`, and the Windows automation bridge maps Avalonia `IValueProvider` to the Windows Value pattern. Self-contained .NET is eligible for qualification because the .NET runtime is MIT.
- **Qt 6/QML** — existing baseline only. ECO PR #85 already proved a clean-Windows package with 65 UIA descendants, 20 buttons, 52 keyboard-focusable elements, incremental search and accessible/selectable transcript. It is not rebuilt in this lane unless all three new candidates fail.

## Gate

Every candidate gets the same synthetic Matter surface:

1. searchable Matter field;
2. incremental `war` filtering without Enter/Search button;
3. five or more keyboard-reachable actions;
4. transcript content exposed through Windows accessibility, not painted pixels only;
5. process/window launch proof with developer paths removed;
6. package size and runtime-dependency inventory;
7. no real evidence, model, network service or ECO backend.

The build phase may use GitHub-hosted toolchains. The consumer probe removes developer/build paths before launch. No runnable executable is uploaded as a workflow artifact; only text diagnostics may be uploaded if later required.

## Decision rule

A compile-only pass is insufficient. ECO requires usable keyboard and Windows accessibility behaviour. A framework that fails the search ValuePattern or transcript-accessibility gate is `HOLD` even if it is otherwise fully FOSS. A framework that passes accessibility but requires a non-FOSS runtime remains `HOLD` against ECO's strict all-FOSS target unless that boundary is explicitly changed later.
