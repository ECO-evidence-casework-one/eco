# Threat model

## Status of this record

This document identifies protected assets, principal threats and required control families. It does not certify that every listed control is fully implemented or independently qualified on the current `main` branch. The canonical implementation and release position is recorded in [`CURRENT_STATUS.md`](CURRENT_STATUS.md) and [`docs/control/CURRENT_RELEASE_GATE.md`](docs/control/CURRENT_RELEASE_GATE.md).

## Protected assets

- original evidence bytes;
- hashes and provenance;
- OCR and extracted text;
- user confirmations and notes;
- matter records and timelines;
- backups and encryption keys;
- audit and change records.

## Principal threats

- malicious or malformed imported files;
- disguised executables and unsafe archives;
- parser crashes and memory exhaustion;
- prompt-injection wording inside evidence;
- unsupported AI claims;
- accidental modification or deletion;
- stale or concurrent writers silently erasing newer workspace state;
- path aliases, symlinks, junctions, reparse points and parent substitution;
- corrupt or hostile backups;
- interrupted creation, migration, reset or restore;
- model or component substitution;
- unauthorised local access;
- unintended network communication;
- diagnostic or support exports disclosing case content;
- compromised release pipeline;
- loss of the only repository, signing or operational owner.

## Required control families

- content-signature detection before parsing;
- application-preserved originals, separate derived data and fresh source verification before downstream use;
- streaming hashes and bounded processing;
- encrypted workspace and authenticated backups;
- explicit workspace creation, reopen, migration, reset and recovery transactions;
- alias-safe object-bound ownership and stale-writer conflict prevention;
- staged backup validation plus independently qualified, crash-recoverable activation and rollback;
- source-constrained answers and receipts;
- restricted-input gates for health-related and other excluded generated-processing uses;
- no AI command execution or arbitrary file access;
- no cloud fallback and independently verified bundled-runtime network behaviour;
- privacy-safe diagnostic export;
- public build provenance, complete corresponding source and trusted code signing;
- synthetic adversarial and concurrency test corpus;
- accountable publisher, security-response and continuity arrangements before distribution.

## Current unqualified boundaries

The present source is not approved for real evidence.

Current `main` performs staged backup validation but still activates portable restore through pathname-based renames with best-effort rollback. Draft PR #11 contains substantial object-bound and authenticated-restore work but remains blocked by ownership, concurrency, creation and Linux cleanup findings.

Full parser isolation, production OCR, local generative AI, accessibility qualification, installer hardening, privacy-safe diagnostics, final bundled-runtime network proof, actual-build SBOM reconciliation and external security review remain release gates.
