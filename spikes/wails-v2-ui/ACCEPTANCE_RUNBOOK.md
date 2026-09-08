# ECO Wails v2 Windows Acceptance Runbook

Status: human acceptance procedure for the non-production Wails v2 architecture spike.

This is **not** a production ECO build and must use **synthetic test workspaces only**.

## Before running

1. Download the GitHub Actions artifact named `eco-wails-v2-windows-acceptance-<commit>` from the matching workflow run.
2. Extract it to a disposable local folder.
3. Verify `eco-ui-spike.exe` against `eco-ui-spike.sha256.txt` with PowerShell:

```powershell
(Get-FileHash .\eco-ui-spike.exe -Algorithm SHA256).Hash.ToLowerInvariant()
Get-Content .\eco-ui-spike.sha256.txt
```

The hashes must match exactly before the EXE is used.

4. Read `BUILD_INFO.txt` and confirm the commit/run matches the acceptance receipt.
5. Keep all test workspaces disposable and synthetic. Do not open a real ECO evidence workspace with this spike.

## Core journey

Run these actions without relying on the existing Win32 application:

1. Start the spike.
2. Create a new synthetic workspace in a new, non-existing folder.
3. Confirm the workspace summary appears.
4. Create a synthetic matter.
5. Refresh the summary and confirm the matter appears.
6. Close the workspace.
7. Open the same workspace again.
8. Confirm the matter survives reopen.
9. Close the application normally.
10. Reopen it and prove the workspace route is no longer locked by the previous process.

## Keyboard-only acceptance

Repeat the core journey without using a mouse after launch.

Record PASS/FAIL for:
- all actionable controls reachable by keyboard;
- logical focus order;
- visible focus at every step;
- Enter/Space activation where appropriate;
- no keyboard trap;
- focus returns to a sensible location after errors/actions;
- closing/reopening does not strand focus or ownership state.

## Narrator and NVDA

Run the same core journey with Windows Narrator and separately with a stable NVDA release.

Record whether controls expose meaningful:
- accessible name;
- role;
- state/value;
- error/status changes;
- focus movement;
- grouping/reading order.

Do not record a PASS merely because speech occurs. The spoken information must identify the control and its state sufficiently to perform the task.

## Display scaling and layout

Run at:
- 100%
- 125%
- 150%
- 175%
- 200%

At every scale verify:
- no clipped required text;
- no overlapping controls;
- buttons remain operable;
- focus indicator remains visible;
- content can still be reached at the minimum supported window size;
- no horizontal/vertical region becomes permanently unreachable.

## High contrast and colour independence

Enable a Windows high-contrast theme and repeat the core journey.

Verify:
- text remains readable;
- focus is visible;
- borders/control states remain distinguishable;
- success/error state is not colour-only;
- disabled/selected/focused states remain understandable.

## WebView2 behaviour

Record the installed WebView2 runtime version.

Test normal launch. If a safe disposable machine/test environment allows it, also verify failure/missing-runtime behaviour without damaging the primary workstation. Failure must be explicit and understandable rather than a silent exit.

## Automated assistance

Use the qualified external stack where available:
- FlaUI for deterministic UI driving;
- Axe.Windows for automated UIA rules;
- Accessibility Insights for Windows for manual UIA inspection;
- NVDA for real screen-reader behaviour.

Automated zero-error output is not an accessibility conformance claim and cannot replace the manual checks above.

## Receipt

Copy `WINDOWS_ACCEPTANCE_RECEIPT_TEMPLATE.json` to a new receipt file and fill in only observed facts.

Rules:
- unperformed checks remain `NOT_RUN`;
- failures remain `FAIL` until genuinely corrected and rerun;
- include exact ECO commit/build, Windows version, display scale, WebView2 version, NVDA version and local report hashes where generated;
- do not put private evidence, private filenames or sensitive screenshots in the receipt.

## Promotion decision

The spike may pass this gate only when all mandatory journeys have evidence and no material accessibility/lifecycle blocker remains.

Passing this gate permits engineering work to begin on the production Startup / Workspace Chooser. It does **not** retire the existing Win32 UI and does **not** establish WCAG/EN 301 549 conformance.
