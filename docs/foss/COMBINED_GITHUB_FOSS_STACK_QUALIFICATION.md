# Combined GitHub FOSS stack qualification

Date: 2026-09-05

This staging-only qualification now combines the following independently green GitHub-first product slices/layers without merging them to `main`:

- native PDF reader: `integration/native-pdf-reader-20260904` / PR #114;
- verified Tesseract runtime bundle: `integration/tesseract-runtime-bundle-20260904` / PR #115;
- bounded MBOX reader: `integration/mbox-reader-20260904` / PR #116;
- hostile-input fuzz baseline: `integration/hostile-input-fuzz-baseline-20260904` / PR #117;
- qualified optional PDF page renderer: `integration/pdfium-page-renderer-20260904` / PR #119;
- bounded multi-page PDF preview navigation: `integration/pdfium-page-navigation-20260905` / PR #121.

The first four slices were collision-qualified together by draft PR #118. The PDF renderer was then qualified independently: its exact `klippa-app/pdfium-cli` v0.11.2 Windows WebAssembly runtime passed end-to-end ECO vault registration, preserved-source page rendering, page-specific output and deliberate executable-tamper rejection in Actions run `33913936762`; PR #119 passed the normal Linux, Gitleaks, source-policy, deterministic Windows build, Syft and Cosign pipeline in run `33914224993`.

Adding PR #119 to the already-green #118 stack produced exactly two merge conflicts: `THIRD_PARTY_NOTICES.md` and `cmd/eco/main_windows.go`; `internal/eco/local_tools.go` merged automatically. The staging resolution retained the native PDF and MBOX notices, appended the exact optional PDFium-runtime notice, preserved Tesseract registration/completion handling, placed Tesseract and PDF renderer controls together on the Trust & Settings runtime row, and retained the renderer's citation-page and bounded visual-preview path. Run `33947021567` passed the controlled resolution tests, and the resulting five-slice staging PR #120 then passed the full ordinary ECO Linux/Gitleaks/source-policy/deterministic-Windows/Syft/Cosign pipeline in run `33947320511`.

The multi-page navigation layer reuses the same already-qualified PDFium runtime and adds no new donor. Source qualification run `33947757143` passed tests, vet, Windows cross-build and source policy. Exact-runtime Windows run `33947946928` independently verified the pinned renderer, opened a fresh vault, preserved a deterministic three-page PDF, proved ECO reported exactly three source-bound pages, rendered pages 1/2/3 with distinct page shapes, and rejected page 4. PR #121 then passed its full normal Linux/Gitleaks/source-policy/deterministic-Windows/Syft/Cosign pipeline in run `33948083342`.

Adding PR #121 onto the already-green #120 combined stack produced exactly one conflict: `cmd/eco/main_windows.go`. All navigation backend code, tests and Win32 key constants merged automatically. The controlled resolver retained #120's combined Tesseract/PDF-renderer Settings UI as authoritative and applied only the page-navigation state, page-count-aware initial preview, asynchronous page-change handlers and navigation display controls to that one file. Run `33948381144` passed the exact-conflict resolver plus `go mod tidy`, all Go tests, `go vet`, Windows cross-build, offline/source policy and diff checks before the combined source was persisted.

Temporary resolver/workflow files were removed after qualification. The inherited `scripts/foss/ACQUIRE_DONORS.cmd` was restored byte-for-byte to the `main` CRLF blob (`7c272f0f96b9c8c27cc3902b2e0eeb85c7b74370`), leaving the delta from the already-green #120 staging branch at exactly five navigation files.

This branch remains staging-only. It does not supersede the constituent draft PRs and does not authorize a merge to `main`. The final software-side control is the ordinary complete ECO PR pipeline against this clean combined branch. The separate controlling 8 GB Acer qualification is still required before making low-spec PDF rendering/navigation performance claims.
