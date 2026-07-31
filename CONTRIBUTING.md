# Contributing

ECO welcomes careful contributions that preserve its offline, evidence-safe and accessible design.

## Before contributing

- Read the privacy, security, release and threat-model documents.
- Use synthetic test data only.
- Do not upload personal evidence, credentials, private correspondence or proprietary files.
- Do not add telemetry, advertising, cloud APIs, accounts or online fallbacks.
- Do not add proprietary or ambiguously licensed dependencies.
- Do not make capability claims without tests and evidence.

## Workflow

1. Open or comment on an issue describing the problem or proposal.
2. Create a focused branch.
3. Add or update tests.
4. Run `go test ./...` and `go vet ./...`.
5. Open a pull request describing security, privacy, accessibility and migration effects.
6. A project reviewer must approve changes before merge.

## Component intake

Any proposed OCR engine, AI runtime, model, parser or bundled binary must identify its exact version, source, hashes, licence, redistribution terms, offline behaviour, maintenance position and security implications.

## Licence

By submitting a contribution, you agree that it may be distributed under GPL-3.0-only.
