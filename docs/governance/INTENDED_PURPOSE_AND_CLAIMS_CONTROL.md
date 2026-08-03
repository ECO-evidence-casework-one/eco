# ECO intended purpose and public-claims control

**Control effect:** project-governance boundary when present on the canonical branch; adoption does not prove implementation conformance, close issue #16 or issue #20, determine regulatory classification or approve a release or deployment  
**Control date:** 3 August 2026  
**Current release position:** source development only; no approved public binary, real-evidence use or deployment

## Purpose of this control

This document defines the intended product boundary for Evidence & Casework One (ECO) and the claims that may or may not be made about it.

It is a project-governance control. It is not a legal opinion, medical-device classification, forensic accreditation, accessibility conformance report or deployment approval.

Where this document conflicts with a stricter release gate, risk decision or legal requirement, the stricter control applies.

## Control layers and current source baseline

This control operates at four separate levels:

1. **Project-governance boundary:** defines what the ECO project may approve, endorse, promote or supply.
2. **Implementation conformance:** requires objective evidence that the exact application build enforces the boundary.
3. **Legal or regulatory classification:** depends on actual functionality, claims, users, distribution and deployment context and is not determined by this document.
4. **Release or deployment approval:** remains controlled by the canonical release gate and every open P0/P1 finding.

The implementation observation in the health-input section is tied to application source at commit `001e4fc2d55f8c29a9eaa20a0f741162433aadad`. Later documentation and control commits, including canonical `main` reviewed at `aac8f6f3d5d9314d765cc14546d66ec7fc376aec`, must not be treated as proof that the Ask implementation changed unless the relevant application-source delta and exact tests are independently reviewed.

## Intended product purpose

ECO is intended to be a user-controlled, local Windows desktop tool that helps a person:

- create and maintain application-controlled local copies of user-selected files;
- calculate and record cryptographic hashes for byte-comparison and integrity checking;
- organise files, notes and conversations into local matters;
- produce clearly identified derived readings such as extracted text, OCR suggestions, previews and search indexes;
- search verified local sources and navigate back to supporting source regions where available;
- receive source-linked organisational suggestions and draft text for the user to review, correct, confirm or reject;
- export user-selected records with clear provenance and limitations.

ECO is intended to support **information organisation and user-led casework preparation**. It does not replace the user, a qualified professional, a public authority, a court, a healthcare professional or an emergency service.

## Intended users

The intended primary user is an individual working with files and casework that they are lawfully entitled to hold and review.

The ECO project will not approve, endorse, supply, promote or support an official deployment for advisers, charities, public bodies, healthcare organisations, legal practices or other institutions until a separate deployment assessment, accountable publisher, contractual and data-role allocation, accessibility evidence, security evidence and any required sector-specific review exist. No project-endorsed institutional use is currently approved.

## Intended inputs

The intended inputs are local files and user-entered text deliberately selected by the user, including documents, images and supported archive or office formats.

During the current development stage, only synthetic and non-sensitive test material is permitted.

ECO must not silently collect data from accounts, cloud services, remote systems, microphones, cameras, email or other sources.

## Intended outputs and status labels

ECO outputs may include:

- preserved local copies and integrity receipts;
- extracted text and OCR-derived text;
- search results and source navigation;
- user notes and confirmations;
- generated organisational suggestions and drafts within the approved non-professional boundary;
- diagnostic, backup and export records.

Every material assertion or exported item must identify its status where relevant, including:

- original source material;
- application-preserved copy;
- deterministic extraction;
- OCR-derived reading;
- generated suggestion;
- user-confirmed fact;
- user note;
- corrected, disputed or superseded material.

Generated or derived output is not original evidence merely because it is displayed beside evidence.

## Required human control

ECO must give the user a clear and meaningful opportunity to inspect the supporting source, understand the status and limitations of generated or derived material, correct or reject it, and decide whether to use or send it.

ECO must not claim that it sent, uploaded, deleted, contacted, submitted, decided, diagnosed or changed anything unless a deterministic application action actually occurred and produced a verifiable receipt.

No generated output may be presented as a final professional conclusion, adverse decision or completed external action.

This human-review requirement does not transfer, exclude or reduce any duty that may belong to a future publisher, supplier, deployer, controller, professional adviser or institution. Product design, warnings and support arrangements must not rely on a blanket statement that the user is responsible for everything ECO produces.

## Explicitly excluded uses

ECO is not intended or approved to be used as any of the following.

### Legal and judicial

- legal advice or a substitute for a lawyer;
- an authoritative source of current law, procedure, limitation periods or court rules;
- an autonomous legal decision-maker;
- a system for deciding credibility, liability, entitlement, guilt, innocence or admissibility;
- a tool used by or on behalf of a judicial authority to apply law to facts without a fresh regulatory and legal assessment.

ECO may organise user-supplied legal or casework material and produce reviewable drafts, but it must identify that it cannot verify current law unless the relevant source has been supplied and independently checked.

### Medical, health and safeguarding

ECO is not intended or approved for:

- diagnosis, screening, prognosis or clinical risk prediction;
- treatment, medication or dosage recommendation;
- clinical triage, emergency-call classification or prioritisation;
- physiological or mental-health monitoring;
- a substitute for a healthcare professional, crisis service or emergency service.

#### Health-related input classification

For this control, **health-related source material** means:

- a file the user designates as health, medical, clinical or mental-health material;
- a document whose primary purpose is to record, communicate or support information about physical or mental health, care, diagnosis, treatment, medication, tests, appointments, fitness, capacity or clinical risk;
- a health-related section or embedded clinical record within a mixed-purpose document, such as a benefits appeal, housing submission or legal bundle containing medical evidence.

For mixed-purpose material, the restricted boundary applies to any generated processing of the health-related section or clinical facts. Where ECO cannot reliably isolate the health-related content from the rest of the file, the entire file must be treated as health-related for generated processing.

This classification does not prevent byte-preserving display, deterministic extraction, exact-text search or user-requested sorting on explicit non-clinical metadata.

#### Handling approved under this control for health-related source material

Under this control, approved handling of health-related source material is limited to the same non-clinical document functions available to other source material:

- preserve the selected file;
- calculate and verify byte-comparison hashes;
- display the original or application-preserved copy;
- perform deterministic text extraction or clearly labelled OCR;
- search exact extracted or OCR text and display all matching occurrences;
- navigate to the supporting page or source region;
- organise files using user-chosen matters or tags;
- sort records using explicit metadata selected by the user, such as filename, import date or recorded file date, without interpreting the content;
- store a chronology, summary or note written by the user.

Functions approved under this control must not semantically select, omit, reorder, combine or paraphrase clinical facts on the user's behalf.

#### Handling not approved under this control for health-related source material

No current ECO feature is approved under this control to use generative AI, a language model or semantic application-authored rules to create a summary, synopsis, chronology, priority list or action list from health-related source material.

The prohibition applies even where the output is described as administrative, general-purpose or document organisation. It includes:

- selecting which symptoms, diagnoses, findings, medications or events appear important;
- omitting clinical information from a shorter account;
- paraphrasing or rewriting a clinical statement;
- combining information from different records or dates;
- reordering clinical facts into a narrative or priority sequence;
- identifying change, deterioration, improvement, patterns or trends;
- assigning urgency, risk, significance or relevance;
- generating recommendations, questions for a clinician or health-related action points.

The prohibition concerns **semantic transformation of health content**. It does not include deterministic exact-text matching, byte-preserving display, source navigation or a user-requested sort using explicit metadata that does not inspect or infer clinical meaning.

Any proposal to perform semantic selection, omission, reordering, combination, paraphrase or generation from health-related source material requires a fresh MHRA and wider legal assessment before implementation, testing with real health data or promotion. The product's actual functionality, inputs, outputs and presentation control this boundary; changing the label or adding a disclaimer does not make the function permitted.

#### Representative acceptance examples

**Within the current boundary:**

- show every occurrence of an exact medication name in the source text;
- sort user-selected records by their recorded file date without interpreting the contents;
- let the user type and save their own note or chronology;
- display an OCR reading beside the original image with confidence and provenance;
- search a mixed-purpose benefits appeal for an exact phrase and return every matching source location without summarising the medical evidence.

**Outside the current boundary and requiring reassessment:**

- create a concise summary of a hospital letter;
- choose the most important symptoms or diagnoses;
- produce a medical timeline by extracting and combining clinical events;
- state that a condition appears to be worsening or stable;
- suggest what the user should ask a clinician or do next;
- summarise or prioritise the clinical evidence embedded in a benefits, housing or legal document.

#### Current implementation conformance gap

At application-source baseline `001e4fc2d55f8c29a9eaa20a0f741162433aadad`, `Vault.Ask` classifies a question, ranks source segments, limits and reorders selected passages, and composes summary, date, action, comparison or explanation-oriented output without first applying the health-related input classification above.

This is a source-specific observation, not a permanent statement about every later commit. A later build may be described as conforming only after the relevant application-source delta, whole-vault health gate, isolation boundary, prompts, memory, citations, output and synthetic non-disclosure tests are independently reviewed.

Current canonical `main` was reviewed at `aac8f6f3d5d9314d765cc14546d66ec7fc376aec`. Its later documentation changes do not by themselves close issue #20 or establish implementation conformance.

The observed behavior is not the permitted exact-text search that returns every matching occurrence. For health-related or mixed-purpose source material, the reviewed Ask path can perform semantic selection and reordering outside this intended-purpose boundary.

Issue #20 records the required implementation gate. Until issue #20 is independently closed:

- the current source must not be described as enforcing this health-related boundary;
- if a vault contains any health-related source material, Ask ECO, local-model and other semantic or generated-answer routes must remain unavailable for that entire vault unless an independently verified isolation mechanism proves that restricted material cannot enter ranking, prompts, context, memory, citations or output;
- scoping a question to an unrelated file or topic is not sufficient isolation;
- synthetic boundary tests may exercise these routes only to prove rejection and non-disclosure;
- deterministic exact-text search and the other expressly permitted non-semantic functions remain available;
- adoption of this governance document does not approve the current Ask implementation for health-related material;
- real, sensitive or irreplaceable evidence remains prohibited in every event.

This conformance gap does not itself determine medical-device classification. It means only that the reviewed behavior is not yet aligned with this intended-purpose control.

### Forensic and evidential authority

ECO is not intended or approved as:

- a forensic laboratory or forensic science unit;
- an accredited method or validated forensic process;
- proof of authenticity, authorship, truth, creation time, lawful acquisition or admissibility;
- a guarantee that evidence will be accepted by a court, tribunal, regulator or public authority.

A SHA-256 value can assist in showing whether compared bytes match. It does not by itself establish any of the matters above.

### High-consequence decisions

ECO is not intended or approved for:

- autonomous or materially determinative decisions about benefits, housing, credit, insurance, employment, education, migration, policing, sentencing, essential services or emergency response;
- credibility, fraud, risk, vulnerability or eligibility scoring of a person;
- emotion recognition, biometric categorisation or profiling.

The ECO project will not approve, endorse, supply, promote or support an official ECO deployment for any excluded high-consequence purpose unless a fresh legal and regulatory assessment, accountable organisation, technical controls, contractual allocation, accessibility evidence and independent approval exist. No such deployment is currently approved.

This project control does not purport to regulate an unauthorised independent fork or every act of a third party. It controls official ECO claims, project-endorsed supply, support and deployment decisions.

### Operational and emergency action

ECO is not intended or approved for:

- emergency, crisis or safeguarding response;
- autonomous sending, filing, reporting, contacting or deletion;
- covert monitoring, surveillance or evidence acquisition;
- bypassing security, access controls, legal restrictions or professional duties.

## Claims matrix

### Claims currently permitted

The following statements may be used when accurate in context:

- ECO is under active source development.
- The public repository contains GPL-3.0-only project source.
- There is no approved signed end-user release.
- Development work uses synthetic and non-sensitive test material only.
- ECO is designed around local processing and no required cloud account.
- Current release and deployment gates remain closed.

These statements must not imply that every later candidate or bundled component has already passed the stated design goal.

### Claims permitted only after exact-build evidence

The following require objective evidence tied to the exact released artefact:

- fully offline or zero external network communication;
- one self-contained executable;
- encrypted storage or encrypted metadata;
- source-backed or verified-source output;
- immutable or tamper-evident records;
- privacy-preserving diagnostics;
- accessible, WCAG-conformant or EN 301 549-conformant;
- secure, hardened or penetration-tested;
- production OCR, PDF investigation or reliable AI assistance;
- reproducible build, complete SBOM or complete licence compliance;
- suitable for ordinary users, real evidence or professional work;
- any performance, accuracy, reliability or resource-use figure.

The claim must state the tested version, method, environment, dataset or scope and material limitations.

### Claims prohibited without formal authority

The following must not be used unless a competent, independent authority has actually granted the stated status and the exact scope is identified:

- certified, approved, accredited or regulator-approved;
- legally compliant or compliant with all applicable law;
- medically approved or clinically safe;
- forensic, court-ready, admissible or evidentially conclusive;
- guaranteed accurate, secure, anonymous, private or immutable;
- accessible or disability-compliant as an unqualified absolute;
- public-sector ready, NHS ready or procurement approved;
- EU AI Act compliant, CRA compliant or CE/UKCA marked;
- professional legal, medical, forensic or safeguarding advice.

## Consumer-facing accuracy rule

Public wording must not omit information that would materially change a reasonable person's understanding of ECO's capability, safety, release status or limitations.

Synthetic test results must not be presented as real-world accuracy. A performance statement must identify at least the test set, sample size, method, version and non-generalisation limits.

A disclaimer cannot cure functionality, interface wording or promotional material that objectively presents ECO as having a regulated or professional purpose.

## EU AI transparency assessment

Before any official ECO availability, supply, promotion or deployment in the European Union, the accountable organisation must perform a feature-level and role-specific assessment under Regulation (EU) 2024/1689 as amended by Regulation (EU) 2026/1744 and the European Commission's final Article 50 guidelines published on 20 July 2026.

The assessment must identify:

- whether the future organisation is acting as provider, deployer, importer, distributor or another relevant actor for each route;
- which ECO functions directly interact with natural persons;
- which outputs are generated or manipulated by AI;
- which information, labelling or machine-readable marking obligations apply to each feature;
- any applicable transitional rule or exemption;
- the exact user-interface, export and documentation evidence used to demonstrate the decision.

ECO must clearly identify direct AI interaction wherever Article 50 requires it. No statement of EU AI Act compliance may be made until the exact role, feature scope and implementation evidence are independently approved.

## Foreseeable misuse controls

The product and documentation must anticipate and safely handle requests to:

- obtain a legal conclusion or current-law answer;
- diagnose a condition or recommend treatment;
- generate or paraphrase a summary, chronology, priority or action list from health-related source material;
- interpret health-related source material as a clinical finding or priority;
- decide whether a person is truthful, eligible, dangerous or fraudulent;
- rank or score people for an adverse decision;
- treat OCR or generated text as original evidence;
- use a hash as proof of authenticity or admissibility;
- send or submit generated text without review;
- rely on ECO during an emergency or immediate safeguarding crisis;
- upload full diagnostics or case material to a public support channel.

Where a request falls outside the intended purpose, ECO should state the boundary in plain language, avoid an authoritative answer and provide only a safe, non-professional next step appropriate to the product's role.

## Mandatory reassessment triggers

A fresh legal, regulatory, privacy, security and claims review is required before any of the following is implemented, enabled or promoted:

- cloud, remote API, telemetry, account or developer-access functionality;
- collection of user evidence for support, analytics, training or model improvement;
- any AI-generated or application-generated semantic transformation of health-related source material, including selection based on clinical meaning, omission, clinical reordering, combination, paraphrase, summary, chronology, priority, trend, risk flag or action suggestion;
- diagnosis, treatment, clinical triage, health risk scoring or medical recommendations;
- current-law databases, automated legal conclusions or judicial decision support;
- credibility, eligibility, vulnerability, fraud or risk scoring of people;
- autonomous sending, filing, deletion, contact or submission;
- public-sector, private-institutional, healthcare or justice-sector deployment;
- any EU availability, supply or deployment;
- commercial support, paid distribution or another change that may alter regulatory or open-source treatment;
- a new model, runtime, parser, evidence format, cryptographic design or network architecture;
- a claim of certification, compliance, accessibility, security, accuracy or professional suitability;
- any material expansion of intended users, inputs, outputs or operating environment.

Deterministic exact-text search, byte-preserving display, source navigation and user-requested sorting by explicit non-clinical metadata do not trigger reassessment merely because they match or reorder records. Reassessment is required where the function interprets, selects or transforms health content by meaning.

## Document control and ownership

This document should control the intended-purpose wording used in:

- README and website material;
- application onboarding and help;
- model and system cards;
- release notes and download pages;
- procurement and partner material;
- demonstrations, screenshots and presentations.

Any conflicting statement must be corrected or explicitly identified as historical.

Before public release, an accountable publisher must formally accept ownership of this control, its review cycle and the consequences of expanding the intended purpose. These responsibilities must not be assigned to an individual developer by default.

## Official reference basis

This control was drafted with reference to:

- MHRA, *Crafting an intended purpose in the context of Software as a Medical Device*: https://www.gov.uk/government/publications/crafting-an-intended-purpose-in-the-context-of-software-as-a-medical-device-samd
- MHRA, *Borderline products: how to tell if your product is a medical device*: https://www.gov.uk/guidance/borderline-products-how-to-tell-if-your-product-is-a-medical-device
- MHRA, *Digital mental health technology: qualification and classification*: https://www.gov.uk/government/publications/digital-mental-health-technology-qualification-and-classification
- CMA, *Unfair commercial practices (CMA207)*: https://www.gov.uk/government/publications/unfair-commercial-practices-cma207
- Forensic Science Regulator, *Statutory Code of Practice version 2*: https://www.gov.uk/government/publications/forensic-science-activities-statutory-code-of-practice-version-2
- Regulation (EU) 2024/1689 as amended by Regulation (EU) 2026/1744: https://eur-lex.europa.eu/legal-content/EN/TXT/?uri=CELEX:32026R1744
- European Commission, final *Guidelines on transparency obligations for providers and deployers of AI systems*, 20 July 2026: https://digital-strategy.ec.europa.eu/en/library/guidelines-transparency-obligations-providers-and-deployers-ai-systems

These sources do not by themselves determine ECO's final classification or the applicability of every cited duty. Classification and obligations depend on the actual functionality, claims, distribution model, legal role, users and deployment context at the relevant time.
