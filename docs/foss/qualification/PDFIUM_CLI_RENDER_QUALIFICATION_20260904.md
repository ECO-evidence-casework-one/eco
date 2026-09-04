# pdfium-cli Windows renderer qualification

Date: 2026-09-04

Status: PASS for the pinned WebAssembly Windows renderer on a GitHub-hosted Windows runner. This is not low-spec Acer evidence and does not approve release.

- Upstream: klippa-app/pdfium-cli
- Licence: MIT; upstream states embedded Wazero and PDFium are Apache-2.0.
- Release: v0.11.2
- Exact source tag commit: 260c846dbbd180fdc478a2771e9dae9914164846
- WebAssembly asset: pdfium-webassembly-windows-amd64
- WebAssembly bytes: 16988160
- WebAssembly SHA-256: b56c3c405111ae68cc99b225f8627ea25ec5a7cb3188bdfca67b4cac5df2189f
- Mean one-page 1200px render time: 2695 ms
- Slowest of three WebAssembly renders: 2829 ms
- Maximum live working set observed: 286.9 MiB
- Maximum rendered PNG bytes: 19536

- Native comparison asset: pdfium-native-windows-amd64
- Native asset bytes: 31034735
- Native asset SHA-256: 93033804cce3cc4e5fc425abaec3d3ce9f37736d3600dbeb1e5f1cc676816a39
- Native clean-runner result: failed-or-unavailable
- Native diagnostic: 

Decision: prefer the single-file WebAssembly release for ECO adapter work unless low-spec testing proves it unusably slow. The native route is not preferred if it requires external runtime libraries.
