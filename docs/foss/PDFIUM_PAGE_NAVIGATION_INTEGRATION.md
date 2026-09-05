# PDFium multi-page preview navigation

Date: 2026-09-05

## Decision

Extend the already-qualified optional `klippa-app/pdfium-cli` v0.11.2 renderer; add no new donor. ECO uses the renderer's existing `info --output-type json` surface to obtain a bounded page count from the same verified preserved PDF source.

## Behaviour

- Initial PDF preview requests page metadata once, bound to the preserved object SHA-256.
- The current page is shown as `page X of Y` when metadata is available.
- Left/Right Arrow and PageUp/PageDown request the previous/next page without closing the preview window.
- Page changes are rendered asynchronously so the native preview window remains responsive.
- While a page is rendering, duplicate navigation requests are ignored and the preview reports the target page.
- Each new page still goes through `RenderEvidencePDFPageWithRegisteredPDFiumContext`, so the registered runtime is re-verified and the preserved source is freshly verified before rendering.
- Page navigation clears any citation-region highlight when moving away from the cited page.
- If page-count metadata is unavailable, the existing page-1 / exact-citation-page preview remains usable; navigation is simply disabled.

## Safety

The existing renderer limits remain unchanged: no runtime download, exact v0.11.2 SHA/size identity, 45-second render deadline, source <=512 MiB, bounded diagnostics, width/height/pixel/PNG limits and temporary derivative cleanup. Metadata stdout is additionally capped at 8 MiB and page count is constrained to ECO's existing safe render range.

## Qualification

Source qualification run `33947757143` passed all Go tests, `go vet`, Windows cross-build, offline/source policy and diff checks before the product commit.

Exact-runtime Windows qualification run `33947946928` independently downloaded and SHA-256-verified the pinned v0.11.2 WebAssembly executable, opened a fresh ECO vault, preserved a deterministic three-page PDF, registered the runtime through ECO, proved `PDFEvidenceInfoWithRegisteredPDFium` returned exactly 3 source-bound pages, rendered pages 1, 2 and 3 through the registered preserved-source path with deliberately different page dimensions, and rejected page 4. No renderer binary or rendered page was uploaded as a repository artifact.

The remaining software gate is the normal Linux/Windows/source-policy/Gitleaks/Syft/Cosign PR pipeline. The 8 GB Acer remains the controlling low-spec hardware gate for performance claims.
