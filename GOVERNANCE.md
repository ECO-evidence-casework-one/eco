# Governance

ECO is maintained through the [`ECO-evidence-casework-one`](https://github.com/ECO-evidence-casework-one) GitHub organisation.

## Roles

Current named assignments are listed in [MAINTAINERS.md](MAINTAINERS.md).

- **Organisation owners:** project stewardship, repository administration and signing approval.
- **Committers:** trusted contributors with direct write access.
- **Reviewers:** contributors authorised to review external contributions.
- **Release/signing approvers:** maintainers authorised to manually approve signing and publication.

One person may initially hold more than one role. Roles and permissions must remain visible and reviewable. A signing request remains a separate manual approval even when the approver also authored the change.

## Controlling product boundary

The official ECO intended purpose, excluded uses and public-claims boundary is defined in [`docs/governance/INTENDED_PURPOSE_AND_CLAIMS_CONTROL.md`](docs/governance/INTENDED_PURPOSE_AND_CLAIMS_CONTROL.md).

That record becomes controlling only when merged into the canonical branch after independent review. Governance adoption must remain separate from:

- proof that the application implements the boundary;
- legal or regulatory classification;
- appointment of a publisher, supplier, controller or professional role;
- real-evidence, signing, release or deployment approval.

The current release gate remains controlling where implementation evidence is incomplete or public wording conflicts.

## Decision principles

Priority order:

1. evidence integrity and user safety;
2. offline privacy and security;
3. factual and source-backed behaviour;
4. accessibility and usability;
5. free and open-source compatibility;
6. performance and additional features.

## Changes

Substantial architecture, intended-purpose, licence, privacy, signing or release-policy changes require a documented issue or pull request and explicit owner approval.
