# ECO publisher and stewardship gate

**Control status:** draft governance control for issue #17  
**Control date:** 3 August 2026  
**Current position:** no accountable publisher or operating organisation appointed; all public, institutional, healthcare and EU gates remain closed

## Purpose

This document defines the organisational responsibilities that must be accepted before Evidence & Casework One (ECO) can be distributed to ordinary users or supplied for institutional use.

It is a release and governance gate. It does not appoint a publisher, create a company, transfer liability, approve a release or require any individual developer to accept organisational duties.

## Current control decision

ECO currently has no settled legal organisation responsible for publishing, support, security response, privacy complaints, accessibility complaints, contracts, insurance, continuity or withdrawal of releases.

Development with synthetic and non-sensitive information may continue. Public binary distribution, institutional procurement, healthcare use and EU availability remain blocked.

## Non-assignment rule

No project document, repository setting, pull request, issue, release record or informal practice may assign the following responsibilities to Karl, another individual contributor or a sole GitHub account by default:

- legal publisher or manufacturer;
- supplier or contracting party;
- company director, trustee or equivalent office-holder;
- data controller or processor for users' case material;
- privacy, accessibility or product complaints handler;
- security vulnerability and incident-response operator;
- user-support operator;
- code-signing certificate owner or sole release approver;
- insurer, indemnifier or personal liability bearer;
- public spokesperson for regulated, clinical, forensic or legal claims.

An individual may accept a defined role only through a separate, explicit and informed decision with appropriate organisational support. Contribution to source code or documentation does not itself constitute acceptance of any role above.

## Required accountable organisation

Before ordinary-user or institutional distribution, an established legal organisation must formally accept the ECO publisher and stewardship role.

The organisation must be capable of:

- entering contracts in its own name;
- maintaining appropriate governance and financial records;
- operating monitored communication routes;
- managing security and product incidents;
- making release, withdrawal and end-of-support decisions;
- obtaining legal, accessibility, security, privacy and insurance advice where required;
- maintaining continuity when an individual maintainer is unavailable.

The organisation's name, legal status, jurisdiction, address or formal service route, responsible governing body and relevant contacts must be recorded before release.

## Responsibility matrix

### 1. Publishing and release authority

The accountable organisation must:

- decide whether a candidate may be published;
- confirm that source, binary, manifest, SBOM, licence notices, hashes and signatures identify the same artefact;
- maintain the authoritative release page and approved download route;
- prohibit unofficial or post-signing modification of release files;
- withdraw, revoke or warn against a release when material risk is discovered;
- preserve release and withdrawal decisions in an auditable record.

### 2. Security vulnerability response

The accountable organisation must:

- operate a monitored private vulnerability-reporting route;
- publish supported versions and realistic response expectations;
- triage reports using a documented severity process;
- preserve confidential evidence and avoid requiring public disclosure of exploit detail;
- coordinate fixes, mitigations, advisories and release withdrawal;
- communicate clearly whether an update is a temporary mitigation or a complete fix;
- perform post-incident review and maintain a vulnerability record.

A public repository issue must not be the only route for reporting a sensitive vulnerability.

### 3. Software maintenance and support lifecycle

The accountable organisation must:

- define the support period for each release;
- identify which versions receive security fixes;
- provide a safe manual update and rollback process suitable for an offline application;
- notify users of material security updates and end-of-support dates;
- define what happens when a bundled model, runtime, parser or dependency becomes unsupported or vulnerable;
- maintain an exit path allowing users to export or retain their own records without dependence on the publisher.

Support commitments must be deliverable and must not depend solely on spare personal capacity.

### 4. Privacy and data-role control

For each distribution or deployment model, the accountable organisation must determine and document:

- whether it is a controller, joint controller, processor or neither for each data flow;
- whether support, diagnostics, crash reports or complaint handling may receive personal data;
- which organisation receives and stores those records;
- lawful basis, purpose, minimisation, access, retention and deletion arrangements where applicable;
- breach and complaint escalation routes;
- restrictions preventing users from being asked to publish real evidence or case records in public support channels.

Where the organisation acts as a controller and data-protection complaint duties apply, it must provide a clear complaint route, acknowledge complaints within the applicable period, investigate them and communicate the outcome.

### 5. Accessibility complaints and reasonable adjustments

The accountable organisation must:

- provide a monitored route for reporting accessibility barriers;
- distinguish accessibility reports from security, privacy and general product complaints;
- record, prioritise and respond to barriers affecting disabled users;
- provide accessible alternative communication and reasonable adjustment routes where required;
- maintain truthful accessibility documentation and avoid unsupported conformance claims.

### 6. Product complaints and harmful-output incidents

The accountable organisation must operate a process for:

- materially misleading output;
- invented controls or actions;
- evidence-provenance failure;
- harmful legal, medical, forensic or high-consequence suggestions;
- loss, corruption or unintended disclosure of local records;
- repeated usability or accessibility failures affecting vulnerable users.

The process must distinguish urgent security or safeguarding concerns from ordinary support requests and must define escalation, reproduction, containment, correction and user communication.

### 7. Code signing and release authenticity

The accountable organisation must:

- control or formally govern access to signing services and certificates;
- separate build, review and signing approval where practical;
- require manual approval of the exact final artefact;
- preserve signing logs, timestamps, certificate status and file hashes;
- maintain recovery and revocation arrangements;
- ensure no file mutation occurs after signing.

No single unrecoverable personal account may be the sole long-term signing authority.

### 8. Repository and organisation continuity

The accountable organisation must:

- maintain more than one protected recovery owner where the platform permits;
- use strong authentication and recovery controls;
- review repository, organisation, release and signing permissions periodically;
- preserve recovery instructions outside one person's device or account;
- test recovery from account loss, maintainer incapacity and unauthorised access;
- define who can freeze, archive or transfer the repository during an incident.

### 9. Independent review and risk acceptance

High-risk changes must not be approved solely by their author.

The organisation must define:

- which changes require independent technical, legal, privacy, accessibility or regulatory review;
- who may approve a release;
- who may accept residual risk;
- when risk acceptance expires;
- which risks may never be accepted without external specialist advice;
- how dissenting review findings are preserved.

A high or critical exception must be written, evidence-based, time-limited and approved by someone other than the change author where practical.

### 10. Contracts, procurement, liability and insurance

Before institutional supply, the accountable organisation must provide or decide:

- the contracting legal entity;
- product and service description;
- intended purpose and prohibited uses;
- licence and support terms;
- service, security and accessibility responsibilities;
- controller/processor allocation;
- incident and complaint routes;
- limitation and allocation of liability;
- insurance position;
- exit, data portability and termination arrangements;
- accurate AI and supply-chain disclosures requested by the buyer.

The repository owner or individual developer must not make warranties or indemnities on behalf of a non-existent organisation.

### 11. Regulatory ownership

Before any regulated or geographically expanded use, the accountable organisation must own the decision and supporting assessment for:

- UK medical-device or digital mental-health boundaries;
- forensic or criminal-justice use;
- public-sector and high-consequence decision support;
- EU AI Act roles and obligations;
- EU Cyber Resilience Act manufacturer, open-source steward or other role;
- product-liability and consumer-protection exposure;
- any required registration, conformity, reporting or market-surveillance contact.

EU availability remains blocked until a responsible legal person and operational reporting route have been established where required.

## Required public routes

Before release, the following routes must exist and be tested:

| Route | Minimum purpose |
|---|---|
| Security | Private reporting of vulnerabilities and sensitive technical findings |
| Privacy | Data-protection questions and complaints where the organisation has a relevant role |
| Accessibility | Barriers, adjustments and accessible alternatives |
| Product complaints | Misleading output, evidence handling, reliability and safety concerns |
| General support | Installation, ordinary use and published support boundaries |
| Legal notices | Formal contact for licences, rights, contracts and notices |

The routes may be operated by the same organisation, but their responsibilities and handling rules must be distinguishable.

## Minimum evidence pack

The publisher gate cannot pass without:

- formal organisational acceptance of the publisher/stewardship role;
- legal identity and governing authority record;
- named role-holders and deputies;
- repository and signing access register;
- supported-version and end-of-support policy;
- vulnerability disclosure and response procedure;
- release, update, rollback and withdrawal procedure;
- privacy and deployment-role matrix;
- complaints and accessibility procedures;
- incident-response and harmful-output procedure;
- continuity and account-recovery plan;
- contracting, liability and insurance decision record;
- tabletop or live tests of the principal routes and recovery controls.

## Acceptance tests

- [ ] A named legal organisation has formally accepted the publisher and stewardship role.
- [ ] The acceptance record identifies its governing authority and authorised decision-makers.
- [ ] No release, repository, signing or emergency-recovery function depends on one unrecoverable personal account.
- [ ] A synthetic private vulnerability report reaches the monitored route and follows the documented process.
- [ ] Supported-version, update, rollback, withdrawal and end-of-support procedures pass a tabletop exercise.
- [ ] Privacy, accessibility, product-complaint and general-support routes are live and correctly separated.
- [ ] A synthetic incident demonstrates triage, containment, communication, correction and post-incident review.
- [ ] An account-loss or maintainer-unavailability exercise demonstrates repository and release continuity.
- [ ] Institutional materials identify the contracting party, support model, data roles, liability position, insurance decision and exit arrangements.
- [ ] Role records confirm that Karl or another contributor has not been assigned organisational duties merely by developing or contributing to ECO.
- [ ] Public documentation identifies the accountable organisation without implying certification, regulatory approval or wider responsibility than it has accepted.

## Stop rules

Stop before ordinary-user distribution where:

- no accountable organisation has accepted the publisher role;
- vulnerability reporting and security maintenance are not operational;
- release authenticity, supported versions or withdrawal arrangements are undefined;
- privacy, accessibility and product complaint routes do not exist;
- continuity depends on a single personal account.

Stop before institutional supply where:

- no contracting legal entity exists;
- support, data roles, liability, insurance or exit arrangements are unresolved;
- the organisation cannot provide accurate procurement and AI disclosures.

Stop before healthcare, justice-sector or EU availability where:

- the responsible organisation has not completed the relevant legal and regulatory assessment;
- mandatory reporting, conformity or legal-representative arrangements are absent;
- the intended purpose or actual functionality exceeds the approved boundary.

## Official reference basis

This control was drafted with reference to:

- NCSC, *Software Security Code of Practice*: https://www.ncsc.gov.uk/section/software-security-code-of-practice/overview
- NCSC, *Software Security Code of Practice — Assurance Principles and Claims*: https://www.ncsc.gov.uk/guidance/software-security-code-of-practice-assurance-principles-claims
- NCSC, *Secure deployment and maintenance*: https://www.ncsc.gov.uk/collection/software-security-code-of-practice-implementation-guidance/secure-deployment-maintenance
- NCSC, *Communication with customers*: https://www.ncsc.gov.uk/collection/software-security-code-of-practice-implementation-guidance/theme-4-communication-with-customers
- ICO, *How to deal with data protection complaints*: https://ico.org.uk/for-organisations/how-to-deal-with-data-protection-complaints/
- GOV.UK, *PPN 017: Improving transparency of AI use in procurement*: https://www.gov.uk/government/publications/ppn-017-improving-transparency-of-ai-use-in-procurement
- Regulation (EU) 2024/2847, Cyber Resilience Act: https://eur-lex.europa.eu/eli/reg/2024/2847/2024-11-20/eng

These references do not appoint an organisation or determine every legal role. The applicable duties depend on the actual publisher, distribution model, processing, market, product functionality and deployment context.
