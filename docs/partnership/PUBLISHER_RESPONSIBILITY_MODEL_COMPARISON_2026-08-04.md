# ECO publisher responsibility-model comparison

**Issues:** #17 and #62  
**Decision date:** 4 August 2026  
**Control relationship:** project control-room analysis; not external legal advice or organisational acceptance  
**Current position:** no publisher, steward, specialist supplier, partner or outreach target is appointed

## 1. Question examined

This record compares two possible organisational models for a future official ECO distribution:

1. **single-organisation model** — one established legal organisation directly operates every publisher/steward function;
2. **accountable-steward-plus-specialists model** — one established legal organisation remains the sole accountable official publisher/steward while contracting with bounded specialist organisations for defined services.

It does **not** approve a third model in which publisher responsibility is informally divided among several organisations without one clear accountable legal entity.

The comparison follows:

- [`PUBLISHER_AND_STEWARDSHIP_GATE.md`](../governance/PUBLISHER_AND_STEWARDSHIP_GATE.md);
- [`PUBLISHER_ACCEPTANCE_CHECKLIST.md`](../governance/PUBLISHER_ACCEPTANCE_CHECKLIST.md);
- the issue #62 research method and candidate register;
- the first limited-role candidate batch;
- the Open Knowledge Foundation full/near-full `HOLD` assessment.

## 2. Controlling decision

### Preferred legal and operational structure

The preferred future model is:

> **one clearly identified accountable official ECO publisher/steward, permitted to use contracted specialist organisations under written, auditable and replaceable arrangements.**

The accountable steward must remain responsible for the whole official ECO distribution even where a specialist performs work.

Specialist involvement must not create:

- uncertainty over who users contact;
- gaps between contracts;
- circular referrals between organisations;
- conflicting release or safety decisions;
- duties falling back onto an individual contributor;
- a claim that no organisation is responsible because each performed only one part;
- hidden joint control of user data or project assets;
- dependence on an informal personal relationship.

### Rejected structure

The following structure is not approved:

> several organisations independently holding fragments of ECO responsibility, with no organisation accepting overall publisher authority, product accountability, incident command and continuity.

That model would not satisfy issue #17.

## 3. Why one accountable steward remains necessary

ECO requires decisions that cannot safely be fragmented or left to consensus after an incident, including:

- whether an exact candidate may become an official release;
- whether a release must be refused, withdrawn or declared unsupported;
- whether a vulnerability, harmful output, accessibility failure or evidence-integrity defect is severe enough to stop distribution;
- who controls the official name, repository and release route;
- who holds signing authority and can revoke or recover it;
- who answers users, regulators, institutions and insurers;
- who owns legal classification, product claims and deployment limits;
- who carries contractual duties and approved residual risk;
- who preserves continuity if a maintainer or specialist is unavailable.

A user must not need to determine whether an incident belongs to the fiscal host, developer, accessibility assessor, model supplier, security consultant or repository owner before receiving a response.

The accountable steward may delegate performance. It must not delegate away accountability without a new formal role decision.

## 4. Model A — one organisation directly operates all functions

### Description

One established organisation employs or directly controls the people, systems and resources needed for:

- governance and finance;
- FOSS/legal and licence review;
- software maintenance;
- security and vulnerability response;
- accessibility assurance;
- privacy and complaints;
- release engineering and signing;
- user support;
- institutional contracting;
- regulatory assessment;
- repository and continuity recovery.

### Advantages

- one legal identity and governing body;
- one public contact structure;
- simpler authority and escalation;
- fewer contracts and role boundaries;
- easier incident command;
- clearer insurance and liability allocation;
- less risk of specialists blaming one another;
- easier continuity planning if functions are genuinely institutionalised.

### Risks and limitations

- very few organisations publicly prove every required capability;
- mission fit may be broad while practical product capability is weak;
- an organisation may overstate internal expertise and avoid independent challenge;
- maintaining every specialist function can be costly;
- the organisation may be strong in FOSS or civic technology but weak in native Windows accessibility, signing, high-consequence output assurance or user support;
- concentration increases the effect of organisational failure, strategy change or funding loss;
- a single organisation can still depend on one internal individual unless deputies and recovery are tested.

### Acceptance condition

Model A is acceptable only if the candidate passes every relevant item in the publisher acceptance checklist or identifies a funded and controlled route to obtain missing expertise without weakening its own accountability.

## 5. Model B — one accountable steward plus contracted specialists

### Description

One established legal organisation is the official ECO publisher/steward. It retains final authority and contracts with specialist organisations or qualified professionals for bounded services.

Possible specialist functions include:

- FOSS fiscal, copyright, trademark and licence support;
- accessibility testing and disabled-user research;
- penetration testing and security review;
- independent release/provenance inspection;
- UK and EU legal/regulatory advice;
- insurance brokerage and product-liability advice;
- code-signing service or managed certificate custody;
- privacy/data-protection advice;
- specialised user research or cognitive-accessibility work.

### Advantages

- permits the steward to obtain genuine specialist capability rather than pretending one organisation can do everything;
- creates opportunities for independent challenge;
- can replace one supplier without transferring the whole project;
- can scope synthetic-only work and prevent access to real user evidence;
- may fit the evidence found so far: different organisations show strengths in FOSS hosting, civic technology and accessibility;
- can reduce the risk that specialist duties fall informally on one contributor;
- allows competitive procurement or periodic reassessment where appropriate.

### Risks and limitations

- contracts can leave gaps or overlapping duties;
- confidential information may be disclosed to more organisations;
- specialists may use cloud tools, accounts or data-retention practices incompatible with ECO;
- a supplier may withdraw, fail or change ownership;
- responsibility may become blurred unless the steward remains publicly accountable;
- release delays and coordination costs increase;
- separate insurance and indemnity limits may not align;
- suppliers may produce advisory reports without accepting operational remediation duties;
- subcontracting chains may be hidden;
- a specialist's branding or certification language may cause unsupported ECO claims;
- multiple organisations may each rely on the others' evidence without one party verifying the complete release.

### Acceptance condition

Model B is acceptable only where the accountable steward itself passes the publisher gate and each specialist is governed by a written agreement that preserves the steward's final accountability.

## 6. Model C — fragmented or collective responsibility without one accountable publisher

### Description

Different organisations informally perform different functions, but no legal organisation accepts overall publisher authority and operational responsibility.

Examples include:

- one group hosts funds;
- another holds the repository;
- another signs releases;
- another answers accessibility complaints;
- volunteers maintain code;
- nobody owns release withdrawal, incident command, liability or continuity.

### Decision

**Rejected.**

### Reasons

This model would create unacceptable uncertainty concerning:

- official identity;
- product claims;
- contracting authority;
- data roles;
- vulnerability response;
- release withdrawal;
- signing revocation;
- complaints and support;
- insurance and liability;
- regulatory reporting;
- continuity and project transfer.

A memorandum describing cooperation would not cure the absence of one accountable legal organisation unless it creates a legally and operationally effective lead-publisher structure.

## 7. Non-delegable steward responsibilities

The following responsibilities must remain with the accountable official steward even where specialists advise or perform tasks:

| Responsibility | Steward must retain |
|---|---|
| Official release decision | final approve/refuse/withdraw authority |
| Intended purpose and prohibited uses | final product-purpose and claims authority |
| Official name and endorsement | control of official identity and release designation |
| Issue-gate closure | decision that required evidence is sufficient |
| Security incident command | severity, containment, advisory and withdrawal command |
| Supported versions | final maintenance and end-of-support decisions |
| Signing approval | approval of the exact final artifact and revocation decisions |
| Repository continuity | recovery, freeze, transfer and emergency authority |
| Complaints accountability | one route and final responsibility for outcome |
| Data-role decisions | final controller/processor assessment for ECO activities |
| Institutional contracts | contracting party, warranties, liability and exit decisions |
| Regulatory classification | ownership of feature/role/market/deployment assessments |
| Insurance | decision that cover or recorded risk position is adequate |
| Specialist governance | selection, contract, monitoring, replacement and escalation |
| Public communication | authoritative warnings, corrections and status statements |

A specialist may recommend a decision or operate a process. It must not become the unrecorded de facto publisher through actual conduct.

## 8. Functions that may be contracted

The following functions may be obtained from specialists where scope, evidence, privacy and accountability are controlled:

| Specialist function | Minimum contract controls |
|---|---|
| Accessibility audit/user research | synthetic-only materials unless separately approved; accessible deliverables; AT/device matrix; data minimisation; findings and limitations; remediation route |
| FOSS/licence advice | exact component/source scope; conflicts; corresponding-source obligations; privilege/confidentiality where applicable; no proprietary-relicensing condition |
| Security testing | private route; test authorisation; synthetic data; vulnerability confidentiality; severity method; retest; disclosure coordination; deletion/retention |
| Privacy advice | exact data flows; roles; support/complaint data; retention; processors/subprocessors; breach route |
| Regulatory/legal advice | jurisdiction, feature, role, claims and deployment scope; assumptions; review date; no unsupported certification language |
| Build/provenance inspection | exact source/artifact identity; SBOM reconciliation; toolchain; reproducibility; signature; adverse findings preserved |
| Signing service | organisation-controlled approval; exact hash; key custody; multi-person recovery; revocation; logs; no post-signing mutation |
| Insurance advice | exact activities and territories; exclusions; limits; claims route; continuity after insurer/broker change |
| UX/cognitive-accessibility research | synthetic scenarios; disabled-user protections; consent/data roles; no transfer of case evidence; findings tied to tested build |

## 9. Required contract terms for every specialist

Every specialist agreement must address, as applicable:

1. legal identity and authorised signatory;
2. exact services, deliverables and excluded services;
3. ECO commit, build or document scope;
4. use of synthetic information and prohibition on unnecessary real evidence;
5. data roles, confidentiality, access, retention, deletion and breach reporting;
6. approved tools, locations and subprocessors;
7. free/open-source and publication requirements;
8. intellectual-property and report-use rights;
9. independence and conflicts of interest;
10. competence, named role and deputy arrangements;
11. service levels and escalation;
12. obligation to preserve adverse findings;
13. correction and retest;
14. insurance, liability, warranties and limits;
15. security and incident notification;
16. termination, handover and continuity;
17. prohibition on public claims, endorsement or contact with users without authority;
18. prohibition on subcontracting without approval where material;
19. audit/evidence access for the steward;
20. governing law and dispute route.

A purchase order or informal email is not sufficient for a high-risk specialist function.

## 10. Data-protection and evidence-custody model

### Default rule

Specialists must not receive real ECO user evidence, case files, conversations, diagnostics or identifiers merely to perform their role.

Use:

- synthetic cases;
- generated documents;
- redacted public-safe screenshots;
- seeded test canaries;
- controlled technical logs without content;
- isolated test builds.

### Where personal data cannot be avoided

Before sharing personal data, the steward must document:

- why synthetic information is insufficient;
- the lawful purpose and role allocation;
- data minimisation and alternatives;
- explicit categories and individuals affected;
- access, location, encryption, retention and deletion;
- processor/joint-controller terms where applicable;
- international transfers;
- incident and complaint routes;
- approval authority;
- evidence that the relevant ECO gate permits the activity.

The current project position does not approve such sharing.

## 11. Security incident command

Under Model B:

- every specialist must report material concerns immediately through the steward's private route;
- the steward owns severity classification and cross-supplier coordination;
- the steward decides whether to stop, warn, withdraw, revoke or publish an advisory;
- specialists must preserve evidence and avoid unilateral public disclosure unless law requires it;
- the steward maintains one incident chronology and decision log;
- disagreement must be recorded, not hidden;
- a specialist's failure to respond must not prevent release withdrawal or public warning.

An accessibility, legal or FOSS specialist discovering a security or harmful-output issue must have an escalation route even where that issue is outside its contracted speciality.

## 12. Complaints and support model

Users must receive one official front door for ECO matters.

The steward may internally route:

- security reports;
- privacy complaints;
- accessibility barriers;
- harmful-output or evidence-provenance complaints;
- ordinary support;
- legal notices.

The public must not be told to diagnose which contractor caused the problem.

The steward remains responsible for:

- acknowledgement;
- appropriate investigation;
- coordination;
- outcome communication;
- correction or withdrawal decisions;
- records and learning;
- accessible alternatives and reasonable adjustments where applicable.

A specialist may respond directly only under controlled authority and must not request public submission of real evidence.

## 13. Release and signing model

A specialist may build, inspect or sign an artifact only if:

- the steward approved the exact candidate and scope;
- source, manifest, SBOM, notices, hashes and build receipt are reconciled;
- the signing request identifies the final immutable file hash;
- the steward retains approval/revocation authority;
- key custody and recovery are organisationally controlled;
- logs are preserved;
- no specialist can independently declare a build official;
- post-signing mutation is prohibited and tested;
- the steward can withdraw the release if the specialist becomes unavailable.

A fiscal host, GitHub account owner, certificate provider or build service does not become the official ECO publisher merely through technical possession of an asset.

## 14. Repository and asset ownership

Before Model B operates, the steward must decide and record ownership or control of:

- official repository and organisation;
- domains and website;
- ECO name and logo;
- copyright or necessary licences;
- signing certificates and accounts;
- release archives and source mirrors;
- security advisory records;
- support and complaint records;
- build systems and reproducibility inputs;
- project funds and contracts.

Specialists should receive only the permissions required for their function. Access must be removable without loss of project continuity.

No asset may depend solely on an individual contributor or an unrecorded specialist account.

## 15. Insurance and liability allocation

The steward must not assume that a specialist's professional indemnity or public-liability insurance covers ECO's complete product risk.

The responsibility record must identify:

- the steward's own insurance or explicit decision not to obtain cover;
- each specialist's relevant cover;
- exclusions for software, cyber, AI, professional advice, US/Canada, medical/legal use or public-sector supply;
- contractual liability limits and indemnities;
- whether one policy responds before or after another;
- who manages and notifies claims;
- what happens after termination or insolvency;
- whether run-off cover exists where needed.

An individual contributor must not become the uninsured residual liability bearer.

## 16. Regulatory ownership

Specialists may advise on:

- data protection;
- accessibility;
- medical-device boundaries;
- legal-services boundaries;
- EU AI Act roles;
- Cyber Resilience Act roles;
- consumer/product safety and liability;
- public procurement.

The steward must own the final documented decision for the actual feature, claim, market and deployment.

No specialist report may be converted into a general statement that ECO is legally compliant, certified, safe, accessible or accurate.

A deployment partner or institutional buyer may acquire separate duties. That does not remove the steward's duties for its official distribution and claims.

## 17. Specialist withdrawal and continuity

Every high-risk specialist relationship must have an exit plan covering:

- advance notice where practicable;
- immediate termination for security, conflict or misconduct;
- return/deletion of information and credentials;
- handover of work, evidence and open findings;
- replacement-provider transition;
- preservation of licences and report-use rights;
- revocation of repository, signing and system access;
- ongoing confidentiality;
- incident cooperation after termination;
- public communication where service continuity is affected.

The steward must be able to stop official releases while a critical specialist function is unavailable.

## 18. Decision matrix

| Criterion | Model A: one organisation operates all | Model B: accountable steward + specialists | Model C: fragmented responsibility |
|---|---|---|---|
| Clear official accountability | strong if genuine | strong if steward remains accountable | unacceptable |
| Access to specialist expertise | may be limited internally | strong | variable and uncontrolled |
| Contract complexity | lower | higher | high/unclear |
| Incident command | simpler | requires disciplined lead | unclear |
| User contact route | one | one through steward | fragmented |
| Independence/challenge | may be weak | can be strong | unclear |
| Supplier failure resilience | internal concentration risk | replaceable with exit plans | gaps likely |
| Data-sharing risk | lower number of entities | higher; must be minimised | uncontrolled |
| Signing/release authority | one organisation | steward must retain | unclear |
| Regulatory ownership | one organisation | steward must retain | unclear |
| Fit with current public research | no candidate yet proves all functions | better matches observed specialist strengths | not acceptable |
| Issue #17 compatibility | possible | preferred if controlled | fails |

## 19. Recommended future assessment architecture

### Organisational layer

Assess one candidate as the sole accountable publisher/steward against every acceptance-checklist item.

### Specialist layer

Where the steward lacks internal evidence, identify whether it can:

- employ qualified staff;
- fund external specialists;
- contract and supervise them;
- understand and act on their findings;
- replace them without losing continuity.

### Release layer

After organisational acceptance, the exact release candidate still requires independent technical, legal, accessibility, provenance and security evidence.

These layers must not be collapsed. A good specialist report does not appoint a publisher, and a capable publisher does not prove a release.

## 20. Candidate-research implications

The current research supports the following interpretations only:

- Software Freedom Conservancy may warrant deeper review as a limited FOSS fiscal/legal host;
- AbilityNet may warrant deeper review as a limited accessibility specialist;
- mySociety remains on hold for a bounded civic-technology role;
- Open Knowledge Foundation remains on hold for the full/near-full steward role.

No current result shows that these organisations should work together, that any would accept a role, or that a multi-party structure is commercially or legally feasible.

## 21. Required evidence before Model B may pass

The accountable steward must provide:

- a complete internal responsibility map;
- governing-body acceptance of retained duties;
- specialist procurement and conflict policy;
- contract templates or executed agreements;
- data-flow and access matrix;
- incident command plan;
- complaints routing and public front door;
- signing/release authority matrix;
- repository and asset permission register;
- insurance/liability analysis;
- specialist withdrawal and replacement plans;
- tabletop exercises covering supplier failure, disagreement and urgent withdrawal;
- proof that no duty defaults to an individual contributor.

## 22. Tabletop tests required

Before an official release, test at least:

1. the accessibility specialist reports a critical blocker immediately before signing;
2. the security specialist and developer disagree on severity;
3. the signing provider becomes unavailable;
4. the fiscal/legal host ends the relationship;
5. a specialist accidentally receives a file containing personal data;
6. an institutional buyer requests a prohibited authority-side feature;
7. the primary maintainer is unavailable during an incident;
8. a supplier publishes an unsupported statement implying ECO certification;
9. a specialist's report is incomplete or demonstrably wrong;
10. two specialists each say the other owns remediation.

The steward must demonstrate that it can stop, decide, communicate, preserve evidence and continue without transferring duties to one contributor.

## 23. Current decision and stop point

**Decision:** Model B—one accountable steward plus controlled specialists—is the preferred future operating model because it preserves one accountable legal entity while allowing genuine specialist expertise.

**Model A remains acceptable** where one organisation can prove all capabilities internally or through personnel it directly controls.

**Model C is rejected.**

This model decision does not:

- identify the future steward;
- create a candidate shortlist;
- approve any specialist;
- authorise outreach or procurement;
- approve contracts or funding;
- appoint an organisation or individual;
- permit real evidence or support intake;
- approve signing, an executable, release or deployment.

The next permitted research step is to test whether a full-role candidate could govern Model B and whether the limited-role candidates' public evidence would support specialist due diligence. Stop before contact or shortlist creation.
