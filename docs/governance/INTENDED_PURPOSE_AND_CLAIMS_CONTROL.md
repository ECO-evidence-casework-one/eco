# ECO intended purpose and public claims control

**Authority:** controlling only when this record is present on canonical `main`; branch and pull-request copies are proposals  
**Issue:** #16  
**Application conformance:** not established by this document  
**Release effect:** none by itself

## 1. Authority and limits of this document

On canonical `main`, this document defines the intended project boundary for official ECO development, documentation, demonstrations, promotion, supply and support.

A copy or change on any other branch or pull request is a proposal and has no controlling effect. A change becomes controlling only when merged into canonical `main` after independent review.

Adoption of this governance record does **not** by itself:

- prove that the application implements the boundary;
- determine legal or regulatory classification;
- approve real-evidence use;
- approve a public binary, signing or release;
- appoint a publisher, supplier, controller, processor, support operator or professional adviser;
- approve institutional, healthcare, justice-sector or EU deployment.

Those questions require separate implementation evidence, role assessment and release approval.

## 2. Intended purpose

Evidence & Casework One is intended to be a private, local evidence and casework assistant for people working with material held on their own computer.

Its intended purpose is to help a user preserve, organise, search, review and understand supplied evidence and casework while keeping the user connected to the original source.

ECO is intended for user-controlled preparation and understanding. It is not intended to replace a court, public authority, healthcare professional, regulated legal professional, forensic laboratory, employer, landlord, insurer or other decision-maker.

## 3. Intended users

Intended users may include:

- an individual working on their own evidence or casework;
- a person helping that individual with their permission;
- an authorised advocate or professional using ECO only within their own competence, authority and organisational rules;
- developers and reviewers using synthetic information for controlled testing.

A user's review of an output does not transfer or reduce any duty that may belong to a future publisher, supplier, deployer, controller, processor, professional adviser or institution.

## 4. Intended operating environment

The intended official product is a self-contained Windows desktop application operating locally on the user's device.

The project design boundary is:

- fully offline application operation;
- no required account, cloud service, telemetry or advertising;
- no routine transfer of case files or generated material to the developer or publisher;
- original evidence preserved separately from derived readings;
- free and open-source project code and approved bundled components;
- meaningful operation on the project's low-spec Windows baseline.

These are product requirements, not release claims. They must be independently verified for the exact final executable and every bundled runtime or model before public release.

## 5. Intended inputs

ECO may work with user-supplied evidence and casework concerning subjects such as:

- benefits and public services;
- housing and homelessness;
- healthcare and medical evidence;
- complaints and safeguarding;
- employment and education;
- consumer, financial and contractual disputes;
- legal correspondence and proceedings;
- other personal or organisational casework.

The presence of health, legal or other sensitive subject matter is not, by itself, a reason to disable evidence assistance.

Use with real, sensitive or irreplaceable material remains blocked until the relevant evidence-integrity, privacy, security, output-boundary and release gates are independently satisfied.

## 6. Permitted evidence assistance

Subject to exact source binding, visible output status, meaningful user review and the implementation gates recorded elsewhere, ECO may be designed to:

- import, preserve and identify original files;
- calculate and verify file hashes;
- display and navigate supplied material;
- perform deterministic extraction and OCR;
- search exact text and retrieve source passages;
- identify names, dates, deadlines, events and requested actions stated in documents;
- organise user-selected material;
- produce source-linked summaries, chronologies and comparisons of what supplied documents say;
- help the user prepare questions, notes, letters and draft responses;
- identify possible omissions, conflicts or inconsistencies for the user to examine;
- explain uncertainty, missing support and the need for checking.

ECO may explain what a supplied document states. It must not convert evidence assistance into an unsupported professional, clinical, forensic or official conclusion.

## 7. Required output status

The application and every export must distinguish clearly between:

1. **Original source text** — content preserved from the supplied material.
2. **Extraction or OCR** — machine-derived reading that may contain errors.
3. **ECO-generated suggestion** — an application or model output requiring review.
4. **User confirmation** — a statement the user has deliberately confirmed.
5. **User-authored note** — content written by the user.

Generated text must not be presented as verified fact merely because ECO produced it.

Where a statement is derived from supplied evidence, ECO should retain a usable route back to the supporting source location. Where support is absent, uncertain or conflicting, the output must say so.

## 8. Medical and clinical boundary

ECO is not intended to be a medical device, clinical decision-support system or healthcare professional. This project statement does not decide the legal classification of a final product; classification must be assessed from the actual features, outputs, claims, users and deployment.

ECO must not be presented, configured or relied upon to:

- diagnose a condition or infer an unstated diagnosis;
- recommend medication, treatment or a clinical course of action;
- perform prognosis, triage, monitoring or clinical-risk assessment;
- make an emergency or safeguarding assessment;
- direct clinical care or materially influence a clinical decision;
- replace a healthcare professional, emergency service or safeguarding route.

ECO may identify and summarise what a supplied medical letter or record states, locate dates and passages, and help a user prepare questions or correspondence, provided it does not add a diagnosis, treatment recommendation or clinical-risk conclusion.

## 9. Legal-services boundary

ECO is not intended to be a solicitor, barrister, authorised legal representative or authoritative current-law service. This project statement does not decide whether a future feature or activity is regulated; the actual conduct, authority, feature and deployment must be assessed.

ECO must not be presented, configured or relied upon to:

- conduct litigation;
- exercise a right of audience;
- undertake another reserved legal activity without lawful authorisation or exemption;
- claim to represent the user before a court, tribunal or authority;
- guarantee legal correctness, admissibility, prospects, remedy or outcome;
- replace independent legal advice where professional advice is required.

ECO may help a user organise legal material, locate relevant passages, explain what supplied sources say and prepare user-controlled drafts. Current-law statements require current authoritative sources and must not be presented as professional legal advice.

## 10. Evidence and forensic boundary

ECO is not a forensic laboratory, accredited method, expert witness or guarantee of authenticity or admissibility.

Hashing, preservation records, provenance links and integrity checks are technical functions. They do not by themselves establish legal authenticity, chain of custody, evidential weight or admissibility.

The project must not use claims such as `forensic`, `forensically sound`, `court-ready`, `admissible` or `expert-grade` unless the exact claim, scope and evidence have received separate approval.

## 11. Decision-making, profiling and high-consequence use

ECO is intended to assist a user preparing and understanding their own material. It is not intended to decide another person's rights, access, credibility or risk.

ECO must not be presented, configured or relied upon to:

- determine or score eligibility, entitlement, credibility, honesty or dangerousness;
- grant, deny, reduce, revoke or reclaim a benefit, service, right or opportunity;
- rank people for housing, employment, education, healthcare, insurance, migration, policing or another essential service;
- profile a natural person for a high-consequence decision;
- autonomously make or materially influence an official adverse decision.

An individual using ECO to prepare their own case must not be treated as equivalent to an authority-side or institution-side decision workflow.

Any proposed official supply, approval, endorsement, promotion or support for use by a public authority, court, tribunal, healthcare body, employer, landlord, insurer, education body or benefits decision-maker requires a fresh feature-level, role-level and deployment-level legal and safety assessment.

## 12. Emergency and safeguarding boundary

ECO is not an emergency, crisis or safeguarding service.

The application must not imply that it monitors the user, detects emergencies reliably or has contacted support. Where a user request indicates an immediate emergency or safeguarding need, ECO should provide a calm boundary and direct the user to appropriate real-world support without pretending the application has taken action.

## 13. Human review

ECO outputs are intended to support human understanding and preparation.

Meaningful review requires the user to be able to:

- see the source and output type;
- inspect relevant supporting passages;
- identify uncertainty or missing material;
- correct or reject a suggestion;
- decide whether and how to use a draft.

A generic warning or mandatory tick box is not sufficient where the interface obscures the source, output status or uncertainty.

Human review does not cure a prohibited purpose and must not be used to justify diagnosis, professional representation, profiling, scoring or official adverse decision-making.

## 14. Privacy and data-role boundary

The intended official application processes case material locally and does not require the project to receive the user's files.

That design materially limits project access but does not predetermine every data-protection role. Controller, processor and joint-controller status must be assessed for each actual processing activity and deployment.

The assessment must be repeated if ECO or a future organisation introduces:

- cloud or remote processing;
- telemetry or analytics;
- support intake containing case material;
- account-linked storage or synchronisation;
- institutional administration or monitoring;
- access to user files or generated outputs;
- another purpose for processing personal information.

Users must not be required to submit real evidence or sensitive case material to obtain ordinary support.

## 15. Public claims control

Public wording must match the exact implemented and independently verified state.

Unless separately supported and approved, ECO must not be described as:

- certified;
- compliant;
- secure;
- accurate;
- accessible;
- production-ready;
- release-ready;
- forensic or forensically sound;
- court-ready or admissible;
- a legal adviser;
- a medical device or clinical system;
- an autonomous or official decision-maker.

A qualified statement such as `designed to`, `intended to` or `source-level rule` must not be used to obscure a known implementation gap.

Synthetic test results must state the test material, sample size, method, limitations and why the result must not be generalised to real-world accuracy.

## 16. Official project identity, open-source rights and forks

This control governs official ECO project decisions, official releases, project-endorsed supply, promotion and support.

It does not remove or narrow lawful rights under ECO's open-source licence to copy, modify or convey the code.

Open-source redistribution rights do not automatically grant authority to:

- describe an independent build as an official ECO release;
- imply endorsement, support, signing or approval by the ECO project;
- use project names, marks or release channels in a misleading way;
- represent that an independent deployment passed ECO's governance or release gates.

An independent fork or redistribution should identify itself truthfully, preserve applicable licence obligations and avoid false endorsement. The official project should identify approved releases through its authoritative release channel, exact hashes and signature information.

This document controls what the ECO project will approve, endorse, supply, promote or support. It is not a universal command to independent third parties, and it does not determine duties that applicable law may attach to their actual conduct.

## 17. Cross-surface consistency

The same boundary must be reflected in:

- README and repository status documents;
- application onboarding, help and safety messages;
- AI system instructions and model cards;
- exports and generated reports;
- release notes and known limitations;
- screenshots, demonstrations and training material;
- website, partner, procurement and institutional material;
- support and incident-response wording.

Where wording conflicts, the more restrictive current release gate applies until the conflict is corrected and independently reviewed.

## 18. Change triggers

A fresh legal, regulatory, safety and claims review is required before the project approves or promotes a change that would:

- add diagnosis, treatment, monitoring, prognosis, triage or clinical-risk functions;
- provide professional representation or reserved legal activity;
- score, profile, rank or decide about people;
- materially influence an institutional or official decision;
- add cloud processing, telemetry, accounts or project access to case material;
- supply or support an institutional, healthcare, justice-sector or authority-side deployment;
- make ECO available in a jurisdiction or distribution model not covered by the current assessment;
- make a new compliance, accuracy, security, accessibility, forensic or professional claim.

The review must consider the actual feature, intended users, outputs, marketing, deployment, legal roles and data flows. A feature name or disclaimer cannot substitute for the assessment.

## 19. Implementation and release dependencies

This document does not close the linked implementation and organisational gates.

At minimum:

- issue #20 must prove the permitted evidence-assistance and prohibited-output boundary across deterministic and local-model routes;
- evidence preservation, workspace integrity, restore and Ask concurrency must pass their active gates;
- diagnostic privacy and offline-operation claims must be independently verified;
- accessibility claims require objective evidence;
- issue #15 must reconcile the exact final executable, SBOM, licences, source and signing records;
- issue #17 requires an accountable organisation to accept publisher and continuity duties before official public or institutional release;
- the exact final product requires fresh jurisdiction, feature and deployment assessment before institutional, healthcare, justice-sector or EU availability.

## 20. Acceptance criteria for this governance control

Before issue #16 can close:

- this record is merged into canonical `main` after exact-head independent review;
- README, current status and release-gate wording do not contradict it;
- application and model-facing wording has a tracked conformance plan;
- automated documentation checks cover prohibited and unapproved claims;
- synthetic tests demonstrate the medical, legal, forensic, profiling, scoring and official-decision boundaries;
- source text, OCR, generated suggestions, confirmations and notes remain distinguishable;
- change triggers are represented in the release and governance process;
- no document treats adoption as implementation proof, regulatory classification or release approval.

Issue #16 should remain open until the merged record and cross-surface evidence have been independently verified.
