# Combined GitHub FOSS stack qualification

Date: 2026-09-05

This staging-only qualification combines five independently green draft slices without merging them to `main`:

- native PDF reader: `integration/native-pdf-reader-20260904` / PR #114;
- verified Tesseract runtime bundle: `integration/tesseract-runtime-bundle-20260904` / PR #115;
- bounded MBOX reader: `integration/mbox-reader-20260904` / PR #116;
- hostile-input fuzz baseline: `integration/hostile-input-fuzz-baseline-20260904` / PR #117;
- qualified optional PDF page renderer: `integration/pdfium-page-renderer-20260904` / PR #119.

The first four slices were already collision-qualified together by draft PR #118. The PDF renderer was then qualified independently: its exact `klippa-app/pdfium-cli` v0.11.2 Windows WebAssembly runtime passed end-to-end ECO vault registration, preserved-source page rendering, page-specific output and deliberate executable-tamper rejection in Actions run `33913936762`; PR #119 then passed the normal Linux, Gitleaks, source-policy, deterministic Windows build, Syft and Cosign pipeline in run `33914224993`.

Adding PR #119 to the already-green #118 stack produced exactly two merge conflicts: `THIRD_PARTY_NOTICES.md` and `cmd/eco/main_windows.go`. `internal/eco/local_tools.go` merged automatically. The staging resolution retained the native PDF and MBOX notices, appended the exact optional PDFium-runtime notice, preserved Tesseract registration/completion handling, placed Tesseract and PDF renderer controls together on the Trust & Settings runtime row, and retained the renderer's citation-page and bounded visual-preview path. No other file received manual conflict treatment.

After that controlled resolution, run `33947021567` passed conflict-set enforcement, `go mod tidy`, all Go tests, `go vet`, Windows cross-build, offline/source policy and diff checks. Temporary resolver/workflow files were then removed and the inherited `scripts/foss/ACQUIRE_DONORS.cmd` CRLF bytes were restored exactly to the `main` blob (`7c272f0f96b9c8c27cc3902b2e0eeb85c7b74370`) so no line-ending churn remains.

This branch is still staging-only. It does not supersede the constituent draft PRs and does not authorize a merge to `main`. The next control is the ordinary full ECO PR pipeline against this clean five-slice combined branch. The separate low-spec Windows/Acer runtime qualification for PDF visual rendering also remains required before claiming suitability on the controlling 8 GB target.
