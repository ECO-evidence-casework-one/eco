# Known limitations

The canonical current project position is recorded in [`CURRENT_STATUS.md`](CURRENT_STATUS.md) and the current release decision in [`docs/control/CURRENT_RELEASE_GATE.md`](docs/control/CURRENT_RELEASE_GATE.md). This file summarises limitations; it does not open any release gate.

**Current reviewed `main`:** `34e96faea669ed29e4c8f63b3c333ee642f29dbf`  
**Status date:** 5 September 2026

## What is no longer a limitation

The following older statements are obsolete on current `main`:

- native PDF text reading is now implemented through a qualified BSD-3-Clause Go reader;
- optional PDF page rendering is implemented through the exact-qualified `pdfium-cli` WebAssembly runtime;
- bounded multi-page PDF preview navigation is implemented;
- Tesseract OCR has a verified complete Windows runtime-bundle path and Settings registration flow;
- MBOX ingestion is implemented with a bounded MIT donor reader;
- permanent hostile-input fuzz targets now cover MBOX, ZIP, EML/MIME, XML/Office text and file sniffing;
- issue #65 is closed and must not be described as blocking every runtime/model integration;
- the four historical unsigned Actions artifacts listed in issue #24 now report `expired: true`.

These improvements do **not** make ECO release-ready or suitable for real evidence.

## Current technical limitations

- Issue #3 remains open: preserved-object/source-truth acceptance is not fully closed for every failure/restart/large/slow-storage path.
- Issue #4 remains open P0: first-run/reopen/reset/workspace ownership and stale-state control are not fully qualified on current `main`.
- Issue #12 remains open: Ask verification cost and Ask/restore serialisation are not fully bounded/qualified.
- Issue #5 remains open: deterministic grounding and llama.cpp integration exist, but the full offline-AI reliability/hostile-prompt/model-regression acceptance is incomplete.
- Issue #6 remains open: responsiveness, cancellation, worker-failure recovery and long-duration low-resource evidence are incomplete across all heavy operations.
- Issue #7 remains open: keyboard-only, Narrator/NVDA, focus, scrolling and 100/150/200% DPI acceptance is incomplete.
- Issue #8 remains open: PDF page navigation exists, but full workspace/document search, deterministic match receipts, visible match highlighting and accessible previous/next match navigation are incomplete.
- Issue #14 remains open: redacted support diagnostics and final clean-machine packet/DNS/firewall-deny qualification are incomplete.
- Handwriting recognition is not implemented.
- Perspective correction remains a transformation foundation; an interactive four-corner editor is not complete.
- Auto-crop/deskew remain reading-view suggestions and may be wrong; originals remain unchanged.
- The exact final GGUF model package for llama.cpp is not approved or bundled. Official Qwen model weights are not distributed as an official GitHub release asset, so model provenance remains a separate controlled decision.
- Docling/OCRmyPDF source/adapters exist, but they are not mandatory critical-path runtimes. Their larger Python/model environments are not treated as proof of a final packaged Windows product.

## Release and distribution limitations

- No trusted Authenticode-signed end-user ECO executable exists.
- No approved final installer/uninstaller exists.
- Windows application version-resource metadata/final packaging remains incomplete.
- Issue #15 remains open: CI generates/reconciles SBOMs and performs private Cosign signing/tamper rehearsal, but that is not the same as a final packaged one-file artefact with complete runtime/model inventory, licences, corresponding source and Authenticode identity.
- Issue #17 remains open: no accountable legal publisher/steward, support operator, complaints handler or continuity owner is appointed.
- Issue #24 remains open even though the four historic artifacts are expired; manual-dispatch, private-handoff, future release-automation and stale-branch controls remain outstanding.
- Smart App Control / ordinary clean-machine acceptance for a final signed candidate is not proved.
- Long-duration stability and crash-recovery evidence for a final candidate is incomplete.
- The CI-built executable remains runner-only and deliberately not uploaded by the normal build workflow.

## Evidence and use limitations

- Real, sensitive or irreplaceable evidence remains blocked.
- Issues #16, #20, #23 and #46 remain open for intended purpose, user-facing claims and prohibited high-consequence outputs.
- ECO must not be presented as a doctor, lawyer, forensic laboratory, emergency service, eligibility/credibility scorer or authority-side decision system.
- Ask/model outputs are not approved as professional advice or independently verified truth.
- Source-linked assistance may be designed around supplied evidence, but users must be able to inspect sources, uncertainty and output status.
- No public-sector, institutional, healthcare, clinical, justice-sector or EU/EEA deployment is approved.

## Low-spec evidence boundary

The controlling 8 GB Acer passed the exact PDF visual-rendering/navigation qualifier:

- 7.68 GiB total RAM;
- about 1.63 GiB available before the test;
- median ECO page render 8103 ms;
- slowest render 8319 ms;
- peak renderer working set 291.4 MiB;
- page count/navigation/out-of-range rejection passed.

This qualifies that PDF path on the Acer. It does **not** prove the whole application, local model, long-duration execution, installer, accessibility or clean-release experience on that machine.

## Historical / superseded controls

- PR #71 remains an open draft historical V40 design lane, but issue #69 supersedes the V40 application shell as a product parent. Do not merge or revive it as the application baseline.
- PR #72 contains useful workspace-ownership primitives but is stale relative to current `main` and explicitly incomplete for issue #4. Reuse/adapt small proven pieces; do not merge wholesale.
- Issue #69 is a superseded historical control record, not an active 9 August delivery instruction.
- Issue #65 is closed; older documents that say it blocks all integration are stale.

## Safety position

No current source milestone should be described as an approved stable release, production-ready, court-ready, forensic, legally compliant, medically compliant, fully accessible or suitable for real evidence.

Do not disable Smart App Control, antivirus or other Windows protections to run an unsigned development build. Do not upload real evidence, private diagnostics, credentials or identifying case material to public repository records.
