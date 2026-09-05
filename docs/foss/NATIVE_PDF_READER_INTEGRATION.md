# Native PDF reader integration

Date: 2026-09-04

ECO vendors `ledongthuc/pdf` at exact commit `b3c860c2375335b0bc6676c430107a553725991d` under BSD-3-Clause to provide a small Go-native path for text-bearing PDFs.

The current upstream commit declares Go 1.24.1, but prior qualification proved the complete upstream test suite and known-text extraction under Go 1.23.12 after changing only that module metadata directive. ECO retains the exact source files and records the local Go 1.23 directive in the vendored module.

ECO wraps the donor with additional controls: regular-file and 512 MiB input bounds, 10,000-page and 20,000-segment bounds, panic containment, bounded diagnostics, page-level extraction, per-page warnings, and `Page` / `PageHint` / `Origin=pdf-native` segment provenance. A PDF with no native text is not falsely declared readable; the caller is told that registered local OCR may be required.

The donor is compiled through a local Go `replace` under `third_party/`. Syft 1.51.1 identifies this compiled module in binary and source scans but reports its version as `UNKNOWN`, while Go build-info records the exact compiled pseudo-version and local replacement path. ECO therefore preserves Syft's raw JSON/SPDX files and reconciles only a matching `UNKNOWN` local-replace package after validating all of the following: the replacement stays inside the controlled `third_party/` boundary; its `go.mod` module identity matches the compiled dependency; retained licence and `ECO_PROVENANCE.md` files exist; the pseudo-version commit is bound by provenance; and `THIRD_PARTY_NOTICES.md` identifies the component. The final reconciled SBOMs and both raw scanner outputs are SHA-256-bound in `sbom-receipt.json`. Arbitrary `UNKNOWN` packages are not accepted.

This integration deliberately reduces Docling from a critical native-PDF dependency. Docling source remains acquired for later advanced layout/table use, but its standard model pipeline uses separate Hugging Face model assets and is not required for baseline native PDF text extraction.
