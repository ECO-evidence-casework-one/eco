# ECO full-role candidate assessment — mySociety

**Issue:** #62  
**Related issue:** #17  
**Assessment date:** 4 August 2026  
**Recorded:** approximately 14:25 BST / 15:25 CEST  
**Canonical base reviewed:** `3c7c69586cac195d146188e6b914db12f6391815`  
**Legal organisation assessed:** mySociety, company number `03277032`, charity number `1076346`  
**Related organisation considered:** SocietyWorks Ltd, company number `05798215`, wholly owned trading subsidiary  
**Role assessed:** possible full or near-full accountable ECO publisher/steward under the controlling accountable-steward-plus-specialists model  
**Research relationship:** project control-room review using public sources; not organisationally independent due diligence  
**Overall provisional status:** `HOLD`  
**Outreach:** prohibited; neither mySociety nor SocietyWorks has been contacted or approved for contact

## 1. Plain-English conclusion

mySociety has substantial public evidence of legal continuity, charitable governance, open-source software maintenance, operational security, public-service contracting, incident monitoring, support, accessibility work and privacy/data-role experience.

In several operational areas, the public evidence is stronger than the evidence found for Open Knowledge Foundation. That does not make mySociety suitable or preferred.

The main full-role obstacle is structural and product-specific:

- mySociety and SocietyWorks are separate legal organisations with separate boards and an arm's-length relationship;
- the reviewed services are hosted web and public-sector systems that use accounts, client support systems, monitoring, integrations and personal-data processing;
- ECO requires a fully offline, no-account, no-telemetry, no-user-evidence-custody Windows application;
- no reviewed public evidence establishes native Windows release, Authenticode, one-file packaging, exact executable provenance, offline updates/rollback, product withdrawal, harmful-output controls or acceptance of ECO's legal and regulatory boundaries.

No fatal exclusion is proven. No full-role qualification is proven.

The correct result is `HOLD`.

## 2. Relationship to the earlier mySociety record

The first candidate batch assessed mySociety only as a possible civic-technology and public-interest software collaborator and recorded `HOLD`.

This document is a separate role assessment. It asks whether **mySociety itself** could become the one official accountable ECO publisher/steward required by issue #17, while using SocietyWorks or other specialists under controlled agreements.

The two findings must not be merged:

- limited collaborator role: `HOLD`;
- full/near-full accountable steward role: `HOLD`.

Neither finding implies interest, availability, ranking, shortlist status or contact approval.

## 3. Legal identity and organisational structure

### 3.1 mySociety

Companies House records mySociety as:

- company number `03277032`;
- active;
- a private company limited by guarantee without share capital using the `Limited` exemption;
- incorporated on 12 November 1996;
- registered at 483 Green Lanes, London, N13 4BS;
- operating under SIC 62090, other information technology service activities;
- with accounts made up to 31 March 2026;
- with a confirmation statement dated 22 October 2025 and the next statement due in November 2026.

The Charity Commission records:

- charity number `1076346`;
- organisation type: charitable company;
- reporting up to date and on time;
- Gift Aid recognition;
- a multi-policy governance framework including complaints, conflicts, reserves, financial controls, risk, safeguarding and serious-incident reporting.

This is strong legal-identity and governance evidence. It is not an ECO acceptance decision.

### 3.2 SocietyWorks Ltd

Companies House records SocietyWorks Ltd as:

- company number `05798215`;
- active;
- a private limited company;
- incorporated on 27 April 2006;
- registered at the same Green Lanes address;
- with accounts made up to 31 March 2026;
- with a confirmation statement dated 1 April 2026.

mySociety describes SocietyWorks as its wholly owned trading subsidiary. SocietyWorks delivers services for local authorities and other clients, data APIs and custom development. mySociety staff may work across charitable and commercial projects, with staff time charged to SocietyWorks.

### 3.3 Board and accountability split

mySociety states that:

- the charity's trustees oversee all group activities;
- SocietyWorks has its own board;
- membership overlaps but is never identical;
- the organisations maintain an arm's-length relationship;
- each board meets approximately quarterly.

This creates a material issue #17 question. An ECO proposal must name one legal organisation as accountable publisher. It cannot rely on an informal statement that “the group” is responsible.

Before a positive decision, a written model would need to establish:

- whether mySociety or SocietyWorks is the official publisher;
- which board has authority to accept the role;
- which legal entity contracts, signs, insures, receives complaints and owns regulatory decisions;
- how staff shared across the two entities are instructed and supervised;
- how conflicts, incidents and financial flows are escalated;
- whether one entity is a contracted specialist to the other;
- how responsibility survives subsidiary, staffing or board changes.

## 4. Mission and intended-purpose fit

mySociety's charitable objects and public mission concern citizenship, civic responsibility, public-authority information, community development, democratic participation, research and effective use of information technology.

Its ethical statement says SocietyWorks products should not impinge on human agency or discriminate.

This is materially aligned with ECO's user-control and non-scoring principles.

The public mission does not specifically establish responsibility for personal evidence and casework software involving benefits, health, housing, complaints or legal matters. The governing body would need to decide that ECO falls within its powers, mission, risk appetite and resources.

Public-sector and authority-facing work is not a fatal conflict. It creates a need for strict separation from:

- authority-side eligibility or adverse decisions;
- credibility or risk scoring;
- surveillance or profiling;
- clinical, legal, forensic or emergency claims;
- institutional access to users' local case files;
- deployment models that weaken user control.

## 5. Open-source and software stewardship evidence

mySociety has a long-running public GitHub organisation with hundreds of repositories. Prominent projects include:

- Alaveteli;
- FixMyStreet;
- TheyWorkForYou;
- data, API, parser and integration projects.

Alaveteli is publicly described as an international open-source platform. Its repository shows extensive commit history, contribution processes, automated testing, issue and pull-request workflows, and a stable release branch.

This supports:

- long-term FOSS maintenance;
- public development and review;
- project reuse in multiple jurisdictions;
- contributor and release processes;
- complex service operation and upgrades.

It does not establish:

- willingness to accept ECO under its current licence and boundaries;
- one-file Windows packaging;
- native Windows desktop engineering;
- model/runtime licensing and bundling;
- exact executable SBOM, notices, hashes and signatures;
- offline update and rollback;
- formal official-versus-fork identity controls for ECO;
- Authenticode certificate/key governance.

## 6. Security and operational capability

### 6.1 Positive hosted-service evidence

SocietyWorks publicly states that it is Cyber Essentials Plus certified. Its secure-hosting material describes:

- dedicated and cloud hosting;
- automated testing, deployment and monitoring;
- least-privilege access;
- two-factor authentication for administrator accounts;
- firewalls and routine patching;
- encrypted backups;
- UK and Ireland data centres whose operators hold ISO 27001 certification;
- multi-location redundancy;
- daily backup checking and test restores;
- disaster recovery and temporary cloud recovery;
- redundant source-code storage.

SocietyWorks also publishes:

- severity levels;
- monitoring and escalation;
- response targets;
- maintenance-notice rules;
- a public status page with scheduled maintenance and resolved incidents.

This is strong evidence of real operational discipline for hosted services.

### 6.2 Private vulnerability gap

At the assessment date, the GitHub Security pages for the public Alaveteli and FixMyStreet repositories stated that no `SECURITY.md` policy was detected and showed no published security advisories.

This does not prove that mySociety lacks internal security reporting, triage or remediation. It does mean the reviewed public evidence does not establish an ECO-compliant confidential vulnerability route covering:

- supported versions;
- private submission;
- acknowledgement and confidentiality;
- severity and triage;
- fixes and mitigations;
- advisories;
- release withdrawal;
- user warnings;
- post-incident records.

Client Freshdesk accounts and general service monitoring are not a substitute for a public private-vulnerability route for an anonymously distributed desktop product.

### 6.3 Hosted controls are not desktop controls

Cloud hosting, server backups and service monitoring do not prove:

- safe offline desktop storage;
- native Windows sandboxing and permissions;
- parser/model/runtime containment;
- installer or one-file executable integrity;
- Authenticode and SmartScreen reputation;
- local diagnostic privacy;
- offline recovery and export;
- revocation or withdrawal of a compromised executable;
- low-spec Windows performance.

Those controls would need separate evidence and specialist capability.

## 7. Privacy, accounts, telemetry and evidence custody

### 7.1 Existing data-processing model

mySociety's public privacy page states that running its websites means user data will be in its care for a period, that services have specific privacy policies, and that support emails may be accessed and forwarded internally.

The page also records use of services such as Mailchimp, Google tools, cookies, Facebook elements and Google Analytics. It was last updated in March 2021.

SocietyWorks' privacy material describes:

- collection of business contact and support information;
- Mailchimp and Google tools;
- direct marketing;
- indefinite retention of contact-form messages for continuity in some circumstances.

SocietyWorks' data-sharing agreement states that it and clients may act as joint controllers for public reports and requests. It describes staff access, moderation, client administrative access, rights handling, breach notification, security and retention responsibilities.

This is evidence of explicit data-role experience. It is also a material mismatch with ECO's default model.

### 7.2 ECO separation required

The current operating model does not prove that mySociety would require ECO to use cloud services, telemetry, accounts or evidence custody. No fatal exclusion is inferred.

A full-role decision would require binding evidence that ECO remains separate:

- application operation fully offline;
- no mandatory account;
- no telemetry, analytics, advertising or hidden online service;
- no routine publisher access to case files;
- no real evidence required for ordinary support;
- no automatic diagnostics containing matter names, filenames or contents;
- public documentation and support that do not encourage evidence uploads;
- product-specific privacy and data-role matrix;
- fresh approval before any online or support-data feature.

The March 2021 mySociety privacy page is materially old for this future product assessment. Existing service-specific policies cannot substitute for an ECO-specific current assessment.

## 8. Accessibility evidence

SocietyWorks states that its web solutions are WCAG 2.2 AA compliant. Its public accessibility material describes:

- responsive interfaces;
- keyboard and browser controls;
- screen-reader testing of key journeys;
- contrast controls;
- simple step-by-step workflows;
- operation without JavaScript in relevant services;
- routine automated testing;
- recommendations for independent accessibility audits.

This is positive evidence of accessibility awareness and web-service practice.

It is not proof of ECO accessibility because ECO is intended as a native or self-contained Windows desktop application. Public evidence does not establish:

- Windows UI Automation and assistive-technology behaviour;
- high contrast, magnification and reflow in the final shell;
- keyboard-only operation of every control;
- cognitive accessibility for complex evidence workflows;
- screen-reader behaviour in document previews and AI output;
- representative disabled-user testing;
- an ECO accessibility-barrier and reasonable-adjustment route;
- truthful conformance reporting for the exact executable.

A public WCAG claim must not be copied across to ECO without objective product-specific evidence.

## 9. Support, complaints and incident response

### 9.1 Positive evidence

Public sources support:

- a team containing operations, finance, HR/support, developers, technical operations, delivery, account management and SocietyWorks leadership;
- a Charity Commission policy register containing complaints handling, complaints procedures, risk, safeguarding and serious-incident reporting;
- client service levels and escalation;
- public status reporting;
- data-protection rights and complaint information;
- ethical-use principles.

### 9.2 Remaining ECO gaps

The reviewed public material does not establish one monitored ECO front door with internally separated routes for:

- private vulnerabilities;
- privacy complaints;
- accessibility barriers and reasonable adjustments;
- materially misleading or harmful generated output;
- evidence-provenance and source-link failures;
- data loss, corruption or disclosure;
- general ordinary-user support;
- release withdrawal and urgent warnings.

Freshdesk client accounts and contractual SLAs are designed for institutional customers. They are not evidence of a sustainable anonymous end-user support model for a free offline desktop application.

## 10. Release, signing and supply-chain governance

The reviewed public evidence does not establish that mySociety or SocietyWorks currently publishes a native Windows desktop application comparable to ECO.

No adequate evidence was found for:

- Authenticode signing;
- certificate ownership and authorised signers;
- key custody, revocation and recovery;
- source/build/manifest/SBOM/licence/hash/signature reconciliation;
- reproducible or independently repeatable Windows builds;
- no post-signing mutation;
- supported desktop versions;
- offline update and rollback;
- emergency release withdrawal;
- official download and verification instructions;
- bundled model/runtime/parser response;
- low-spec Windows qualification.

This is a material full-role gap. It is not evidence that the organisation could not acquire or contract the capability.

## 11. Financial, contractual and continuity evidence

### 11.1 Positive evidence

Public registers show:

- long-established active legal organisations;
- current company accounts periods;
- current confirmation statements;
- Charity Commission reporting on time;
- a multi-policy governance framework;
- a trading subsidiary serving institutional clients;
- commercial support, hosting and development arrangements;
- multi-person boards and staff;
- public operational status and continuity practices.

The Charity Commission's financial summary for the year ending 31 March 2025 records income of approximately £2.63 million, including donations/legacies and trading income.

These facts support organisational substance. They do not prove unrestricted funds, spare capacity or willingness to fund ECO.

### 11.2 SocietyWorks charge record

Companies House records one SocietyWorks charge as technically outstanding but also states that all property or undertaking has been released from it. The person entitled is the parent organisation under its former name.

This appears to be an intra-group historical security arrangement rather than external lender evidence, but the legal effect should not be assumed without reviewing the instrument.

It is not recorded as evidence of default, distress or inability.

### 11.3 Remaining full-role evidence

A positive ECO decision would require:

- governing-body authority;
- approved budget and staffing;
- contracting legal entity;
- authorised signatories;
- insurance and product-liability decision;
- warranties and indemnity limits;
- repository, domain, brand and signing ownership;
- deputies and recovery owners;
- supplier replacement and termination process;
- continuity if the trading subsidiary, key staff or funding stream changes;
- protection against duties falling to an individual contributor.

## 12. Regulatory and high-consequence boundaries

mySociety and SocietyWorks have substantial public-sector, transparency and civic-participation experience. SocietyWorks' ethical statement opposes restrictions on human agency and discrimination.

This is relevant positive evidence, but not ECO regulatory qualification.

The organisation would need formally to own and fund feature-, market- and deployment-specific decisions concerning:

- intended purpose and public claims;
- UK medical-device boundary where features change;
- reserved legal services and litigation boundaries;
- forensic and justice-sector use;
- benefits, housing and other authority-side high-consequence decisions;
- UK product safety, consumer and product-liability duties;
- EU AI Act roles;
- EU Cyber Resilience Act roles;
- accessibility obligations;
- security and mandatory reporting;
- independent forks versus official ECO identity.

Existing public-sector work cannot be treated as evidence that these ECO classifications have been assessed or accepted.

## 13. Fatal-exclusion screen

| Fatal exclusion | Decision | Evidence and limitation |
|---|---|---|
| Proprietary relicensing or closed official fork required | `NO EVIDENCE OF EXCLUSION` | Strong public FOSS history; no ECO commitment exists |
| Mandatory cloud processing of ECO evidence | `NOT PROVEN` | Existing operating model is hosted/cloud, but no evidence it would require ECO to change |
| Mandatory accounts, telemetry, analytics or advertising | `NOT PROVEN / MATERIAL RISK` | Existing services use accounts, cookies, analytics and third parties; binding ECO separation absent |
| Routine access to users' case files | `NOT PROVEN / MATERIAL RISK` | Existing services process and moderate personal data; ECO no-custody commitment absent |
| Premature real-data pilot | `UNKNOWN` | Institutional service and user-research model exists; synthetic-only ECO commitment absent |
| Unsupported professional/compliance claims | `NO FATAL EVIDENCE` | Public claims exist but ECO claims controls are unaccepted |
| Prohibited medical/legal/forensic/scoring/decision use | `NO FATAL EVIDENCE / REQUIRES SEPARATION` | Ethical statement supports agency and non-discrimination; authority-side role boundaries unaccepted |
| Individual contributor must accept duties/liability | `NO EVIDENCE OF EXCLUSION` | Established entities exist; no ECO allocation exists |
| Single-person continuity dependency | `NO EVIDENCE OF EXCLUSION AT GROUP LEVEL` | Boards, staff and operational controls exist |
| Cannot operate security/complaint/withdrawal/signing duties | `UNKNOWN / MATERIAL` | Strong hosted operations; desktop vulnerability, signing and withdrawal evidence absent |
| No suitable legal identity | `NO EVIDENCE OF EXCLUSION` | Two active legal entities exist; exact accountable one unresolved |
| Governance incompatible with ECO boundaries | `UNKNOWN / MATERIAL` | Strong mission alignment, but hosted/public-sector/data-custody model requires binding ring-fencing |

**Fatal-exclusion result:** no fatal exclusion is established. Material full-role evidence remains incomplete. Status: `HOLD`.

## 14. Issue #17 duty-gap matrix

| Required domain | Public support | Decision |
|---|---|---|
| Active legal identity | strong, current company and charity evidence | `SUPPORTED` |
| One accountable legal publisher | two-entity group structure; no ECO allocation | `HOLD` |
| Governing authority | boards and trustees evidenced | `HOLD`: no ECO resolution or delegation |
| Mission/public-interest fit | strong civic and user-agency alignment | `SUPPORTED IN GENERAL` |
| FOSS stewardship | extensive public repositories and long-term maintenance | `STRONGLY SUPPORTED IN GENERAL` |
| Fully offline/no-custody compatibility | not established; current model is hosted and data-processing | `HOLD` |
| Native Windows/one-file delivery | not established | `HOLD` |
| Private vulnerability route | major repositories show no detected policy; hosted security controls exist | `HOLD` |
| Security incident operations | strong hosted-service evidence | `PARTLY SUPPORTED`; desktop model absent |
| Supported versions/maintenance | contractual hosted SLA exists | `HOLD` for desktop lifecycle |
| Release withdrawal | not established | `HOLD` |
| Authenticode/key governance | not established | `HOLD` |
| Build provenance/SBOM/licensing | FOSS processes exist; exact desktop artifact control unproven | `HOLD` |
| Privacy/data-role capability | strong joint-controller and policy experience | `SUPPORTED IN GENERAL`; ECO no-custody model unaccepted |
| Accessibility | strong web-service evidence | `PARTLY SUPPORTED`; Windows/ECO evidence absent |
| Complaints/support | policies, staff and client SLA evidenced | `PARTLY SUPPORTED`; anonymous ECO routes absent |
| Repository/organisational continuity | long-lived entity, team and redundancy evidence | `PARTLY SUPPORTED`; ECO recovery allocation absent |
| Financial sustainability | current filings, trading subsidiary and material income | `PARTLY SUPPORTED`; capacity/budget/insurance absent |
| Contracts/institutional capability | strong public evidence | `SUPPORTED IN GENERAL` |
| Liability/insurance | not adequately established | `HOLD` |
| Legal/regulatory ownership | policy and public-sector experience plausible | `HOLD`: no ECO acceptance/classification process |
| Specialist governance | group/subsidiary and client-contract experience | `PARTLY SUPPORTED`; controlling ECO model unaccepted |
| Operational tabletop tests | none for ECO | `HOLD` |
| No contributor fallback | organisations exist | `HOLD`: written protection and funding absent |

## 15. Comparison with Open Knowledge Foundation

This is a capability comparison, not a ranking or shortlist.

### mySociety appears publicly stronger in:

- current UK charity reporting;
- explicit complaints, risk and serious-incident policy listings;
- Cyber Essentials Plus evidence through SocietyWorks;
- hosted-service access control, backup, recovery and monitoring;
- public incident/status reporting;
- contractual service levels and support operations;
- public-sector accessibility practice;
- explicit controller/joint-controller documentation.

### OKFN appears publicly stronger in:

- current cross-platform desktop product experience;
- local-first/offline desktop architecture;
- local-model AI in a desktop application;
- a closer public precedent for end-user desktop distribution.

### Neither establishes:

- acceptance of the one accountable ECO publisher role;
- complete private vulnerability and desktop incident process;
- Authenticode/key governance;
- exact ECO build/SBOM/signature provenance;
- official release withdrawal;
- anonymous offline end-user support;
- insurance/product liability;
- full UK/EU regulatory ownership;
- binding no-custody and prohibited-use commitments;
- operational tabletop evidence.

The comparison confirms the value of testing multiple candidates. It does not justify combining them, contacting either, or creating a shortlist.

## 16. Provisional decision

**Status:** `HOLD`

**Reasoned conclusion:**

mySociety is a credible, established and operationally mature public-interest technology organisation with strong FOSS, public-service, security, accessibility, privacy and support evidence.

Its current public operating model is not ECO's model. It relies substantially on hosted services, institutional client accounts, monitoring, integrations and personal-data processing across a charitable company and an arm's-length trading subsidiary. The reviewed record does not establish one accountable ECO legal entity, native Windows release/signing, exact artifact provenance, private desktop vulnerability intake, offline support/withdrawal, product complaints, insurance/liability, regulatory ownership or binding no-custody acceptance.

A full-role positive decision, shortlist or outreach proposal would be premature.

## 17. Permitted next action

Permitted under issue #62:

1. preserve this full-role `HOLD` assessment;
2. conduct a source-bounded control comparison with the OKFN record;
3. verify whether a public group-wide vulnerability route exists outside the reviewed GitHub Security pages;
4. inspect current published annual-report or contractual evidence only through valid auditable processes;
5. assess another full-role candidate only where it offers a materially different model;
6. deepen limited-role evidence only where it can change a decision;
7. stop for control review before shortlist or outreach.

## 18. Prohibited next action

This record does not authorise:

- contact with mySociety or SocietyWorks;
- a meeting, demo, quotation, proposal, grant or sponsorship application;
- naming either organisation as preferred or shortlisted;
- combining mySociety, OKFN, Conservancy or AbilityNet into an assumed partnership;
- transfer of repository, source, brand, domain or signing authority;
- disclosure of real evidence, private diagnostics or credentials;
- a real-data, public-sector, benefits, housing, healthcare or justice pilot;
- signing, executable publication, release or deployment;
- any statement that either organisation knows about, supports or approves ECO.

## 19. Public source register

| ID | Source owner | Record | Access date | Proposition | Limitation |
|---|---|---|---|---|---|
| M1 | Companies House | `MYSOCIETY overview` | 2026-08-04 | legal identity, active status, accounts and confirmation statement | registry facts only |
| M2 | Charity Commission | `mySociety governance` | 2026-08-04 | charity identity, on-time reporting and listed policies | does not expose policy contents or ECO authority |
| M3 | mySociety | `Structure and governance` | 2026-08-04 | charitable company, SocietyWorks subsidiary, shared staff, boards and arm's-length relationship | self-published; no ECO allocation |
| M4 | mySociety | `Board` | 2026-08-04 | current trustees and multi-person governance | biographies do not establish acceptance |
| M5 | mySociety | `Team` | 2026-08-04 | CEO, operations, finance, support, developers and technical operations | public role list does not prove availability |
| M6 | mySociety | `Privacy and security` | 2026-08-04 | data custody, service policies, third parties, cookies, analytics, Cyber Essentials and March 2021 update date | materially old and website/service-specific |
| M7 | mySociety GitHub | organisation and public repositories | 2026-08-04 | extensive active FOSS estate | repository count/activity is not product qualification |
| M8 | mySociety GitHub | `Alaveteli` | 2026-08-04 | long-running international FOSS, testing, issues and pull requests | hosted web platform, not Windows desktop |
| M9 | GitHub | `mysociety/alaveteli/security` | 2026-08-04 | no security policy detected and no published advisories shown | public GitHub view; internal processes may exist |
| M10 | GitHub | `mysociety/fixmystreet/security` | 2026-08-04 | no security policy detected and no published advisories shown | public GitHub view; internal processes may exist |
| M11 | Companies House | `SOCIETYWORKS LTD overview` | 2026-08-04 | separate active trading company and current filings | does not allocate ECO duties |
| M12 | Companies House | `SOCIETYWORKS LTD charges` | 2026-08-04 | historical charge shown outstanding with all property/undertaking released; parent as entitled person | instrument not reviewed; no distress inference |
| M13 | SocietyWorks | `Hosted and secure` | 2026-08-04 | Cyber Essentials Plus, access control, backups, UK/Ireland hosting, monitoring and recovery | promotional/contractual hosted-service evidence |
| M14 | SocietyWorks | `Service Level Agreement` | 2026-08-04 | severity, monitoring, response targets and Freshdesk client process | institutional hosted-service model |
| M15 | SocietyWorks | public status page | 2026-08-04 | current service monitoring, maintenance and incident history | not desktop security or release withdrawal |
| M16 | SocietyWorks | `Accessible services` | 2026-08-04 | WCAG claims, screen-reader/keyboard testing and audit advice | web-service claims, not ECO conformance |
| M17 | SocietyWorks | `Data sharing and security agreement` | 2026-08-04 | controller/joint-controller roles, staff access, breach and rights procedures | public-service data-custody model differs from ECO |
| M18 | SocietyWorks | `Privacy policy` | 2026-08-04 | support/contact data, third parties, marketing and indefinite support-message retention | not ECO-specific |
| M19 | SocietyWorks | `Ethical statement` | 2026-08-04 | human agency and non-discrimination boundaries | high-level policy only |
| M20 | SocietyWorks | `Board` | 2026-08-04 | separate board and relevant commercial/public-sector expertise | no ECO authority or acceptance |

---

**No-interest statement:** mySociety and SocietyWorks have not been contacted. This record does not imply knowledge, interest, consent, endorsement, capacity, acceptance or appointment.
