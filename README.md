# Evidence & Casework One

**Evidence & Casework One (ECO)** is a free and open-source Windows desktop application under active development for preserving, organising, reviewing and understanding evidence and casework locally.

> **Development status:** early native source preview. Do not use unsigned development builds with real, sensitive or irreplaceable evidence.

## Project principles

- Fully offline application operation
- No accounts, cloud service, telemetry or advertising
- Original evidence preserved separately from derived readings
- Streaming local hashing, encrypted storage and integrity checks
- Source-backed outputs that distinguish documents, OCR suggestions, ECO suggestions, user confirmations and notes
- Candidate-specific application state with explicit create, reopen, migrate, recover and selected-workspace reset flows
- Calm, accessible and plain-language operation
- Free and open-source components only

## Current source milestone

`ECO-V25-20260731-N2-P1`  
**Version 25 N2 — Native Document Vision Foundation Preview 1**

The N2 P1 source adds:

- conservative photographed-page boundary detection;
- non-destructive auto-crop and deskew preview modes;
- skew, glare, uneven-lighting, edge-cutoff and probable double-page assessment;
- adaptive black-and-white reading enhancement;
- bounded high-resolution preview processing;
- perspective-correction foundations;
- coordinate-bearing OCR words, lines, receipts and source segments;
- source-hash and coordinate validation before OCR results can enter the encrypted workspace;
- exact OCR citation-region highlighting foundations;
- exclusion of very low-confidence OCR segments from Ask ECO retrieval.

A bundled OCR engine and generative local language model are **not yet included**. The OCR provenance and source-region system is implemented first so future engines cannot inject unvalidated text into the evidence index.

## Later local development and independent control

The source currently published on `main` remains `ECO-V25-20260731-N2-P1`.

Later native Windows candidates are developed and independently inspected before any source promotion. These private candidates do not automatically replace the public source milestone and are not releases.

`ECO-V32-20260801-M1-P1` was independently reproducible but was held before Windows execution because runtime-provenance, final one-file packaging and current release-evidence requirements were incomplete.

Version 34 is under active development and has not yet entered independent intake.

See [Current project status](docs/status/CURRENT_STATUS.md) and the [current release gate](docs/control/CURRENT_RELEASE_GATE.md).

## Important release warning

There is currently **no approved signed end-user release**. New unsigned Windows executables may be blocked by Windows Smart App Control. Do not weaken Windows security controls to run an ECO development build.

Official releases will appear only on this repository's **Releases** page and will include checksums, source provenance, known limitations and signature-verification instructions.

## Documentation

- [Known limitations](KNOWN_LIMITATIONS.md)
- [Privacy and offline operation](PRIVACY.md)
- [Security policy](SECURITY.md)
- [Code signing policy](CODE_SIGNING_POLICY.md)
- [Maintainers and signing roles](MAINTAINERS.md)
- [Building from source](BUILDING.md)
- [Development workspace lifecycle](docs/architecture/WORKSPACE_LIFECYCLE.md)
- [Release policy](RELEASE_POLICY.md)
- [Roadmap](ROADMAP.md)
- [Threat model](THREAT_MODEL.md)
- [Contributing](CONTRIBUTING.md)
- [Governance](GOVERNANCE.md)

## Code signing policy

ECO has initiated the SignPath Foundation application process. The application has **not** been approved and no claim is made that SignPath currently signs ECO.

If accepted, official signed release pages will include the required wording:

> Free code signing provided by SignPath.io, certificate by SignPath Foundation.

See the [Code signing policy](CODE_SIGNING_POLICY.md), [maintainer roles](MAINTAINERS.md) and [application-status record](docs/security/SIGNPATH_APPLICATION_STATUS.md).

## Privacy statement

This program will not transfer any information to other networked systems unless specifically requested by the user or the person installing or operating it.

## Licence

ECO is licensed under **GNU General Public License v3.0 only** (`GPL-3.0-only`). See [LICENSE](LICENSE).

Copyright © 2026 Evidence & Casework One contributors.
