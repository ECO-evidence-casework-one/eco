# Threat model

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
- corrupt or hostile backups;
- model or component substitution;
- unauthorised local access;
- unintended network communication;
- compromised release pipeline.

## Core controls

- content-signature detection before parsing;
- immutable originals and separate derived data;
- streaming hashes and bounded processing;
- encrypted vault and authenticated backups;
- staged transactional restore;
- source-constrained answers and receipts;
- no AI command execution or arbitrary file access;
- no cloud fallback;
- public build provenance and trusted code signing;
- synthetic adversarial test corpus.

## Out of scope for current preview

The present source is not approved for real evidence. Full parser isolation, bundled OCR, local generative AI, installer hardening and external security review remain future release gates.
