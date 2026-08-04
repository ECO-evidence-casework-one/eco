# Current ECO project status

**Status date:** 4 August 2026  
**Control update:** 13:49 BST / 14:49 CEST  
**Canonical `main` at this checkpoint:** `90825496cda639ef6be8c052a8ad9f2ebe2f1464`  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Release position:** source development only; no approved V40 tag, signed executable or ordinary-user release

## Status authority and evidence terminology

This file is ECO's canonical public status summary. The live Control Board in issue #22 carries operational detail. Issues, pull requests, raw workflow logs, artifact inventories and dated audit packs preserve supporting evidence.

ECO distinguishes:

- **branch head** — the proposed source commit;
- **tested PR merge tree** — a synthetic pull-request merge checked out by Actions;
- **merged-tree verification** — a separate comparison of the final merge/squash tree;
- **post-merge CI** — a workflow that actually ran on the final merged commit.

A green PR check is not exact-head or post-merge evidence unless the raw checkout proves it.

## Current live state

| Area | Current position |
|---|---|
| Canonical source | `main` at `90825496cda639ef6be8c052a8ad9f2ebe2f1464` |
| Workspace repair | Draft PR #72 at `a6bc1f0898d529b5f9eebab76f757e0926f85f86`; documentation-only; inspection HOLD |
| First usable Matter journey | Draft PR #71 at `05983719b29be02f44ef2b4e7ec09a8166514aa6`; synthetic design-only prototype; inspection HOLD |
| Superseded workspace branch | PR #11 is closed unmerged and must not be revived or merged wholesale |
| M1.18 orchestrator | Merged as `b66cab51ad0da118303afd7009065e25a27e6ab7`; isolated source only |
| M1.18 integration | P0 issue #65 blocks every adapter, model, runtime, IPC, Ask ECO, evidence and persistence connection |
| Public executable artifacts | Four historical unsigned Actions artifacts remain live; issue #24 remains P0 |
| Publisher/steward | Preferred model selected: one accountable official steward with controlled replaceable specialists; no organisation appointed or authorised for outreach |
| Real or sensitive evidence | Blocked |
| Signed public executable | None |
| Institutional, healthcare, justice-sector and EU availability | Blocked |

## Controlling development sequence

Issue #53 controls development balance and recovery.

### Phase B — workspace integrity

PR #72 is a current-main replacement direction for stale PR #11, but its two architecture/test documents are not yet accepted as a controlling implementation contract.

The inspection HOLD requires correction of:

1. **OW-01 contradiction** — a stale-CAS scenario cannot let Process A reopen while Process B still holds the full-lifetime exclusive writable lease;
2. **incomplete CAS identity** — the expected state must bind revision/generation, authenticated metadata digest, audit/state-chain head, target object identity and owner transaction;
3. **substitution semantics** — tests must distinguish safe continuation through an already-retained authorised object from forbidden redirection through a replacement pathname;
4. **audit classification** — run `30905573823` tested synthetic PR merge `24241f9addf2e6d5f1d68d721b0c5aa492abf228`, not branch head `a6bc1f08…` directly.

Issue #4 remains fully open. No workspace P0 property is implemented or proved by the documentation-only PR.

### Phase C — first usable Matter journey

PR #71 provides a materially improved visual direction, but it is not native application implementation, accessibility evidence or release evidence.

The inspection HOLD requires correction of:

- pointer-only Matter cards and other keyboard failures;
- visible dead controls and settings that imply behaviour they do not provide;
- missing modal focus entry, containment and return;
- incomplete progress/live-region semantics and narrow/high-zoom resilience;
- filename-only Ask ECO citations instead of exact page/region-aware source navigation;
- omission of issue #65 as a hard prerequisite for M1.18 integration;
- release wording that could narrow rather than block core accessibility failures;
- unrestricted offline wording before issue #14 exact-runtime proof.

Run `30905512134` tested synthetic PR merge `6016792666d7ad7d7e8b6413ad27c1213c44a5d0`, not branch head `05983719…` directly.

Issues #7, #8, #14, #65 and #69 remain open.

### Phase D — controlled assistance integration

Phase D is stopped by issue #65.

M1.18 currently trusts an injected eraser's success return without independently proving that orchestrator-owned buffers were zeroed. Synchronous non-cooperative dependencies can also defeat deadlines, and late callbacks can make final receipt counts incomplete.

No M1.18 import or adapter work may begin until issue #65 passes its erasure, dependency-lifetime, callback-lifetime and process-termination acceptance tests.

## Evidence, AI and high-consequence output gates

- Issue #3 remains open for preserved-source truth and depends in part on issue #12.
- Issue #12 remains open for bounded verification cost, restore/Ask serialisation and persistence-error propagation.
- Issue #5 remains open for the offline AI controller.
- Issue #20 remains P0 for diagnosis, treatment, clinical-risk, professional-representation, forensic-guarantee, profiling, scoring and authority-side adverse-decision boundaries.
- Issue #46 remains open for consistent UI, model-facing, export and public-claims conformance.

ECO may be designed to help users navigate and organise supplied evidence, identify document-stated facts and prepare source-linked notes or drafts. It is not approved to act as a doctor, lawyer, forensic authority, emergency service, eligibility scorer or institutional decision maker.

Current Ask ECO and future generated routes are not approved for real or public evidence use.

## Accessibility and exact source navigation

Issue #7 blocks public preview until the core journey is usable without a mouse, focus remains visible and ordered, every visible control works or is clearly disabled, long content remains reachable, and at least one screen reader can identify controls and statuses.

Issue #8 requires stable citations to the exact supporting page or nearest source region, including OCR provenance/confidence and accessible source-location navigation. A filename alone is insufficient.

The 9 August target in issue #69 cannot waive these requirements.

## Provenance, packaging and distribution

P0 issue #15 remains open. Historical/source-level manifest, receipt, SBOM and notice records do not describe an approved current executable.

Before any executable distribution, one exact final file must reconcile with immutable source identity, packaged-content manifest, actual-build SBOM, licences/notices, runtime/model provenance, clean-machine evidence, reproducibility evidence, hash, trusted signature and release receipt.

Issue #24 remains P0. Current workflows do not intentionally upload the runner-built executable, but these historical artifacts remain live and prohibited:

| Artifact ID | Origin | Expiry UTC |
|---|---|---|
| `8854774165` | former PR #19 run `30810944362` | 10 August 2026 11:49:38 |
| `8856536245` | former PR #11 run `30815339549` | 10 August 2026 12:54:07 |
| `8863951645` | former PR #18 run `30833597696` | 10 August 2026 16:47:05 |
| `8865678638` | former PR #18 run `30838068198` | 10 August 2026 17:45:50 |

Do not download, execute, test or redistribute them. A one-time control task is scheduled for 10 August 2026 at 19:00 Europe/London to verify deletion or expiry.

## Publisher, stewardship and outreach

The preferred future model is one clearly identified accountable official ECO publisher/steward that may use contracted, controlled and replaceable specialists while retaining final responsibility for release, withdrawal, intended purpose, security incident command, signing, continuity, complaints, data roles, contracts, liability, insurance and regulatory decisions.

A fragmented model with no accountable lead publisher is rejected.

No organisation is appointed, shortlisted or authorised for outreach. Current candidate research is provisional public-source research only and does not show interest, consent, suitability or willingness. Issue #17 remains open until a named legal organisation formally accepts and proves the required duties.

No role falls to the originating developer or another contributor merely through contribution. Responsibilities may still arise from actual publishing, signing, contracting, supply, data processing, public claims or applicable law.

## Current stop gates

The following remain blocked:

- real, sensitive or irreplaceable evidence use;
- current Ask ECO or local-model use with public or real evidence;
- ordinary-user executable testing;
- public V40 source tag or pre-alpha release while issue #69 gates remain open;
- any public executable upload or redistribution of historical artifacts;
- signing or release-candidate status;
- claims of production OCR, reliable generative AI, complete source investigation, accessibility, security, legal, medical, forensic or regulatory compliance;
- official institutional, healthcare, justice-sector or EU supply or deployment;
- named organisational outreach without a separately reviewed target and contact route.

No deadline, label, disclaimer or risk-acceptance statement can turn a failed gate into a pass.

## Next controlled sequence

1. Correct PR #72's ownership/CAS contract at a new frozen head and re-review it before implementation relies on it.
2. Correct PR #71's accessibility, source-navigation, issue #65 and release-matrix defects at a new frozen head.
3. Implement the workspace P0 repair in small current-main slices with real subprocess evidence.
4. Implement the native first usable Matter journey without connecting M1.18.
5. Correct issue #65 before any controlled assistance adapter work.
6. Recheck the four historical artifacts after their expiry window.
7. Freeze and qualify one exact V40 source candidate only after every applicable issue #69 gate passes.

## Public-record rule

Public repository records must use synthetic and non-sensitive information only. Do not publish personal evidence, private workspaces, credentials, private diagnostics, exploit-level instructions, unapproved executables or model files.