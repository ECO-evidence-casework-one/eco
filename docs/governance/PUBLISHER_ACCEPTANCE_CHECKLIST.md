# ECO publisher and steward acceptance checklist

**Purpose:** assess a proposed established organisation before ECO relies on it as official publisher or steward  
**Issue:** #17  
**Status:** a completed checklist is evidence for a decision; it does not itself appoint the organisation or approve a release

## 1. Candidate record

| Field | Required entry |
|---|---|
| Legal name | |
| Legal form | |
| Registration number, if applicable | |
| Jurisdiction | |
| Registered or formal service address | |
| Public website | |
| Governing body | |
| Authorised decision-maker | |
| Operational lead | |
| Security lead | |
| Accessibility lead | |
| Privacy/data-protection lead | |
| Release/signing lead | |
| Deputies and recovery contacts | |
| Date assessment opened | |
| Evidence review date | |
| Governing-body decision date | |
| Next review date | |

Do not include private personal addresses, identity documents, credentials or sensitive records in the public repository. Preserve confidential due-diligence evidence through an approved private organisational process.

## 2. Decision vocabulary

For each item record:

- **PASS** — adequate evidence exists;
- **HOLD** — evidence or correction is outstanding;
- **FAIL** — the candidate conflicts with a controlling ECO requirement;
- **NOT APPLICABLE** — a reasoned scope decision exists;
- **NOT ASSESSED** — review has not occurred.

A PASS must cite the evidence and its date. Silence, a verbal assurance or a website marketing statement is not sufficient for a high-risk item.

## 3. Immediate exclusion screen

Stop the assessment and record **FAIL** if the proposed organisation requires any of the following as a condition of official stewardship:

- proprietary relicensing of ECO or withholding complete corresponding source;
- a closed official fork;
- mandatory cloud evidence upload;
- mandatory user accounts, telemetry, advertising or analytics;
- routine publisher or developer access to users' case files;
- public submission of real evidence or private diagnostics for support;
- legal, medical, forensic, emergency, scoring or compliance claims beyond the approved intended purpose;
- real-data or institutional pilots before the active gates close;
- personal warranties, contracts, support duties, complaints duties or liability assigned to an individual contributor by default;
- dependence on one personal account, device, certificate or unfunded volunteer for official continuity;
- refusal to maintain truthful SBOM, licensing, provenance, signing and vulnerability records.

| Exclusion screen | Decision | Evidence or reason |
|---|---|---|
| No fatal red flag identified | | |

## 4. Legal identity and authority

| Requirement | Decision | Evidence |
|---|---|---|
| Established legal identity verified | | |
| Legal form and jurisdiction verified | | |
| Governing document or equivalent authority reviewed | | |
| Governing body has power to accept the ECO role | | |
| Authorised signatories identified | | |
| Organisation can contract in its own name | | |
| Formal notice/service route exists | | |
| Conflicts-of-interest process exists | | |
| Funding and financial-record capability is credible | | |
| Decision does not appoint the originating developer or another contributor as director, trustee or equivalent by default | | |

### Structure note

Record whether the candidate is an existing charity, nonprofit, foundation, public-interest institution, company, CIC, university, public body or another legal form.

Do not recommend creation of a new entity merely to complete this checklist. Any new-entity proposal requires a separate options, duties, funding and office-holder assessment.

## 5. Mission and product-boundary fit

| Requirement | Decision | Evidence |
|---|---|---|
| Written commitment to free and open-source operation | | |
| Written commitment to fully offline application operation | | |
| No required cloud, account, telemetry or advertising | | |
| No routine access to users' local evidence | | |
| Commitment to one-file Windows delivery objective | | |
| Commitment to low-spec Windows support | | |
| Acceptance of intended-purpose and prohibited-use controls | | |
| Acceptance of source/OCR/suggestion/confirmation/note separation | | |
| Acceptance that governance wording is not implementation proof | | |
| Acceptance that real evidence remains blocked until qualified | | |

## 6. Open-source, licence and supply-chain capability

| Requirement | Decision | Evidence |
|---|---|---|
| GPL obligations understood | | |
| Corresponding-source process exists | | |
| Dependency and model licence review capability exists | | |
| Actual-build component inventory capability exists | | |
| SBOM generation and reconciliation capability exists | | |
| Third-party notice process exists | | |
| Source/binary/build/signature identity can be reconciled | | |
| Independent forks will not be misrepresented as official ECO | | |
| Official name and endorsement policy can be operated | | |

## 7. Security and vulnerability response

| Requirement | Decision | Evidence |
|---|---|---|
| Monitored private vulnerability route exists or is funded | | |
| Supported-version policy exists | | |
| Severity and triage process exists | | |
| Confidential-report handling is defined | | |
| Fix, mitigation, advisory and withdrawal process exists | | |
| Significant-incident communication plan exists | | |
| Post-incident review process exists | | |
| Free/open-source vulnerability policy is documented | | |
| Response commitments do not depend solely on spare personal capacity | | |

### Required synthetic test

Send a synthetic private vulnerability report. Record:

| Test field | Result |
|---|---|
| Date and time | |
| Route used | |
| Acknowledgement received | |
| Triage performed | |
| Confidentiality preserved | |
| Escalation path followed | |
| Outcome and limitations | |

## 8. Release, signing and maintenance

| Requirement | Decision | Evidence |
|---|---|---|
| Release approval authority is defined | | |
| Author, reviewer and release approver separation is adequate | | |
| Exact source/build/manifest/SBOM/licence/signature reconciliation is supported | | |
| Signing key/certificate governance exists | | |
| Revocation and recovery process exists | | |
| Post-signing mutation is prohibited | | |
| Official download and verification route is defined | | |
| Update and rollback process supports offline users | | |
| Release withdrawal process exists | | |
| End-of-support policy exists | | |
| Vulnerable bundled model/runtime response exists | | |

## 9. Repository and continuity

| Requirement | Decision | Evidence |
|---|---|---|
| Protected-branch and required-review controls can be maintained | | |
| Routine direct-to-release-branch changes are prevented | | |
| Emergency/admin bypass is logged and reviewed | | |
| More than one protected recovery owner exists where supported | | |
| Recovery information is not held only on one device/account | | |
| Repository and organisation recovery has been tested | | |
| Signing continuity has been tested | | |
| Maintainer unavailability plan exists | | |
| Freeze, archive, transfer and withdrawal authority is defined | | |

### Required tabletop exercise

Test loss or unavailability of the primary maintainer and record whether the organisation can:

- secure the repository;
- preserve or withdraw official releases;
- communicate the situation;
- restore authorised access;
- prevent unauthorised signing or publication;
- continue the supported-version decision process.

## 10. Privacy and support-data roles

| Requirement | Decision | Evidence |
|---|---|---|
| Each actual data flow is mapped | | |
| Controller/processor/joint-controller/neither roles are assessed per activity | | |
| Support-data minimisation is defined | | |
| Users are not required to submit real evidence for ordinary support | | |
| Lawful purpose and access controls are defined where applicable | | |
| Retention and deletion rules are defined where applicable | | |
| Data-breach escalation is defined | | |
| Data-protection complaint route exists | | |
| Complaint acknowledgement/investigation/outcome duties are operational | | |
| Cloud, telemetry or account changes require fresh approval | | |

## 11. Accessibility, support and product complaints

| Requirement | Decision | Evidence |
|---|---|---|
| Monitored accessibility-barrier route exists | | |
| Accessible alternative communication is available | | |
| Reasonable-adjustment process is defined where applicable | | |
| Specialist accessibility testing can be funded | | |
| Representative disabled-user testing can be arranged | | |
| Automated scanning is not treated as complete conformance evidence | | |
| General support boundaries are deliverable and published | | |
| Product-complaint route exists | | |
| Harmful or misleading output escalation exists | | |
| Emergency/safeguarding wording does not claim unperformed action | | |

## 12. Legal, regulatory and claims capability

| Requirement | Decision | Evidence |
|---|---|---|
| Organisation accepts the controlling intended-purpose record | | |
| Public claims approval process exists | | |
| Unsupported compliance, security, accuracy and accessibility claims are controlled | | |
| Access to qualified UK legal advice is available when needed | | |
| Access to EU advice is available before EU supply | | |
| Medical-device boundary can be reassessed for actual features/claims | | |
| AI Act roles can be assessed for actual supply/deployment | | |
| Cyber Resilience Act roles can be assessed for actual activity | | |
| High-consequence and authority-side use triggers fresh review | | |
| Independent forks and official ECO identity are distinguished | | |

## 13. Institutional procurement and contracting

Complete only before institutional supply.

| Requirement | Decision | Evidence |
|---|---|---|
| Contracting legal entity identified | | |
| Authorised signatory identified | | |
| Product and support scope defined | | |
| Intended and prohibited uses recorded | | |
| Buyer-requested AI disclosures can be answered accurately | | |
| Data roles and data-processing terms are defined | | |
| Security and incident responsibilities are defined | | |
| Accessibility responsibilities are defined | | |
| Liability and warranty position is approved | | |
| Insurance decision is recorded | | |
| Exit, portability and termination arrangements exist | | |
| No individual contributor is asked to contract personally | | |

PPN 017 is treated as buyer-side policy for its stated in-scope central-government bodies and an optional approach elsewhere, not as universal supplier law.

## 14. Operational acceptance tests

| Test | Decision | Evidence |
|---|---|---|
| Synthetic vulnerability-report test | | |
| Release-withdrawal tabletop | | |
| Significant-incident communication tabletop | | |
| Privacy-complaint test | | |
| Accessibility-barrier report test | | |
| Harmful-output complaint test | | |
| Repository/account-loss recovery test | | |
| Primary-maintainer unavailability test | | |
| Signing revocation/recovery tabletop | | |
| Official-versus-independent-build identification test | | |

## 15. Governing-body acceptance record

The authorised governing body must record:

| Decision field | Entry |
|---|---|
| Duties accepted | |
| Duties excluded or deferred | |
| Conditions and dependencies | |
| Funding/resources approved | |
| Authorised role-holders | |
| Deputies/recovery owners | |
| Effective date | |
| Review date | |
| Withdrawal/termination process | |
| Resolution or decision reference | |

An informal email, expression of interest, staff discussion or technical collaboration is not sufficient.

## 16. Final control decision

| Matter | Decision | Reason |
|---|---|---|
| Candidate passes due diligence | | |
| Governing-body acceptance is valid | | |
| Operational tests pass | | |
| Publisher/steward role may be announced | | |
| Issue #17 may close | | |
| Ordinary-user release gate may be reconsidered | | |
| Institutional gate may be reconsidered | | |

Even after organisational acceptance, an official release remains separately blocked until the exact candidate satisfies every application, evidence-integrity, security, accessibility, provenance, licensing, signing and distribution gate.

## 17. Review statement

Record who prepared and reviewed the checklist, their relationship to the candidate organisation and the evidence they inspected.

Do not describe the review as independent unless the reviewer is meaningfully separate from both the candidate organisation's operational proposal and the ECO work being approved.