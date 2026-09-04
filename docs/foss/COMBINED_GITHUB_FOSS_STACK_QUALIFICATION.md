# Combined GitHub FOSS stack qualification

Date: 2026-09-04

This staging-only qualification combines the four independently green draft slices without merging them to `main`:

- native PDF reader: `integration/native-pdf-reader-20260904` / PR #114;
- verified Tesseract runtime bundle: `integration/tesseract-runtime-bundle-20260904` / PR #115;
- bounded MBOX reader: `integration/mbox-reader-20260904` / PR #116;
- hostile-input fuzz baseline: `integration/hostile-input-fuzz-baseline-20260904` / PR #117.

The PDF and MBOX branches both modify `internal/eco/extract.go` and `THIRD_PARTY_NOTICES.md`; staging resolves those overlaps by retaining both PDF and MBOX imports, extraction cases, implementations and notices. No constituent draft PR or `main` is rewritten by this staging proof.
