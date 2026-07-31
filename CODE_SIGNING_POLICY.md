# Code signing policy

## Current position

ECO does not yet have an approved trusted code-signing service. Unsigned development executables are not normal end-user releases and may be blocked by Windows Smart App Control.

## Intended SignPath route

The project intends to apply to SignPath Foundation after the repository, public release record and project controls satisfy its conditions.

Upon acceptance, official release pages will state:

> Free code signing provided by SignPath.io, certificate by SignPath Foundation.

No such claim applies before acceptance.

## Team roles

Until dedicated teams are created:

- **Authors, committers and reviewers:** organisation owners with repository write access
- **Signing approvers:** organisation owners authorised to approve releases

Role membership must remain visible through the GitHub organisation and repository permissions.

## Release controls

1. Release binaries are built only from this public repository.
2. Builds run in the declared automated workflow.
3. Tests and source checks must pass.
4. Every signing request requires manual approval.
5. All executable project files in a release must be covered by the approved signing arrangement.
6. Product name and version metadata must match the release.
7. Files must not be modified after signing.
8. SHA-256 checksums are generated after signing.
9. Signature verification evidence accompanies each release.
10. Releases include source, licence, privacy policy and known limitations.

## Privacy statement

This program will not transfer any information to other networked systems unless specifically requested by the user or the person installing or operating it.
