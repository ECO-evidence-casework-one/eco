# Current ECO project status

**Status date:** 4 August 2026  
**Control update:** approximately 13:49 BST / 14:49 CEST  
**Canonical `main`:** `3c7c69586cac195d146188e6b914db12f6391815`  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Release position:** source development only; no approved V40 source tag, signed executable or ordinary-user release

## Status authority

This file is ECO's canonical public status summary. Issue #22 carries the live Control Board. Pull requests, issues, raw workflow logs, artifact inventories and dated audit packs preserve supporting evidence.

A pull-request workflow commonly tests a synthetic merge rather than the branch head. ECO therefore records branch head, tested checkout, final merge identity and any post-merge run separately. A green badge alone is not exact-head or post-merge evidence.

## Live position

| Area | Current position |
|---|---|
| Workspace repair | Draft PR #72 at `a6bc1f0898d529b5f9eebab76f757e0926f85f86`; documentation-only; inspection HOLD |
| First usable Matter journey | Draft PR #71 at `05983719b29be02f44ef2b4e7ec09a8166514aa6`; synthetic design prototype only; inspection HOLD |
| Superseded workspace branch | PR #11 is closed unmerged and must not be revived or merged wholesale |
| M1.18 | Merged as isolated source at `b66cab51ad0da118303afd7009065e25a27e6ab7`; P0 issue #65 blocks every integration |
| Historical executable exposure | Four unsigned Actions artifacts remain live; issue #24 remains P0 |
| Publisher/steward | Preferred model is one accountable steward with controlled replaceable specialists; no organisation appointed |
| Candidate research | OKFN deep review completed; full/near-full status remains `HOLD`; no shortlist or outreach authority |
| Real or sensitive evidence | Blocked |
| Signed public executable | None |
| Institutional, healthcare, justice-sector and EU availability | Blocked |

## Development sequence

Issue #53 controls the development sequence: repair the workspace foundation, deliver the first native Matter journey, correct M1.18's integration blockers, then connect one controlled assistance seam.

### PR #72 — Workspace Ownership V2

PR #72 is a current-main replacement direction for stale PR #11, but its two architecture/test documents are not yet accepted as the controlling contract.

The inspection HOLD requires:

1. rewrite OW-01 so a stale-CAS scenario does not let another process reopen while the first still owns the full-lifetime writable lease;
2. bind CAS to the exact authenticated state, including revision/generation, metadata digest, audit/state-chain head, target identity and owner transaction;
3. distinguish safe operation through an already-retained authorised object from forbidden redirection through a replaced pathname;
4. correct audit terminology: run `30905573823` tested synthetic merge `24241f9addf2e6d5f1d68d721b0c5aa492abf228`, not branch head `a6bc1f08…` directly.

Issue #4 remains fully open. No workspace P0 property is implemented or proved by PR #72.

### PR #71 — first usable Matter journey

PR #71 provides a materially improved visual direction but is not native implementation, accessibility evidence or release evidence.

The inspection HOLD requires:

- keyboard-operable controls instead of pointer-only Matter cards;
- removal, implementation or explicit disabling of dead and misleading controls;
- correct modal focus entry, containment and return;
- accessible progress/live-status semantics and narrow/high-zoom resilience;
- exact page/region-aware citations with OCR provenance and source navigation instead of filenames alone;
- issue #65 as a hard prerequisite for any M1.18 connection;
- core accessibility failures treated as release blockers;
- qualified offline wording until issue #14 passes.

Run `30905512134` tested synthetic merge `6016792666d7ad7d7e8b6413ad27c1213c44a5d0`, not branch head `05983719…` directly.

Issues #7, #8, #14, #65 and #69 remain open.

### M1.18 integration — issue #65

Phase D is stopped. M1.18 currently trusts an injected eraser's success return without independently proving that orchestrator-owned buffers were zeroed. Non-cooperative synchronous dependencies can defeat deadlines, and late callbacks can make receipt counts incomplete.

No adapter, model, runtime, IPC, Ask ECO, evidence or persistence connection may begin until issue #65 proves unconditional zeroing, truthful receipts, bounded process termination and callback-lifetime control.

## Evidence, AI and high-consequence output

- Issue #3 remains open for preserved-source truth and depends in part on issue #12.
- Issue #12 remains open for bounded verification cost, restore/Ask serialisation and persistence-error propagation.
- Issue #5 remains open for the offline AI controller.
- Issue #20 remains P0 for clinical, professional, forensic, profiling, scoring and authority-side adverse-decision boundaries.
- Issue #46 remains open for consistent UI, model-facing, export and public-claims conformance.

ECO may be designed to help users navigate supplied evidence, identify document-stated facts and prepare source-linked notes or drafts. It is not approved to act as a doctor, lawyer, forensic authority, emergency service, eligibility scorer or institutional decision maker.

Current Ask ECO and future generated routes are not approved for real or public evidence use.

## Accessibility and source navigation

Issue #7 blocks public preview until the core journey works without a mouse, focus remains visible and ordered, visible controls work or are clearly disabled, long content remains reachable and at least one screen reader can identify controls and statuses.

Issue #8 requires citations to the exact supporting page or nearest source region, with OCR provenance/confidence and accessible navigation. A filename alone is insufficient.

The 9 August target in issue #69 cannot waive these requirements.

## Provenance and distribution

P0 issue #15 remains open. Historical/source-level manifest, receipt, SBOM and notice records do not describe an approved current executable.

Before executable distribution, one exact final file must reconcile with source identity, packaged-content manifest, actual-build SBOM, licences/notices, runtime/model provenance, clean-machine and reproducibility evidence, hash, trusted signature and release receipt.

Issue #24 remains P0. Current workflows do not intentionally upload the runner-built executable, but these historical artifacts remain live and prohibited:

| Artifact ID | Origin run | Expiry UTC |
|---|---|---|
| `8854774165` | `30810944362` | 10 August 2026 11:49:38 |
| `8856536245` | `30815339549` | 10 August 2026 12:54:07 |
| `8863951645` | `30833597696` | 10 August 2026 16:47:05 |
| `8865678638` | `30838068198` | 10 August 2026 17:45:50 |

Do not download, execute, test or redistribute them. A one-time control task is scheduled for 10 August 2026 at 19:00 Europe/London to verify deletion or expiry.

## Publisher, stewardship and outreach

The preferred model is one accountable official ECO publisher/steward that may use contracted, controlled and replaceable specialists while retaining final responsibility for release, withdrawal, intended purpose, security incident command, signing, continuity, complaints, data roles, contracts, liability, insurance and regulatory decisions.

A fragmented model with no accountable lead publisher is rejected.

No organisation is appointed, shortlisted or authorised for outreach. Current candidate research is provisional public-source research only and does not show interest, consent, suitability or willingness.

Open Knowledge Foundation has received a deeper public-source review and remains `HOLD` for a full or near-full role. The review creates no shortlist, relationship or permission to contact it. Issue #17 remains open until a named legal organisation formally accepts and proves the required duties.

No role falls to the originating developer or another contributor merely through contribution. Responsibilities may still arise from actual publishing, signing, contracting, supply, data processing, public claims or applicable law.

## Current stop gates

The following remain blocked:

- real, sensitive or irreplaceable evidence use;
- current Ask ECO or local-model use with public or real evidence;
- public V40 tag or pre-alpha source release while issue #69 gates remain open;
- ordinary-user executable testing;
- public executable upload or redistribution of historical artifacts;
- signing, release-candidate or stable-release status;
- claims of production OCR, reliable generative AI, complete source investigation, accessibility, security, legal, medical, forensic or regulatory compliance;
- official institutional, healthcare, justice-sector or EU supply or deployment;
- named organisational outreach without a separately reviewed target and public contact route.

No deadline, label, disclaimer or risk-acceptance statement can turn a failed gate into a pass.

## Next controlled sequence

1. Correct PR #72 at a new frozen head and re-review it before implementation relies on it.
2. Correct PR #71 at a new frozen head and re-review its controls, citations and accessibility rules.
3. Implement the workspace P0 repair in small current-main slices with real subprocess evidence.
4. Implement the native first usable Matter journey without M1.18.
5. Correct issue #65 before controlled assistance integration.
6. Recheck the historical artifacts after their expiry window.
7. Freeze and qualify one exact V40 source candidate only after every applicable issue #69 gate passes.

## Public-record rule

Use synthetic and non-sensitive information only. Do not publish personal evidence, private workspaces, credentials, private diagnostics, exploit-level instructions, unapproved executables or model files.