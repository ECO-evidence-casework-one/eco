# Current ECO release gate

**Gate record:** `ECO-RELEASE-GATE-20260803-001`  
**Updated:** 3 August 2026  
**Canonical public status:** [`../../CURRENT_STATUS.md`](../../CURRENT_STATUS.md)  
**Current `main` commit reviewed:** `2dcb44ba8541ab7a319de6e1f14f016aafe2ac1b`  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Signed end-user release:** None

## Controlling release decision

ECO remains a source-development project only.

The following remain blocked:

- use with real, sensitive or irreplaceable evidence;
- ordinary-user binary distribution;
- public preview, release-candidate or stable-release status;
- public-sector or institutional deployment;
- healthcare or clinical deployment;
- EU availability;
- claims of legal, forensic, medical, accessibility or regulatory compliance.

## Current gate status

| Gate | Current position |
|---|---|
| Source identity | `main` contains post-V25 development changes; no later named source milestone is approved |
| Evidence preservation | PR #10 merged materially improved controls, but issue #3 is reopened pending independent closure |
| Ask verification and restore concurrency | Blocked by open issue #12 |
| Candidate workspace identity | PR #11 remains draft; exact candidate binding is not independently approved |
| Migration, rollback and reset | PR #11 remains draft with unresolved P0 filesystem and recovery boundaries |
| Runtime and model supply chain | Blocked until exact runtime/model provenance and every redistributed file are reconciled |
| One-file packaging | Blocked until the actual final embedded executable is supplied and independently inspected |
| Signing order and authenticity | Blocked until the final file is assembled, hashed and Authenticode-signed with no later mutation |
| Privacy and offline proof | Incomplete; runtime network behaviour and diagnostic-export privacy require objective evidence |
| OCR and local AI | Not approved as reliable production functionality |
| Accessibility | Blocked pending keyboard, screen-reader, DPI, contrast, scrolling and cognitive-accessibility evidence |
| Security reporting and maintenance | No complete operational vulnerability-reporting, advisory, update or support-period process |
| Publisher and stewardship | No accountable operating organisation appointed |
| Public claims | Only controlled development claims are permitted |

## Issue and pull-request controls

### Issue #3

Issue #3 is open. The implementation merged through PR #10 is not rejected, but it must not be described as independently closed while:

- issue #12 remains open;
- the issue #3 acceptance checklist is not completed with evidence references;
- exact-commit hostile and concurrency testing remains incomplete.

### Issue #12

Issue #12 blocks issue #3 closure and real-evidence approval. It requires bounded verification cost, safe cache invalidation or source-limited verification, and safe serialisation of Ask against restore.

### PR #11 and issue #4

PR #11 remains a draft. It must not be merged or used to close issue #4 until independent re-review confirms correction of the candidate-identity, migration-record, rollback, reparse-point/reset and preservation-interaction findings.

## Stop rules

### Stop before real-evidence testing

Stop where any of the following remains unresolved:

- preserved bytes, recorded hashes and derived readings cannot be reconciled exactly;
- Ask, restore, migration or reset can observe or create mixed or unsafe state;
- filesystem boundaries are not proven against Windows reparse points, junctions and links;
- diagnostic or runtime records may expose case content without a safe redacted mode;
- endpoint, runtime or local IPC risks remain unqualified;
- any real-evidence P0 or P1 finding remains open.

### Stop before public preview

Stop where:

- the executable is unsigned or is not the actual one-file deliverable;
- accessibility evidence is incomplete;
- application help or AI can invent controls, actions or unsupported conclusions;
- no accountable steward accepts security, privacy, complaints, support and continuity duties;
- public claims cannot be supported by the exact released artefact.

### Stop before GitHub Release

Stop where:

- exact source, immutable commit, build receipt, manifest, SBOM or licence notices are absent;
- final model/runtime provenance is incomplete;
- Smart App Control and clean-machine testing have not passed;
- the file was changed after signing;
- any release-blocking P0 or P1 finding remains unresolved.

## Public-record rule

Public status documents may identify defect classes, acceptance tests and release decisions. They must not expose personal evidence, private diagnostics, unpublished binaries, model files, private workspaces or exploit-level instructions.
