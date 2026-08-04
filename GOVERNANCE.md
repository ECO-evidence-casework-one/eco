# Governance

ECO is maintained through the [`ECO-evidence-casework-one`](https://github.com/ECO-evidence-casework-one) GitHub organisation.

## Roles

Current named assignments are listed in [MAINTAINERS.md](MAINTAINERS.md).

- **Organisation owners:** project stewardship, repository administration and signing approval.
- **Committers:** trusted contributors with direct write access.
- **Reviewers:** contributors authorised to review external contributions.
- **Release/signing approvers:** maintainers authorised to manually approve signing and publication.

One person may initially hold more than one technical role during source development. Roles and permissions must remain visible and reviewable. A signing request remains a separate manual approval even when the approver also authored the change.

Technical repository roles do not appoint a legal publisher, supplier, director, trustee, controller, support operator, complaints handler or liability owner.

## Controlling product boundary

The official ECO intended purpose, excluded uses and public-claims boundary is defined in [`docs/governance/INTENDED_PURPOSE_AND_CLAIMS_CONTROL.md`](docs/governance/INTENDED_PURPOSE_AND_CLAIMS_CONTROL.md).

That record becomes controlling only when merged into the canonical branch after independent review. Governance adoption must remain separate from:

- proof that the application implements the boundary;
- legal or regulatory classification;
- appointment of a publisher, supplier, controller or professional role;
- real-evidence, signing, release or deployment approval.

The current release gate remains controlling where implementation evidence is incomplete or public wording conflicts.

## Publisher and stewardship gate

The organisational conditions for an official ECO publisher or steward are defined in:

- [`docs/governance/PUBLISHER_AND_STEWARDSHIP_GATE.md`](docs/governance/PUBLISHER_AND_STEWARDSHIP_GATE.md);
- [`docs/governance/PUBLISHER_ACCEPTANCE_CHECKLIST.md`](docs/governance/PUBLISHER_ACCEPTANCE_CHECKLIST.md).

ECO currently has no appointed official publisher or operating organisation.

The preferred future route is an existing established nonprofit, public-interest institution, open-source foundation or equivalent capable organisation that formally accepts the defined duties. The project does not require the originating developer or another contributor to form a company, CIC, charity or other entity or to become a director, trustee or equivalent office-holder.

Contribution or repository administration alone does not appoint organisational roles. This does not prevent duties arising from actual publishing, signing, contracting, data processing, product claims or applicable law. Individuals must not perform official supply or publishing acts on behalf of ECO before organisational authority and role allocation are documented.

Merging the publisher gate does not satisfy issue #17. Issue #17 closes only after a named organisation completes due diligence, its governing body formally accepts the role and the operational routes and continuity controls pass testing.

## Decision principles

Priority order:

1. evidence integrity and user safety;
2. offline privacy and security;
3. factual and source-backed behaviour;
4. accessibility and usability;
5. free and open-source compatibility;
6. performance and additional features.

## Changes

Substantial architecture, intended-purpose, licence, privacy, publisher, signing or release-policy changes require a documented issue or pull request and explicit owner approval.
