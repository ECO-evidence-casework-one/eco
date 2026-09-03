# ECO GitHub-First FOSS Intake Receipt — 3 September 2026

Status: **PASS — first component set proven compatible on Linux and Windows**

Branch: `integration/github-first-stack-20260903`

## GitHub Actions evidence

Workflow: `GitHub-first FOSS intake`

Run ID: `33723102495`

Jobs:

- Linux: `probe-linux` — PASS
- Windows: `probe-windows` — PASS

The Windows job ran on Microsoft Windows Server 2025 using Go `1.25.14 windows/amd64`.

`go mod verify` reported:

> all modules verified

The combined test then passed these real integration checks:

1. Wails v2 shell API/options load correctly.
2. Bleve creates and queries a **memory-only** Matter search index.
3. `mholt/archives` identifies an actual generated ZIP as an extractable container.
4. `emersion/go-message` parses an RFC/MIME email and decodes its subject.
5. `emersion/go-mbox` reads a two-message MBOX container.
6. Excelize writes and reopens an actual XLSX workbook/cell.
7. `gabriel-vasile/mimetype` performs content-based MIME classification.

Result: all seven checks PASS on Windows and the same combined probe PASS on Linux.

## Exact direct component versions proven together

- `github.com/wailsapp/wails/v2` `v2.14.0`
- `github.com/blevesearch/bleve/v2` `v2.6.1`
- `github.com/mholt/archives` `v0.1.5`
- `github.com/emersion/go-message` `v0.18.2`
- `github.com/emersion/go-mbox` `v1.0.4`
- `github.com/xuri/excelize/v2` `v2.11.0`
- `github.com/gabriel-vasile/mimetype` `v1.4.15`

All transitive Go modules resolved through the Go module checksum system and were accepted by `go mod verify`.

## Security/privacy interpretation

This receipt proves compatibility only. It does **not** authorise a public release or merge to `main`.

The accepted architectural boundaries remain:

- Bleve is memory-only first, preventing a new plaintext persistent case-index copy.
- Archive traversal will be wrapped in ECO size/depth/path/cancellation limits.
- Parsed email, spreadsheet and MIME data are derived records and never replace preserved source bytes.
- Wails is the presentation-shell candidate; ECO's provenance, Vault, confirmation and recovery logic remain authoritative.

## Next proof

Build a real Wails v2 React+TypeScript Windows desktop executable from the exact v2.14.0 CLI/template in GitHub Actions, without publishing the executable artifact. Then prove the chosen accessible frontend primitives and data/relationship components in that shell before migrating ECO screens.
