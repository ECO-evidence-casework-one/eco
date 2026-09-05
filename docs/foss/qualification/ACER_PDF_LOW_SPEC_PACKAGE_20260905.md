# Acer PDF low-spec one-click qualification package

Date: 2026-09-05

Status: package built and self-tested PASS on GitHub-hosted Windows; real Acer result remains required.

- Combined ECO source commit under test: `85c670e2b3a5fc5a3d87d814e896d749d8cd732b`
- Qualifier: `ECO-PDF-LOW-SPEC-20260905.1`
- Exact bundled PDFium: `klippa-app/pdfium-cli` v0.11.2 Windows WebAssembly
- PDFium SHA-256: `b56c3c405111ae68cc99b225f8627ea25ec5a7cb3188bdfca67b4cac5df2189f`
- Artifact name: `ECO_Acer_PDF_LowSpec_Qualification_20260905`
- Artifact ID: `9964235966`
- GitHub artifact digest: `27c15254ca87517f1927aa2103cb430529ab043931f27de6d68af0cafbd868d5`
- Packaged qualifier SHA-256: `8c694aee9727b67fd7656cd28bf5c52ab36227990e7a31d0dfe0c1861fa6f220`
- Packaged runtime SHA-256 recheck: `b56c3c405111ae68cc99b225f8627ea25ec5a7cb3188bdfca67b4cac5df2189f`
- The package is fully offline at runtime and contains no installer.
- The user action is extract once and double-click `RUN_ECO_PDF_LOW_SPEC_QUALIFICATION.cmd`.
- Result files: `ECO_PDF_LOW_SPEC_RESULT.txt` and `.json`.

## Package manifest
- 8c694aee9727b67fd7656cd28bf5c52ab36227990e7a31d0dfe0c1861fa6f220  ECO_PDF_LOW_SPEC_QUALIFIER.exe  4347392 bytes
- b56c3c405111ae68cc99b225f8627ea25ec5a7cb3188bdfca67b4cac5df2189f  pdfium-webassembly-windows-amd64.exe  16988160 bytes
- af20ca81582a0b5d22de81ea7ea1295f1c84e34cb373bd9d7cc7c727c9f46d8d  RUN_ECO_PDF_LOW_SPEC_QUALIFICATION.cmd  85 bytes
- 4b0ca8b6a92ed337af9da1cb063f3311be5f942408808d065cd1146e214f909a  START_HERE.txt  1160 bytes
