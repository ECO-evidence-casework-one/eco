# Known limitations

The canonical current project position is recorded in [`CURRENT_STATUS.md`](CURRENT_STATUS.md) and the current release decision in [`docs/control/CURRENT_RELEASE_GATE.md`](docs/control/CURRENT_RELEASE_GATE.md). This file summarises limitations rather than opening any release gate.

## Recorded V25 N2 P1 source milestone

- No trusted code-signed end-user binary is available.
- A bundled OCR engine is not included.
- The coordinate-bearing OCR receipt parser and vault gate accept validated worker output, but no production OCR worker is enabled.
- Perspective correction is implemented as a tested transformation foundation; an interactive four-corner editor is not present.
- Auto-crop and deskew are conservative reading-view suggestions, not changes to original evidence.
- Crop and deskew suggestions can be wrong and require visual review.
- Handwriting recognition is not implemented.
- Native PDF page rendering is not implemented.
- A generative local language model is not included.
- Ask ECO remains a deterministic source-backed retrieval engine and is not approved as reliable generative AI.
- Direct Windows Narrator, NVDA, high-DPI and long-duration execution evidence is incomplete.
- The current encrypted workspace format remains an early development format and must not hold real, sensitive or irreplaceable evidence.
- Windows application version-resource metadata and the final installer/uninstaller are incomplete.

## Later `main` source and active gates

- PR #10 materially improved preserved-object source binding, but issue #3 remains open pending issue #12 and complete independent evidence.
- Current Ask verification cost and serialisation against restore remain blocked by issue #12.
- Draft PR #11 contains substantial workspace, migration, reset and restore work but remains blocked by exact-head ownership, concurrency, creation and Linux cleanup findings. None of that draft code is approved merely because its CI passed.
- Current `main` still uses pathname-based portable-restore activation with best-effort rollback. Do not describe it as fully crash-recoverable or object-bound.
- The current `main` Windows PowerShell build script does not reliably fail on every native command exit code. A green artifact job must not be treated as proof that every Windows test command was enforced.
- Ask ECO does not yet enforce the proposed health-related generated-processing boundary; issue #20 blocks health-related and mixed-purpose clinical material from generated processing outside synthetic rejection tests.
- Diagnostic export privacy and final bundled-runtime network behaviour are not independently qualified; issue #14 remains open.
- The root V25 SBOM and third-party notices are source-level historical records, not the actual final packaged-artifact provenance required by issue #15.
- The authoritative intended-purpose and excluded-use control remains draft in PR #18 and issue #16 remains open.
- There is no accountable established publisher, support operator, complaints handler or continuity owner; issue #17 remains open.
- Accessibility, responsiveness and page-aware search acceptance evidence remains incomplete under issues #6–#8.
- No public-sector, institutional, healthcare, clinical or EU deployment is approved.

## Safety position

No current source milestone or development branch should be described as an approved stable release, production-ready, court-ready, forensic, legally compliant, medically compliant, accessible or suitable for real evidence.

Do not disable Smart App Control, antivirus or other Windows protections to run an unsigned ECO artifact. Do not upload real evidence, private diagnostics, credentials or identifying case material to public repository records.
