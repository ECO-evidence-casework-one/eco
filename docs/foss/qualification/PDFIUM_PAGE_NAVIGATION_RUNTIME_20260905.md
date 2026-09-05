# PDFium multi-page navigation runtime qualification

Date: 2026-09-05

Status: PASS on GitHub-hosted Windows. This is not the controlling 8 GB Acer performance qualification.

- Runtime: klippa-app/pdfium-cli v0.11.2 Windows WebAssembly
- SHA-256: b56c3c405111ae68cc99b225f8627ea25ec5a7cb3188bdfca67b4cac5df2189f
- ECO opened a fresh vault and preserved a deterministic three-page PDF.
- ECO registered the exact runtime through the normal verified local-tool registry.
- PDFEvidenceInfoWithRegisteredPDFium reported exactly 3 pages and remained bound to the preserved object/SHA-256.
- ECO rendered pages 1, 2 and 3 through the registered preserved-source workflow; deliberately different page shapes produced different dimensions.
- Page 4 was rejected.
- Actions run: 33947946928
- No renderer binary or rendered page was uploaded as a repository artifact.
