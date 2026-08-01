# Current ECO release gate

**Gate record:** ECO-RELEASE-GATE-20260801-001  
**Updated:** 1 August 2026  
**Current public source:** `ECO-V25-20260731-N2-P1`  
**Current development activity:** V34 underway  
**Signed end-user release:** None

## Gate status

| Gate | Current position |
|---|---|
| Candidate identity | Awaiting V34 |
| Source and provenance | Awaiting V34 |
| Runtime and model supply chain | Blocked until exact runtime verification is supplied |
| One-file packaging | Blocked until the actual embedded executable is supplied |
| Windows stability and layout | Not started for V34 |
| Privacy and zero-network proof | Not started for V34 |
| Evidence ingestion | Not started for V34 |
| OCR and local AI | Not started for V34 |
| Accessibility | Not started for V34 |
| Security and recovery | Incomplete |
| Publisher and stewardship | No accountable operating organisation appointed |
| Trusted signing | Not approved |

## Stop rules

The control room must stop before Windows execution where:

- an executable runtime is not pinned and verified before use;
- the model identity or licence is incomplete;
- the source manifest or SBOM is stale;
- the supplied EXE is not the actual one-file deliverable;
- the candidate depends on developer-only folders or separately managed assets;
- source and executable cannot be independently reconciled.

The control room must stop before public beta where:

- accessibility evidence is incomplete;
- no accountable steward accepts support, complaints, security and privacy responsibilities;
- real-evidence use has not received external security, accessibility, privacy and legal review.

The control room must stop before GitHub Release where:

- the binary is unsigned;
- exact source and commit are absent;
- Smart App Control testing has not passed;
- any P0 or P1 finding remains unresolved.
