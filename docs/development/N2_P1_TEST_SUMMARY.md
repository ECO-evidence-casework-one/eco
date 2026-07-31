# Version 25 N2 P1 test summary

**Build ID:** `ECO-V25-20260731-N2-P1`  
**Test date:** 31 July 2026  
**Release class:** source development snapshot; no unsigned end-user executable distributed

## Automated results

- `go test -count=1 ./...`: PASS
- `go vet ./...`: PASS
- offline source and embedded-secret policy scan: PASS
- targeted race-detector run across N2 vision, OCR, vault rollback and retrieval tests: PASS
- Windows x86-64 GUI cross-build: PASS
- deterministic two-build comparison: byte-for-byte PASS
- Windows artifact format: PE32+ x86-64 GUI
- imported runtime DLLs: `kernel32.dll` only

## Deterministic verification artifact

The controlled verification artifact was built only to validate the source pipeline and is deliberately excluded from the public upload package.

- Size: 3,824,640 bytes
- SHA-256: `f1ddb582b307bf4dcdbd606deb5032468fa73a5d648467ffc6b4ce2e9e34c14d`
- Signed: no
- Smart App Control end-user status: unsuitable until trusted signing is operational

## N2-specific tests

- conservative photographed-page boundary suggestion;
- non-destructive crop output;
- skew estimation and correction direction;
- adaptive document thresholding;
- perspective transform and rejection of invalid quadrilaterals;
- coordinate-bearing OCR TSV parsing;
- OCR source-hash binding;
- OCR line/segment correspondence;
- rejection of invalid nested OCR words;
- low-confidence OCR exclusion from retrieval;
- preservation of the encrypted original after OCR application;
- in-memory rollback when authenticated workspace persistence is forced to fail;
- bounded preview processing for large images;
- static native-interface controls for crop, deskew, adaptive view and exact citation-region highlighting.

## Execution limitation

The Windows GUI was cross-built but not executed on a Windows runner or physical Windows test machine in this source milestone. Direct Windows UI automation, high-DPI screenshots, Narrator/NVDA and long-duration stability evidence remain release blockers.
