# Code signing policy

## Current position

ECO does not yet have an approved trusted code-signing service. Unsigned development executables are provenance artifacts only, are not normal end-user releases and may be blocked by Windows Smart App Control.

The SignPath Foundation application process was initiated on 31 July 2026. The web form did not display a submission confirmation, so SignPath support was contacted to confirm receipt or provide an alternative submission route. No approval is assumed.

## Intended SignPath route

If SignPath Foundation accepts ECO, official signed release pages will state:

> Free code signing provided by SignPath.io, certificate by SignPath Foundation.

No such claim applies before acceptance.

## Team roles

Current named role assignments are published in [MAINTAINERS.md](MAINTAINERS.md).

- **Authors and committers:** people authorised to modify project source and build scripts.
- **Reviewers:** maintainers responsible for reviewing contributions from people without direct write access.
- **Signing approvers:** maintainers authorised to manually approve an exact automated build for signing.

During the early one-maintainer phase, one person may hold all roles. Signing approval remains a separate manual action and may not be automated away.

## Release controls

1. Release binaries are built only from this public repository.
2. Source, build scripts and CI configuration are part of the reviewed signing input.
3. Builds run in the declared automated workflow.
4. Tests, source-policy checks and deterministic-rebuild checks must pass.
5. Every signing request requires manual approval.
6. Only ECO artifacts built from ECO-maintained source may be signed under the project arrangement.
7. Product name and version metadata must match the release.
8. Files must not be modified after signing.
9. SHA-256 checksums are generated after signing.
10. Signature-verification evidence accompanies each signed release.
11. Releases include source, licence, privacy policy, SBOM and known limitations.
12. A failed or disputed release is paused rather than bypassing signing controls.

## Privacy statement

This program will not transfer any information to other networked systems unless specifically requested by the user or the person installing or operating it.
