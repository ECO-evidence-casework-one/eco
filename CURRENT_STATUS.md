# Current ECO project status

**Status date:** 3 August 2026  
**Canonical public status record:** this file  
**Current `main` commit reviewed for this status:** `2dcb44ba8541ab7a319de6e1f14f016aafe2ac1b`  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Release position:** development only; no approved public binary

## Status authority

This root `CURRENT_STATUS.md` file is the single current public status authority.

Files under `docs/status/` with dates are historical daily records. The former `docs/status/CURRENT_STATUS.md` path is retained only as a pointer to this canonical record.

## What `main` represents

`VERSION` still records the last named source milestone approved under the earlier milestone process: `ECO-V25-20260731-N2-P1`.

The `main` branch now also contains later controlled source-development changes. Those changes do **not** create a new approved source milestone, release candidate or end-user release merely because they were merged.

No later private executable, model bundle, diagnostic archive, screenshot set, private workspace or personal evidence is represented by this repository.

## Current controlled development position

### Evidence preservation

PR #10 merged source changes intended to make the verified preserved object the source of truth for extraction, OCR, citation and retrieval.

Issue #3 has been reopened. Its implementation is materially improved, but independent closure is pending because post-merge review identified follow-up work recorded in issue #12. Until that work and the complete acceptance evidence are independently verified, issue #3 remains open and use with real, sensitive or irreplaceable evidence remains blocked.

### Workspace identity, migration and reset

PR #11 remains a draft. It proposes clean candidate-specific workspace state, migration, recovery and reset controls, but independent review identified unresolved P0 boundaries. It must not be treated as approved or merged until those findings are corrected and re-reviewed.

### Other active grouped work

- Issue #5: instruction-faithful, source-backed and non-operative offline AI.
- Issue #6: responsive background import, OCR and local-AI work.
- Issue #7: keyboard, screen-reader, DPI, scrolling and working-control accessibility.
- Issue #8: page-aware document search, visible highlights and source navigation.
- Issue #12: bounded Ask verification cost and safe serialisation against restore.

## Current stop gates

The following remain blocked:

- use with real, sensitive or irreplaceable evidence;
- public end-user binary distribution;
- release-candidate or stable-release status;
- claims of reliable generative offline AI assistance;
- claims of completed production OCR or complete native PDF investigation;
- claims of accessibility, forensic, legal, medical or regulatory compliance;
- public-sector, healthcare or EU deployment.

## Release prerequisites

Before the release position can be reconsidered, ECO requires objective evidence for at least:

- exact source, build and binary identity;
- complete corresponding source, manifest, SBOM and licence notices;
- verified model and runtime provenance;
- genuine one-file packaging and clean-machine execution;
- Windows stability, low-resource and long-duration testing;
- no external-network proof for all bundled runtime components;
- evidence-integrity and recovery qualification;
- keyboard, assistive-technology, DPI and cognitive-accessibility evidence;
- trusted Authenticode signing with no post-signing file mutation;
- an accountable publisher or steward for security, privacy, complaints, support and continuity;
- independent closure of all release-blocking P0 and P1 findings.

## Public-record rule

Public updates must remain sanitised. They may record defect classes, controls, acceptance tests and release decisions, but must not publish personal evidence, private diagnostics, exploit-level instructions, confidential screenshots, unpublished binaries, bundled models or private test workspaces.

Use synthetic and non-sensitive information only in this repository.
