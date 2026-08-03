# Evidence & Casework One

**Evidence & Casework One (ECO)** is a free and open-source Windows desktop application under active development for preserving, organising, reviewing and understanding evidence and casework locally.

> **Development status:** early native source development. There is no approved signed end-user release. Do not use development builds with real, sensitive or irreplaceable evidence.

## Project principles

- Fully offline application operation
- No accounts, cloud service, telemetry or advertising
- Original evidence preserved separately from derived readings
- Streaming local hashing, encrypted storage and integrity checks
- Source-backed outputs that distinguish documents, OCR suggestions, ECO suggestions, user confirmations and notes
- Calm, accessible and plain-language operation
- Free and open-source project source, with exact bundled-component licensing and provenance required before distribution

## Current recorded milestone

`ECO-V25-20260731-N2-P1`  
**Version 25 N2 — Native Document Vision Foundation Preview 1**

This remains the value recorded in `VERSION` and the last named source milestone approved under the earlier milestone process.

The `main` branch now also contains later controlled source-development changes. Those later commits do not automatically create a new approved source milestone, release candidate or end-user release.

The V25 N2 P1 foundation introduced:

- conservative photographed-page boundary detection;
- non-destructive auto-crop and deskew preview modes;
- skew, glare, uneven-lighting, edge-cutoff and probable double-page assessment;
- adaptive black-and-white reading enhancement;
- bounded high-resolution preview processing;
- perspective-correction foundations;
- coordinate-bearing OCR words, lines, receipts and source segments;
- source-hash and coordinate validation foundations;
- exact OCR citation-region highlighting foundations;
- exclusion of very low-confidence OCR segments from Ask ECO retrieval.

A bundled production OCR engine, OCR language/model package and generative local language model are **not** part of the recorded V25 milestone. V25 provides document-vision and OCR-provenance foundations rather than approved production OCR or generative-AI operation.

## Current controlled development

Later public source-development work is tracked through narrow issues and pull requests with explicit acceptance tests.

- PR #10 merged evidence-preservation and source-binding changes, but issue #3 is reopened pending independent closure and follow-up issue #12.
- PR #11 remains a draft for candidate-specific workspace state, migration, recovery and reset. It is not approved.
- Issues #5–#8 continue to track offline-AI behaviour, responsiveness, accessibility and document navigation.
- Issues #14–#17 track diagnostic privacy and offline claims, actual-build SBOM/licensing/provenance, intended purpose and excluded uses, and accountable publisher/continuity arrangements.

See the [canonical current project status](CURRENT_STATUS.md) and the [current release gate](docs/control/CURRENT_RELEASE_GATE.md).

## Important release warning

There is currently **no approved signed end-user release**. New unsigned Windows executables may be blocked by Windows Smart App Control. Do not weaken Windows security controls to run an ECO development build.

Official releases will appear only on this repository's **Releases** page and will include checksums, source provenance, known limitations and signature-verification instructions.

## Documentation

- [Current project status](CURRENT_STATUS.md)
- [Current release gate](docs/control/CURRENT_RELEASE_GATE.md)
- [Known limitations](KNOWN_LIMITATIONS.md)
- [Privacy and offline operation](PRIVACY.md)
- [Security policy](SECURITY.md)
- [Code signing policy](CODE_SIGNING_POLICY.md)
- [Maintainers and signing roles](MAINTAINERS.md)
- [Building from source](BUILDING.md)
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

This program will not transfer information to other networked systems unless specifically requested by the user or the person installing or operating it. Any local inter-process communication used by a later candidate must be accurately documented and independently tested before release claims are made.

## Licence

ECO is licensed under **GNU General Public License v3.0 only** (`GPL-3.0-only`). See [LICENSE](LICENSE).

Copyright © 2026 Evidence & Casework One contributors.
