# Evidence & Casework One

**Evidence & Casework One (ECO)** is a free and open-source Windows desktop application under active development for preserving, organising, reviewing and understanding evidence and casework locally.

> **Development status:** early native source development. There is no approved signed end-user release. Do not use development builds with real, sensitive or irreplaceable evidence.

## Project principles

- Fully offline application operation
- No accounts, cloud service, telemetry or advertising
- Original evidence preserved separately from derived readings
- Streaming local hashing, encrypted storage and integrity checks
- Source-backed outputs that distinguish documents, OCR suggestions, ECO suggestions, user confirmations and notes
- Calm, accessible and plain-language operation as a design objective; accessibility conformance is not yet claimed
- Free and open-source project source, with exact bundled-component licensing and provenance required before distribution

## Current recorded milestone

`ECO-V25-20260731-N2-P1`  
**Version 25 N2 — Native Document Vision Foundation Preview 1**

This remains the value recorded in `VERSION` and the last named source milestone approved under the earlier milestone process.

The `main` branch now also contains later controlled source-development and repository-control changes. Those later commits do not automatically create a new approved source milestone, release candidate or end-user release.

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
- PR #11 remains a draft for candidate-specific workspace state, migration, recovery and reset. It contains material progress but remains blocked by exact-head ownership, concurrency and filesystem-object findings recorded in issue #22 and on the PR.
- Draft PR #18 proposes intended-purpose and public-claims controls; issue #20 records the unresolved health-related generated-processing implementation gate.
- Issues #5–#8 continue to track offline-AI behaviour, responsiveness, accessibility and document navigation.
- Issues #14–#17 track diagnostic privacy and offline claims, actual-build SBOM/licensing/provenance, intended purpose and excluded uses, and accountable publisher/continuity arrangements.
- P0 issue #24 tracks public Actions distribution of unsigned executables. New uploads were stopped on `main`, but historical artifacts remain unapproved until deleted or expired and the correction is fully evidenced.

See the [canonical current project status](CURRENT_STATUS.md), the [current release gate](docs/control/CURRENT_RELEASE_GATE.md) and the operational [control board](https://github.com/ECO-evidence-casework-one/eco/issues/22).

## Important release warning

There is currently **no approved signed end-user release**. New unsigned Windows executables may be blocked by Windows Smart App Control. Do not weaken Windows security controls to run an ECO development build.

Official releases will appear only on this repository's **Releases** page and will include checksums, source provenance, known limitations and signature-verification instructions.

GitHub Actions artifacts are also a binary-distribution surface. They are not an alternative release channel. At `main` commit `bdc05df444d21d739abf83fa9cf768fc4ab5dd9a`, ECO stopped uploading `ECO.exe` from the public workflow while retaining internal Windows compilation and testing. Four earlier unsigned Actions executable packages identified by issue #24 are not approved test candidates or releases and must not be used or redistributed.

The current `main` Windows build script is not itself an approved release gate. A fail-fast native-command correction exists only in blocked draft PR #11 and has not yet been merged. Do not treat a green Windows job on `main` as independent proof that every native test command was enforced.

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

This program will not transfer information to other networked systems unless specifically requested by the user or the person installing or operating it. This describes the current source-level product rule, not independent qualification of a future bundled runtime. Any local inter-process communication used by a later candidate must be accurately documented and independently tested before release claims are made.

## Licence

ECO is licensed under **GNU General Public License v3.0 only** (`GPL-3.0-only`). See [LICENSE](LICENSE).

Copyright © 2026 Evidence & Casework One contributors.
