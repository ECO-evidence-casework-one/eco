# Native PDF reader integration

Date: 2026-09-04

ECO vendors `ledongthuc/pdf` at exact commit `b3c860c2375335b0bc6676c430107a553725991d` under BSD-3-Clause to provide a small Go-native path for text-bearing PDFs.

The current upstream commit declares Go 1.24.1, but prior qualification proved the complete upstream test suite and known-text extraction under Go 1.23.12 after changing only that module metadata directive. ECO retains the exact source files and records the local Go 1.23 directive in the vendored module.

ECO wraps the donor with additional controls: regular-file and 512 MiB input bounds, 10,000-page and 20,000-segment bounds, panic containment, bounded diagnostics, page-level extraction, per-page warnings, and `Page` / `PageHint` / `Origin=pdf-native` segment provenance. A PDF with no native text is not falsely declared readable; the caller is told that registered local OCR may be required.

This integration deliberately reduces Docling from a critical native-PDF dependency. Docling source remains acquired for later advanced layout/table use, but its standard model pipeline uses separate Hugging Face model assets and is not required for baseline native PDF text extraction.
