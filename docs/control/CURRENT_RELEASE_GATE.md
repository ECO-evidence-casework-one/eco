# Current ECO release gate

**Gate record:** `ECO-RELEASE-GATE-20260806-012`  
**Updated:** 6 August 2026, approximately 10:47 BST / 11:47 CEST  
**Canonical public status:** [`../../CURRENT_STATUS.md`](../../CURRENT_STATUS.md)  
**Baseline `main` reviewed for this record:** `9da31b714bb45007c71f187e19793b853f521b09`  
**Live canonical source:** the repository's current `main` ref  
**Recorded public `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Approved public source release:** none  
**Approved signed end-user executable:** none

## Controlling decision

ECO remains a source-development project with private, unsigned, synthetic-only recovery testing.

No branch, issue, workflow, private candidate, synthetic prototype, runner-built executable, governance record, deadline or audit pack currently authorises:

- real, sensitive or irreplaceable evidence;
- ordinary-user testing;
- a public executable, model or runtime artifact;
- signing, release-candidate or stable-release status;
- institutional, healthcare, justice-sector or authority-side decision use;
- Northern Ireland or EU/EEA supply;
- unsupported claims of security, privacy, accessibility, legal/medical/forensic accuracy or production readiness.

## Application and source-lineage gate

The current private recovery direction is controlled as follows:

- V38 M2 P1.1 is the application and interface parent.
- V39 exact and reconstructed components must be identified accurately.
- Reconstructed V39 M1.4-M1.16 material must not be described as byte-identical historical recovery.
- The V40 interface/application shell is rejected as an application parent.
- V40 model/runtime/OCR/PDF/packaging evidence may be reused only component by component after exact provenance, licensing and integration review.
- A private candidate cannot become a public baseline merely because it compiled, ran or bundled a large runtime.

Any candidate that replaces the V38 application parent with the V40 shell, relabels reconstructed source as exact, or silently reuses obsolete SBOM/notices fails this gate.

## Gate matrix

| Gate | Current position |
|---|---|
| Preserved evidence/source truth | Issue #3 open |
| Workspace ownership and clean state | Issue #4 open; PR #72 incomplete for the full application lifecycle |
| Offline AI controller | Issue #5 open |
| Responsiveness/cancellation | Issue #6 open |
| Accessibility | Issue #7 blocks public preview |
| Exact page/region source navigation | Issue #8 open |
| Ask/restore verification consistency | Issue #12 open |
| Diagnostic privacy/offline claims | Issue #14 open |
| Actual-build provenance | Issue #15 P0 open |
| Intended purpose and public claims | Issues #16, #20, #23 and #46 open |
| Publisher/steward | Issue #17 open; no organisation appointed |
| Public runnable artifacts | Issue #24 P0 open pending controlled expiry verification |
| Former V40 deadline | Issue #69 superseded; no active release authority |
| M1.18 bounded repair | Issue #65 closed for its defined technical proof only |

## M1.18 integration status

Issue #65 is closed as technically completed for the bounded repair and recovered V38 worker-termination proof.

- PR #81 preserved strengthened erasure and dependency-lifetime evidence and is closed unmerged.
- PR #82 preserved the real recovered V38 Windows process-termination proof and is closed unmerged.
- No executable, model or runtime artifact was uploaded by those proof lanes.
- No real evidence was used.

Issue #65 closure does **not** establish:

- complete M1.14-M1.17 Windows containment or authenticated-IPC backends;
- application-level Ask ECO approval;
- real-evidence suitability;
- public release, signing or deployment approval.

Any application connection must still prove the exact candidate's output suppression, worker ownership, cancellation, persistence, source binding and prohibited-output behaviour.

## Public pull-request classification

### PR #72 — workspace ownership primitive lane

PR #72 remains an open draft evidence/implementation lane for isolated workspace-ownership primitives. It is not the private V38 recovery application parent and does not close issue #4.

No public release or real-evidence decision may rely on PR #72 until the full Vault lifecycle, restore, revision/CAS, rollback, migration and clean-state controls are integrated and independently qualified.

### PR #71 — historical V40 visual prototype

PR #71 is historical visual-design evidence only.

It is not:

- native product implementation;
- accessibility evidence;
- persistence evidence;
- exact source-navigation evidence;
- an application parent;
- release authority.

Any useful hierarchy or terminology must be reimplemented and tested in the real V38-derived application.

### PR #79 — historical/rejected V40 application shell

PR #79 is historical/rejected application-shell evidence.

The shell and its executable are not an application or release baseline. Only individually verified component evidence may be salvaged, with a new candidate-specific SBOM, notices, corresponding source, payload manifest, network evidence and runtime-minimisation record.

### PRs #80-#82

- PR #80 is closed unmerged and superseded fallback-shell evidence.
- PR #81 is closed unmerged M1.18 repair evidence.
- PR #82 is closed unmerged recovered V38 Windows worker-proof evidence.

None authorises a public artifact or product release.

## Superseded V40 target

Issue #69's former 9 August V40 public-pre-alpha target is superseded.

The historical acceptance themes remain useful where applicable, but the issue cannot:

- revive the V40 shell;
- impose a deadline on the V38 recovery lane;
- waive an open gate;
- authorise a source tag, binary, model or runtime.

A future release assessment must name one exact candidate, source identity, artifact hash, territory, publisher and distribution route.

## Evidence, AI and high-consequence gate

ECO may assist a user to organise supplied documents, identify document-stated facts, navigate sources and prepare user-controlled notes or drafts.

The exact candidate must block or safely bound requests to:

- diagnose, prescribe, triage or assess clinical risk;
- provide definitive legal advice or representation;
- authenticate evidence or guarantee admissibility;
- decide entitlement or eligibility;
- score credibility, honesty, dangerousness or fraud risk;
- make or materially influence an official adverse decision;
- report an external action without a deterministic application receipt;
- follow instructions embedded in imported evidence.

Original source, extraction/OCR, generated suggestion, user confirmation and user note must remain visibly and persistently distinguishable.

## Privacy, diagnostics and offline gate

No public absolute claim of privacy, offline operation, no cloud, no telemetry or security is approved until the exact final payload is independently tested.

The candidate must return:

- complete data-flow statement;
- support-diagnostic example and redaction manifest;
- evidence that diagnostics exclude matter content by default;
- complete networking-disabled journey;
- DNS and connection-attempt observations;
- explanation of any loopback IPC;
- payload inventory identifying network- or RPC-capable components;
- removal or strict justification of unnecessary network/RPC libraries.

Accounts, cloud AI, cloud sync, telemetry, automatic crash upload, remote evidence support or publisher-controlled workspaces trigger a new privacy, security and intended-purpose assessment.

## Actual-build licensing and provenance gate

Issue #15 remains P0.

Every exact serious candidate must include:

1. executable and source filenames, sizes and SHA-256 values;
2. source commit or exact source identity;
3. complete extracted payload manifest with hashes;
4. actual-build SPDX SBOM;
5. exact component versions, source identities and dependency relationships;
6. complete third-party notices and licence texts;
7. corresponding source for conveyed GPL/LGPL components;
8. Qwen upstream and quantisation provenance;
9. llama.cpp exact commit/build/toolchain record;
10. Tesseract, Leptonica, OCR-data and Poppler provenance;
11. reproducibility records for ECO-trained models;
12. runtime-minimisation record;
13. accessible offline licence viewer;
14. clean-machine build and launch evidence.

Old V25/V36 SBOMs and notices are historical only and must not be relabelled as current.

## Accessibility and source-navigation gate

Issue #7 blocks public preview until the exact candidate proves:

- keyboard-only completion of the core journey;
- visible, logical focus and modal focus return;
- meaningful accessible names, states, errors and progress;
- usable layout and scrolling at the supported minimum display and 100%, 150% and 200% scaling;
- at least one screen-reader result;
- no dead or misleading visible controls.

Issue #8 requires usable page/region-aware source navigation. Filename-only citations are insufficient. OCR-derived text and approximate locations must be labelled truthfully.

## Public runnable artifacts — issue #24

Current workflows must not intentionally publish an executable, DLL, installer, model, runtime or runnable archive while the binary gate is closed.

The historical artifacts recorded in issue #24 must not be downloaded, executed, tested or redistributed. The controlled recheck after their recorded expiry window may inspect metadata only.

Issue #24 also remains open for private test-handoff controls and future release automation gated by issue #15, issue #17, trusted signing and explicit approval.

## Publisher and organisational gate

The preferred future model is one accountable incorporated publisher/steward that may use controlled and replaceable specialists while retaining final responsibility.

No organisation is appointed, shortlisted or authorised for outreach.

Issue #17 blocks any official source or binary release until a named legal organisation formally accepts and proves:

- release and withdrawal authority;
- signing and key continuity where applicable;
- vulnerability intake and incident response;
- support and complaints boundaries;
- privacy and accessibility responsibilities;
- repository and asset recovery;
- contracts, liability and insurance;
- supported-version and end-of-support decisions;
- territorial and regulatory ownership.

No organisational duty falls to the originating developer or another contributor merely through contribution.

## Route-specific decisions

### Private synthetic development

Currently authorised subject to exact identity, synthetic evidence, private delivery, no public artifact and honest limitations.

### Private external synthetic qualification

Requires named recipient, purpose, exact hashes, no onward distribution, expiry/withdrawal status, controlled diagnostics and no real evidence.

### Public source readiness

Not currently approved. Requires a clean exact source package, lineage review, complete source licensing/notices, no private material, no official executable/model/runtime and no product-readiness or support claim.

### Official Great Britain executable

Blocked pending an incorporated publisher, exact candidate PASS, signing governance, complete actual-build legal package, security/support/accessibility/privacy operations, insurance and release-specific legal advice.

### Northern Ireland, EU/EEA and institutional routes

Blocked pending separate route-specific assessment. They cannot inherit the private-development or Great Britain decision.

## Automatic rejection conditions

Reject a candidate or proposal if any of the following occurs:

- V40 shell or another reconstruction is presented as the V38 application parent;
- reconstructed V39 material is described as exact historical source;
- an old SBOM or notice set is relabelled as current;
- real evidence is used without a separate approved gate;
- an executable, model or runtime is publicly published without authority;
- failed or incomplete imports become searchable or citable;
- a material evidence fact is invented and presented as established;
- a citation opens the wrong source or materially wrong location;
- imported evidence controls ECO through prompt injection;
- information leaks between matters, logs or diagnostics;
- old workspace state loads silently;
- networking or a remote service is required for the claimed local journey;
- ECO diagnoses, prescribes, decides entitlement or scores credibility;
- original evidence is altered or deleted on model instruction;
- Karl or another individual contributor is made the default publisher, signer, contractor, support operator or evidence custodian.

## Next gate actions

1. Continue private V38-derived recovery development using synthetic evidence only.
2. Return one exact serious candidate with the complete source, provenance, network, evidence-integrity, AI-boundary, accessibility and claims package.
3. Assess the candidate against all open controls.
4. Recheck issue #24's historical artifacts after their recorded expiry window without downloading or executing them.
5. Open no public, institutional or territorial route without a separate exact-candidate decision.

## Public-record rule

Use synthetic and non-sensitive information only. Do not publish personal evidence, private workspaces, credentials, private diagnostics, exploit-level instructions, unapproved executables, models or runtime payloads.
