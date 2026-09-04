# pdfium-cli Windows WebAssembly render qualification

Date: 2026-09-04

Status: PASS — GitHub-hosted Windows runner qualification only. This is not low-spec Acer evidence and does not approve integration or release.

- Upstream: klippa-app/pdfium-cli
- Licence: MIT; upstream states embedded Wazero and PDFium are Apache-2.0.
- Release: v0.11.2
- Exact source tag commit: 260c846dbbd180fdc478a2771e9dae9914164846
- Asset: pdfium-webassembly-windows-amd64
- Asset bytes: 16988160
- Asset SHA-256: b56c3c405111ae68cc99b225f8627ea25ec5a7cb3188bdfca67b4cac5df2189f
- Packaging: single Windows PE; no separately installed PDFium runtime required for the WebAssembly release type.
- Known one-page PDF: rendered to PNG successfully three times at max width 1200 px.
- Mean render process time on GitHub runner: 2383.33 ms
- Slowest of three runs: 2415 ms
- Maximum observed process peak working set: 0 MiB
- Maximum rendered PNG bytes: 19536
- Malformed PDF: non-zero exit within 15 seconds; no output retained.
- No ECO executable or pdfium-cli binary was uploaded as an Actions artifact.

Decision boundary: suitable for further ECO adapter qualification as an isolated, hash-pinned renderer. Low-spec Windows evidence remains required before adoption.
