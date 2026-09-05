# pdfium-cli isolated PDF page renderer integration

Date: 2026-09-04

## Decision

USE/ADAPT as an isolated optional runtime. ECO does not link `go-pdfium` or PDFium/CGO into the application. The exact `klippa-app/pdfium-cli` WebAssembly Windows release `v0.11.2` is used as a caller-located executable because it is a single file and does not require a separately installed PDFium DLL stack.

## Exact qualified identity

- Upstream: `klippa-app/pdfium-cli`
- Source tag commit: `260c846dbbd180fdc478a2771e9dae9914164846`
- Wrapper licence: MIT; upstream states bundled Wazero and PDFium are Apache-2.0.
- Asset: `pdfium-webassembly-windows-amd64`
- Bytes: `16988160`
- SHA-256: `b56c3c405111ae68cc99b225f8627ea25ec5a7cb3188bdfca67b4cac5df2189f`

GitHub-hosted Windows qualification rendered the same known one-page PDF repeatedly at a 1200-pixel width in roughly 2.7 seconds, with a maximum observed live working set of about 286.9 MiB. The native comparison executable did not render successfully on the clean runner without its external runtime stack. These measurements are development evidence only; they are not the controlling low-spec Acer qualification.

## ECO safety boundary

The runtime registry checks the exact executable identity and self-checks its `render`/`info` command surface. The executable is re-verified immediately before use. ECO then materialises a freshly verified temporary reading copy of one preserved PDF, renders one explicit page with a 45-second default deadline, limits requested width to 320–2000 pixels and height to 3000 pixels, limits the source reading copy to 512 MiB, limits the PNG to 32 MiB and six million pixels, validates the PNG header/dimensions, and deletes the temporary workspace. No rendered page is persisted as evidence and no renderer download occurs inside ECO.

Ordinary PDF preview renders page 1. A source-backed citation with a recorded `Citation.Page` requests that exact page. The existing image-preview window is reused for viewing the rendered derivative; PDF-specific preview rotation is view-only and is not persisted as evidence rotation.

## Actual ECO runtime qualification

Windows workflow run `33913936762` exercised the exact pinned release through ECO rather than calling the renderer in isolation. It independently rechecked the asset size/SHA-256, opened a fresh ECO vault, imported and committed a synthetic two-page PDF through the normal preservation/verification pipeline, registered the renderer through the local-runtime registry, rendered page 1 and page 2 through `RenderEvidencePDFPageWithRegisteredPDFium`, confirmed both rendered derivatives remained bound to the preserved source object/SHA-256 and that the different-size requested pages produced different rendered dimensions, then deliberately modified the registered executable. ECO refused the next render after that modification. The workflow also removed the downloaded executable and uploaded no runnable artifact. Result: **PASS**.

## Remaining gate

Before this renderer can be described as suitable for the controlling low-spec Windows target, the exact runtime plus ECO adapter must be exercised on that real machine. This integration does not waive accessibility, clean-machine, signing, publisher or release gates.
