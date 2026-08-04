# Current ECO project status

**Status date:** 4 August 2026  
**Control update:** approximately 14:28 BST / 15:28 CEST  
**Baseline `main` reviewed for this record:** `9c98588387f5aed6f33371fefbf1eacbc514a5e9`  
**Live canonical source:** the repository's current `main` ref  
**Recorded `VERSION` milestone:** `ECO-V25-20260731-N2-P1`  
**Release position:** source development only; no approved V40 source tag, signed executable or ordinary-user release

## Authority and evidence terminology

This file is ECO's canonical public status summary. Issue #22 carries the live Control Board. Pull requests, issues, raw workflow logs, artifact inventories and dated audit packs preserve supporting evidence.

The baseline SHA identifies the tree reviewed when this record was prepared. It is not a claim that `main` can never advance after the record is merged.

ECO records separately:

- branch head;
- actual tested checkout, including synthetic PR merge refs;
- final merge or squash identity;
- any genuine post-merge workflow;
- reviewer/control-lane relationship.

A green PR check is not exact-head or post-merge evidence unless the raw checkout proves it.

## Live controlled position

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

Issue #53 controls the sequence: repair workspace ownership, implement the native first usable Matter journey, correct M1.18's integration blockers, then connect one controlled assistance seam.

### PR #72 — Workspace Ownership V2

PR #72 is not yet an accepted implementation contract. Its inspection HOLD requires:

1. rewrite OW-01 so stale-CAS defence does not violate the full-lifetime exclusive lease;
2. bind CAS to revision/generation, authenticated metadata digest, audit/state-chain head, target identity and owner transaction;
3. distinguish safe continuation through an already-retained authorised object from forbidden pathname redirection;
4. correct audit terminology: run `30905573823` tested synthetic merge `24241f9addf2e6d5f1d68d721b0c5aa492abf228`, not branch head `a6bc1f08…` directly.

Issue #4 remains fully open. No workspace P0 property is implemented or proved by the documentation-only PR.

### PR #71 — first usable Matter journey

PR #71 provides a materially improved visual direction but is not native implementation, accessibility evidence or release evidence. Its inspection HOLD requires:

- keyboard-operable controls instead of pointer-only cards;
- no dead or misleading visible controls;
- correct modal focus entry, containment and return;
- accessible progress/live-status semantics and narrow/high-zoom resilience;
- exact page/region-aware citations with OCR provenance and navigation instead of filenames alone;
- issue #65 as a hard prerequisite for M1.18;
- core accessibility failures treated as release blockers;
- qualified offline wording until issue #14 passes.

Run `30905512134` tested synthetic merge `6016792666d7ad7d7e8b6413ad27c1213c44a5d0`, not branch head `05983719…` directly.

Issues #7, #8, #14, #65 and #69 remain open.

### M1.18 — issue #65

Phase D is stopped. M1.18 does not yet independently prove erasure of orchestrator-owned buffers; non-cooperative dependencies can defeat deadlines; and late callbacks can make receipt counts incomplete.

No adapter, model, runtime, IPC, Ask ECO, evidence or persistence connection may begin until issue #65 proves unconditional zeroing, truthful receipts, bounded process termination and callback-lifetime control.

## Evidence, AI and high-consequence outputs

Issues #3, #5, #12, #20 and #46 remain open. ECO may be designed to help users navigate supplied evidence, identify document-stated facts and prepare source-linked notes or drafts. It is not approved to act as a doctor, lawyer, forensic authority, emergency service, eligibility scorer or institutional decision maker.

Current Ask ECO and future generated routes are not approved for real or public evidence use.

## Accessibility and exact source navigation

Issue #7 blocks public preview until the core journey works without a mouse, focus remains visible and ordered, visible controls work or are clearly disabled, long content remains reachable and at least one screen reader can identify controls and statuses.

Issue #8 requires citations to the exact supporting page or nearest source region, with OCR provenance/confidence and accessible navigation. A filename alone is insufficient.

The 9 August target in issue #69 cannot waive these requirements.

## Provenance and distribution

P0 issue #15 remains open. Historical/source-level manifest, receipt, SBOM and notice records do not describe an approved current executable.

Issue #24 remains P0. Current workflows do not intentionally upload the runner-built executable, but these historical artifacts remain live and prohibited:

| Artifact ID | Origin run | Expiry UTC |
|---|---|---|
| `8854774165` | `30810944362` | 10 August 2026 11:49:38 |
| `8856536245` | `30815339549` | 10 August 2026 12:54:07 |
| `8863951645` | `30833597696` | 10 August 2026 16:47:05 |
| `8865678638` | `30838068198` | 10 August 2026 17:45:50 |

Do not download, execute, test or redistribute them. A one-time control task is scheduled for 10 August 2026 at 19:00 Europe/London to verify deletion or expiry.

## Publisher, stewardship and outreach

The preferred model is one accountable official ECO publisher/steward that may use contracted, controlled and replaceable specialists while retaining final responsibility.

No organisation is appointed, shortlisted or authorised for outreach. Open Knowledge Foundation's deeper public-source review remains `HOLD` and creates no relationship or permission to contact it. Issue #17 remains open and blocks an official source or binary release until a named legal organisation accepts and proves the required duties.

No role falls to the originating developer or another contributor merely through contribution. Responsibilities may still arise from actual publishing, signing, contracting, supply, data processing, public claims or applicable law.

## Stop gates

The following remain blocked:

- real, sensitive or irreplaceable evidence use;
- current Ask ECO or local-model use with public or real evidence;
- a public V40 tag or pre-alpha source release;
- ordinary-user executable testing;
- public executable upload or historical-artifact redistribution;
- signing, release-candidate or stable-release status;
- claims of production OCR, reliable generative AI, complete source investigation, accessibility, security, legal, medical, forensic or regulatory compliance;
- official institutional, healthcare, justice-sector or EU supply or deployment;
- named organisational outreach without a separately reviewed target and public contact route.

No deadline, label, disclaimer or risk-acceptance statement can turn a failed gate into a pass.

## Next controlled sequence

1. Correct PR #72 at a new frozen head and re-review it.
2. Correct PR #71 at a new frozen head and re-review it.
3. Implement workspace ownership in small current-main slices with real subprocess evidence.
4. Implement the native Matter journey without M1.18.
5. Correct issue #65 before controlled assistance integration.
6. Recheck the historical artifacts after their expiry window.
7. Freeze and qualify one exact V40 source candidate only after every applicable issue #69 gate passes.

## Public-record rule

Use synthetic and non-sensitive information only. Do not publish personal evidence, private workspaces, credentials, private diagnostics, exploit-level instructions, unapproved executables or model files.
