# Evidence & Casework One

**Evidence & Casework One (ECO)** is a free and open-source Windows desktop application under active development for preserving, organising, reviewing and understanding evidence and casework locally.

> **Development status:** early native preview. Do not use the current development source or unsigned builds with real, sensitive or irreplaceable evidence.

## Project principles

- Fully offline operation
- No accounts, cloud service, telemetry or advertising
- Original evidence preserved separately from derived readings
- Local hashing, encrypted storage and integrity checks
- Source-backed outputs that distinguish documents, ECO suggestions, user confirmations and notes
- Calm, accessible and plain-language operation
- Free and open-source components only

## Current build

`ECO-V25-20260730-N1-P3`  
Version 25 N1 — Native Evidence & Vision Foundation Preview 3

The current source provides a native Windows interface, encrypted local vault, streaming file intake, content-signature checks, duplicate detection, native image preview and deterministic source-backed evidence search. Automatic OCR and a generative local language model are not yet bundled.

## Important release warning

There is currently **no approved signed end-user release**. New unsigned Windows executables may be blocked by Windows Smart App Control. Do not weaken Windows security controls to run an ECO development build.

Official releases will appear only on this repository's **Releases** page and will include checksums, source provenance, known limitations and signature verification instructions.

## Documentation

- [Privacy and offline operation](PRIVACY.md)
- [Security policy](SECURITY.md)
- [Code signing policy](CODE_SIGNING_POLICY.md)
- [Building from source](BUILDING.md)
- [Release policy](RELEASE_POLICY.md)
- [Roadmap](ROADMAP.md)
- [Threat model](THREAT_MODEL.md)
- [Contributing](CONTRIBUTING.md)
- [Governance](GOVERNANCE.md)

## Code signing policy

ECO is preparing an application for free open-source code signing through SignPath Foundation. Until acceptance, no claim is made that SignPath provides signing for ECO. See [CODE_SIGNING_POLICY.md](CODE_SIGNING_POLICY.md).

## Licence

ECO is licensed under **GNU General Public License v3.0 only** (`GPL-3.0-only`). See [LICENSE](LICENSE).

Copyright © 2026 Evidence & Casework One contributors.
