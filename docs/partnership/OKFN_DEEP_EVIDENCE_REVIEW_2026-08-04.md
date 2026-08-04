# Open Knowledge Foundation deep evidence review

**Issue:** #62  
**Related issue:** #17  
**Review date:** 4 August 2026  
**Recorded:** approximately 13:49 BST / 14:49 CEST  
**Canonical base reviewed:** `90825496cda639ef6be8c052a8ad9f2ebe2f1464`  
**Role tested:** possible full or near-full accountable ECO publisher/steward under the preferred accountable-steward-plus-specialists model  
**Result:** `HOLD` remains controlling  
**Outreach:** prohibited; no contact, shortlist or implied relationship

## 1. Plain-English conclusion

Open Knowledge Foundation has credible and unusually relevant organisational and technical experience. Current public evidence shows an established UK legal entity, a board and executive structure, current accounts filings, institutional work, open-source infrastructure, a maintained cross-platform desktop product, and local-first/offline AI experience.

That evidence is not sufficient to trust the organisation with official ECO responsibility.

The deeper review identified material governance, security, release, signing, complaints, financial-dependency and evidence-freshness gaps. None proves misconduct, insolvency, incapacity or incompatibility. Together, however, they prevent a positive full-steward decision and prevent shortlist or outreach approval.

The correct status remains `HOLD`.

## 2. Scope and method

This supplement deepens the earlier full/near-full assessment in:

- [`CANDIDATE_ASSESSMENT_OPEN_KNOWLEDGE_FOUNDATION_2026-08-04.md`](CANDIDATE_ASSESSMENT_OPEN_KNOWLEDGE_FOUNDATION_2026-08-04.md);
- [`CANDIDATE_ORGANISATION_RESEARCH_REGISTER.md`](CANDIDATE_ORGANISATION_RESEARCH_REGISTER.md);
- [`PUBLISHER_RESPONSIBILITY_MODEL_COMPARISON_2026-08-04.md`](PUBLISHER_RESPONSIBILITY_MODEL_COMPARISON_2026-08-04.md).

It tests current public evidence against the issue #17 responsibilities that one accountable official steward must retain even when specialists are used.

The review used public official registers, OKFN's own governance and policy pages, official Open Data Editor documentation and the public Open Data Editor GitHub repository.

This is project control-room research, not independent organisational due diligence. No private information, confidential records, personal contact details or representations from OKFN were obtained.

## 3. Decision changes

| Matter | Earlier position | Deep-review position | Effect |
|---|---|---|---|
| Overall full/near-full status | `HOLD` | `HOLD` | unchanged |
| Active UK legal identity | supported | supported | no change |
| Current accounts filing | supported with document-review limitation | 2025 accounts filing verified at Companies House; financial figures not independently extracted here | strengthened but limited |
| Current corporate filings | not fully tested | confirmation statement shown overdue on 4 August 2026 | new adverse governance fact |
| Registered security interests | not tested | one outstanding fixed charge shown at Companies House | new due-diligence dependency; not proof of distress |
| Private vulnerability route | unproven | public ODE Security page says no security policy detected | HOLD strengthened |
| Desktop release activity | partly supported | current releases and GitHub-verified release commit evidenced | strengthened, but not Windows binary-signing proof |
| Authenticode and key governance | unproven | remains unproven | HOLD unchanged |
| Privacy evidence | 2018 policy recorded as stale | still materially stale and organisation-wide rather than ECO-product specific | HOLD strengthened |
| NGOsource equivalency | wording expired December 2025 | no current renewal established from reviewed public pages | HOLD unchanged |
| Complaints and support separation | unproven | conduct process exists, but no complete product/security/privacy/accessibility complaint system found | HOLD unchanged |
| Full financial review | incomplete | Companies House filing status and registered charge verified; accounts PDF not inspected through the available PDF workflow | limitation preserved |

## 4. Current legal identity and filing evidence

### 4.1 Identity

Companies House records Open Knowledge Foundation, company number `05133759`, as an active private company limited by guarantee without share capital. It was incorporated on 20 May 2004.

This supports legal identity and continuity. It does not establish willingness, capacity, tax status, insurance, product liability, regulatory competence or authority to accept ECO.

### 4.2 Accounts

Companies House filing history records accounts for the year ended 31 December 2025 as filed on 9 May 2026.

This confirms a current accounts filing exists. The underlying accounts PDF could not be reliably rendered and inspected through the available PDF-review workflow during this review. No financial totals, reserves, going-concern statements, auditor conclusions or risk disclosures are therefore asserted from that PDF.

Financial capacity remains `HOLD` pending direct, auditable review of the filed accounts and any relevant later records.

### 4.3 Confirmation statement

At the review date, the Companies House overview displayed the next confirmation statement as dated 20 May 2026, due by 3 June 2026, and marked it overdue. The last statement shown was dated 20 May 2025.

This is a current corporate-compliance warning that must be resolved or explained before any shortlist decision. It is not evidence that OKFN is insolvent, fraudulent, inactive or unable to perform work.

### 4.4 Registered charge

Companies House records one charge, created on 19 March 2025 and delivered on 20 March 2025, in favour of Barclays Security Trustee Limited. The charge page describes it as outstanding and states that it contains a fixed charge.

This is a material financial and asset-control due-diligence fact. It may affect assets, lender rights, contractual approvals or continuity depending on the instrument's terms.

The public charge record alone does not prove financial distress, default or inability to act. The charge instrument, secured obligations and effect on any proposed ECO assets would require qualified review before reliance.

## 5. Governance and evidence-freshness findings

### 5.1 Board and executive structure

OKFN's governance pages identify a board responsible for financial and legal probity, with the chief executive accountable to that board. Public officer records and organisational pages support a multi-person governance structure rather than a single-person project.

No public board resolution, delegated-authority record or governing-body acceptance exists concerning ECO. General governance capacity is not organisational acceptance.

### 5.2 NGOsource wording

OKFN's governance page states that an NGOsource Equivalency Determination was valid through December 2025.

At 4 August 2026, that statement is stale for current equivalency unless a renewal exists elsewhere. The review does not treat current US public-charity equivalency as established.

This does not alter OKFN's verified England-and-Wales company status.

### 5.3 Privacy policy

The public privacy policy states an effective date of 25 May 2018. It describes website and service tracking, cookies, uploaded datasets, security practices and data-rights routes.

It is relevant organisational evidence but is too old and too general to establish the required ECO distribution and support data-role model. It does not prove:

- no collection of case material;
- no diagnostic evidence custody;
- current support-data retention and deletion;
- current breach and complaint operations;
- separation of product, privacy, security and accessibility complaints;
- offline application data flows;
- controls for specialist suppliers;
- a current controller/processor/neither analysis for ECO.

A current, deployment-specific data-flow and support-data assessment remains mandatory.

## 6. Open Data Editor technical and release evidence

### 6.1 Relevant positive evidence

Official documentation and the public repository support that Open Data Editor is:

- free and open-source desktop software;
- distributed for Windows, macOS and Linux;
- designed around local-first/offline operation;
- capable of running AI operations locally rather than sending data to a cloud model;
- maintained through public source, issues, pull requests, releases and documentation;
- subject to user research and iterative accessibility/installation work.

This is materially relevant to ECO's desktop, offline, FOSS and privacy direction.

### 6.2 Separate-model limitation

ODE's local AI documentation says users download a model separately. That does not meet ECO's controlling objective of one self-contained Windows executable containing all required approved components.

ODE experience supports general local-AI capability. It does not prove ECO's packaging, low-spec performance, model-licence, supply-chain or one-file release requirements.

### 6.3 Public security-policy gap

The public GitHub Security page for the ODE repository states that no security policy is detected and that no `SECURITY.md` file has been published. The page also showed no published security advisories at review time.

This does not prove that OKFN lacks internal security practices or has ignored vulnerabilities. It does mean the public evidence does not satisfy ECO's requirement for a monitored private vulnerability-reporting route, supported-version policy, triage process, coordinated disclosure, advisories and withdrawal procedures.

A public issue tracker or general email route is not a sufficient confidential vulnerability route for ECO.

### 6.4 Release-signature distinction

The ODE release page records a GitHub-verified signature for the source/release commit associated with the latest reviewed release.

That is useful source-control provenance. It is not proof that the Windows `.exe` is Authenticode-signed, that a trusted certificate chain exists, or that signing keys, certificate requests, revocation, recovery and separation of duties are governed.

The following remain unproven:

- Windows Authenticode status;
- identity of the binary signer;
- certificate custody and recovery;
- revocation procedure;
- exact source-to-binary reconciliation;
- final-artifact hashes;
- actual-build SBOM;
- licence and notice reconciliation;
- reproducible or independently repeatable build evidence;
- prohibition on post-signing mutation;
- official withdrawal and supersession process.

A green GitHub verification badge must not be described as Windows executable-signing evidence.

### 6.5 Support and lifecycle

Public releases, website downloads, documentation, email feedback and GitHub issues demonstrate an active product and feedback surface.

They do not establish:

- which versions are supported;
- security-fix periods;
- end-of-support notices;
- offline update and rollback instructions;
- release withdrawal authority;
- urgent user-warning routes;
- ordinary-user support expectations;
- continuity when maintainers or suppliers are unavailable.

These are mandatory full-steward gaps.

## 7. Complaints and harmful-output controls

OKFN publishes a Code of Conduct with a reporting email, anonymous form, investigation, confidentiality and recusal provisions.

That is positive governance evidence for community conduct. It is not a complete ECO product-complaint system.

No reviewed public record proved separate monitored routes and escalation for:

- security vulnerabilities;
- privacy and data-protection complaints;
- accessibility barriers and reasonable adjustments;
- harmful or materially misleading generated output;
- evidence-provenance failures;
- data loss, corruption or unintended disclosure;
- general product support;
- emergency withdrawal or user warning.

A future accountable steward may operate one public front door, but internally the categories, confidentiality, ownership and escalation must be separated.

## 8. Issue #17 duty-gap matrix

| Required accountable-steward domain | Current public support | Remaining decision |
|---|---|---|
| Legal identity | active UK company and governance structure verified | `SUPPORTED` only for identity |
| Governing authority | board and CEO structure identified | `HOLD`: no ECO acceptance, authority allocation or resolution |
| Financial records | current 2025 accounts filing verified | `HOLD`: accounts substance, reserves, funding, charge terms and capacity unverified |
| FOSS commitment | strong organisational and product evidence | `SUPPORTED IN GENERAL`; ECO acceptance absent |
| Offline/local operation | strong ODE evidence | `PARTLY SUPPORTED`; ECO no-custody and one-file commitment absent |
| Private vulnerability route | no repository security policy detected | `HOLD` |
| Security incident response | technical capability plausible | `HOLD`: no current public operating evidence |
| Supported versions and maintenance | active releases exist | `HOLD`: policy and commitments absent |
| Release withdrawal | no adequate public evidence found | `HOLD` |
| Authenticode and signing governance | GitHub commit verification only | `HOLD` |
| Actual-build provenance and SBOM | no complete public evidence found | `HOLD` |
| Repository and asset continuity | long-lived organisation and repository activity | `HOLD`: recovery ownership and transfer controls absent |
| Privacy and support-data roles | old general privacy policy exists | `HOLD` |
| Accessibility | UX and accessibility work evidenced | `PARTLY SUPPORTED`; ECO testing and complaint route absent |
| Product and harmful-output complaints | conduct and feedback routes exist | `HOLD` |
| Institutional contracting | general services capability evidenced | `SUPPORTED IN GENERAL`; ECO terms, liability and signatory absent |
| Insurance and product liability | no adequate public evidence found | `HOLD` |
| UK/EU regulatory ownership | policy/legal capability plausible | `HOLD`: no ECO role acceptance or classification process evidenced |
| Specialist governance | contracts and project management plausible | `HOLD`: required ECO supplier-control model not accepted or tested |
| No individual-contributor fallback | established organisation reduces risk | `HOLD`: no written acceptance preventing fallback |
| Operational tabletop tests | none concerning ECO | `HOLD` |

## 9. Fatal-exclusion reassessment

No fatal exclusion is established from the reviewed public evidence.

The deeper evidence does not show that OKFN would require:

- proprietary relicensing;
- a closed official fork;
- mandatory ECO cloud evidence upload;
- mandatory telemetry or advertising;
- routine access to users' case files;
- personal liability for an individual contributor;
- prohibited clinical, legal, forensic or scoring functions.

Equally, the public evidence does not show that OKFN would accept ECO's contrary requirements in writing.

The correct decision is therefore `HOLD`, not `UNSUITABLE` and not `POTENTIAL FIT` for the full role.

## 10. Financial and operational cautions

The following must be handled without overstatement:

- an overdue confirmation statement is a filing-compliance fact, not proof of organisational failure;
- an outstanding fixed charge is an asset/finance dependency, not proof of default or insolvency;
- a current accounts filing is not proof of adequate unrestricted reserves or capacity;
- twenty years of operation is not proof of future continuity;
- a broad team page is not proof that staff are available for ECO;
- service and government work is not proof of conflict, but it requires role, data and authority-use separation;
- Digital Public Good recognition is not legal, security, accessibility or release approval for ECO.

## 11. Evidence limitations

This review did not:

- contact OKFN;
- obtain non-public policies, contracts or insurance records;
- inspect private security processes;
- verify Windows executable signatures directly;
- download or execute ODE binaries;
- inspect the complete 2025 accounts PDF through the required PDF screenshot workflow;
- obtain the charge instrument and secured-obligation details;
- test complaint, vulnerability, withdrawal or recovery routes;
- assess staff availability or governing-body willingness;
- obtain legal advice on OKFN's status or the registered charge.

Absence of public evidence is recorded as a gap, not as proof that the capability does not exist.

## 12. Provisional decision

**Status:** `HOLD`

**Reasoned conclusion:**

OKFN remains the strongest full or near-full public-source candidate examined so far. It has meaningful evidence of legal identity, governance, FOSS stewardship, technical delivery, cross-platform desktop releases, local-first software and institutional operations.

The deep review also shows that the essential issue #17 duties are not publicly established. Current corporate filing and charge facts require due diligence; public privacy evidence is stale; the ODE repository publishes no security policy; GitHub commit verification is not Windows binary signing; and complete release, withdrawal, support, complaints, insurance, liability, regulatory and no-custody controls remain unproven.

A shortlist or outreach decision would therefore be premature.

## 13. Permitted next action

Permitted under issue #62:

1. preserve this supplement and the unchanged `HOLD` decision;
2. verify whether the overdue confirmation statement is later filed or status changes;
3. inspect current accounts and charge documents through a valid auditable PDF process if available;
4. inspect public ODE release assets, hashes, SBOMs, signing and support records without downloading or executing prohibited ECO artifacts;
5. compare one further full-role organisation using the same issue #17 matrix;
6. deepen the limited-role records where useful;
7. return the evidence for control review before shortlist creation.

## 14. Prohibited next action

This review does not authorise:

- contact with OKFN;
- an expression-of-interest message;
- a quotation, grant, fiscal-hosting or sponsorship application;
- naming OKFN as preferred, shortlisted or recommended;
- an accessibility, security or technical booking;
- transfer of source, repository, brand, domain or signing authority;
- disclosure of real evidence or private diagnostics;
- a real-data, government, healthcare, justice or institutional pilot;
- formation of a company, CIC, charity or other entity;
- signing, executable publication, release or deployment;
- any statement that OKFN approves, supports or knows about ECO.

## 15. Controlling stop point

The full-steward research lane remains stopped before shortlist and outreach.

One accountable official steward remains mandatory. Specialists may assist only under the controlling responsibility model and cannot replace the accountable organisation.

No candidate status or release gate changes through this supplement.

## 16. Public source register

| ID | Source | Record | Access date | Proposition | Limitation |
|---|---|---|---|---|---|
| D1 | Companies House | `OPEN KNOWLEDGE FOUNDATION overview` — `https://find-and-update.company-information.service.gov.uk/company/05133759` | 2026-08-04 | identity, active status, accounts date, overdue confirmation statement | registry facts only |
| D2 | Companies House | `Filing history` — `https://find-and-update.company-information.service.gov.uk/company/05133759/filing-history` | 2026-08-04 | 2025 accounts filing date and historical filing events | does not replace accounts review |
| D3 | Companies House | `Charges` — `https://find-and-update.company-information.service.gov.uk/company/05133759/charges` | 2026-08-04 | one outstanding registered charge and charge date | instrument and secured obligations not analysed |
| D4 | Companies House | `Officers` — `https://find-and-update.company-information.service.gov.uk/company/05133759/officers` | 2026-08-04 | multi-person officer record | does not prove ECO authority |
| D5 | Open Knowledge Foundation | `Governance` — `https://okfn.org/en/who-we-are/governance/` | 2026-08-04 | board/CEO model, reports/accounts and stale NGOsource wording | self-published organisational page |
| D6 | Open Knowledge Foundation | `Annual Report 2025` — `https://blog.okfn.org/annual-report-2025/` | 2026-08-04 | current activities, strategy and ODE/local-AI work | self-reported; financial PDF not analysed |
| D7 | Open Knowledge Foundation | `Privacy policy` — `https://okfn.org/privacy-policy/` | 2026-08-04 | 2018 effective date, general website/service data practices | stale and not ECO-specific |
| D8 | Open Knowledge Foundation | `Code of Conduct` — `https://okfn.org/en/code-of-conduct/` | 2026-08-04 | conduct reporting, confidentiality, investigation and recusal | not product/security/privacy/accessibility complaint system |
| D9 | Open Knowledge Foundation | `IP policy` — `https://okfn.org/en/ip-policy/` | 2026-08-04 | open content/source and organisational IP route | not ECO asset/signing policy |
| D10 | Open Data Editor documentation | `Responsible AI integration` — `https://opendataeditor.okfn.org/introduction/responsible-ai-integration` | 2026-08-04 | local AI/no-cloud operation; separate model download | does not meet ECO one-file objective |
| D11 | Open Data Editor documentation | `Downloading ODE` — `https://opendataeditor.okfn.org/user-guide/downloading-ode` | 2026-08-04 | Windows/macOS/Linux distribution routes | does not establish signatures, SBOM or support lifecycle |
| D12 | Open Knowledge Foundation / GitHub | `Open Data Editor repository` — `https://github.com/okfn/opendataeditor` | 2026-08-04 | public source, issues, pull requests, actions and releases | public repository evidence only |
| D13 | Open Knowledge Foundation / GitHub | `Open Data Editor security` — `https://github.com/okfn/opendataeditor/security` | 2026-08-04 | no security policy detected; no published advisories shown | absence of public policy is not proof of no internal process |
| D14 | Open Knowledge Foundation / GitHub | `Open Data Editor releases` — `https://github.com/okfn/opendataeditor/releases` | 2026-08-04 | current releases and GitHub-verified release commit | not Authenticode or exact binary provenance proof |
| D15 | Open Knowledge Foundation | `Open Data Editor user testing` — `https://blog.okfn.org/2024/04/16/open-data-editor-user-testing/` | 2026-08-04 | user research and installation/accessibility findings | historical and product-specific |

---

**No-interest statement:** OKFN has not been contacted. This document does not imply knowledge, interest, consent, endorsement, capacity, acceptance or appointment.
