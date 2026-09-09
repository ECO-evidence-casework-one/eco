# ECO FOSS Donor Register — Wave 2

Date: 2026-09-04

Purpose: close the remaining fuzzing, corruption and hostile-input testing gap using GitHub/FOSS-first components.

| Project | Upstream | Licence | ECO role | Boundary |
|---|---|---|---|---|
| Syft | `anchore/syft` | Apache-2.0 | SBOM engine source/provenance mirror | build/release tool; already integrated by pinned binary |
| go-fuzz | `dvyukov/go-fuzz` | Apache-2.0 | Go fuzz harness/reference patterns | development/test only; Go native fuzzing remains preferred where sufficient |
| OSS-Fuzz | `google/oss-fuzz` | Apache-2.0 | mature fuzz harness/corpus/failure-handling patterns | study/adapt testing architecture only |
| SQLite | `sqlite/sqlite` | Public Domain | corruption, fault-injection and recovery test patterns | study/adapt; preserve SQLite public-domain notices/blessing text |
| go-message | `emersion/go-message` | MIT | MIME/email parsing and hostile-message test inputs | direct Go library candidate after API/dependency review |
| go-mbox | `emersion/go-mbox` | MIT | MBOX parsing and hostile mailbox test inputs | direct Go library candidate after API/dependency review |
| archives | `mholt/archives` | MIT | archive/ZIP format handling and adversarial archive patterns | direct Go library candidate after dependency/sandbox review |
| AFL++ | `AFLplusplus/AFLplusplus` | AGPL-3.0 | native external fuzzing tool | EXTERNAL DEVELOPMENT TOOL ONLY; do not copy/link/embed AFL++ code into ECO without a separate licence decision |

## Rules

- Downloading/source inspection is permitted for all entries.
- Permissive/public-domain donors may be adapted only with their notices/licence conditions preserved.
- AFL++ is intentionally isolated as an external development tool because its current root licence is AGPL-3.0.
- This wave does not approve automatic network access, cloud fuzzing, or hosted corpus services in ECO.
- Exact acquired commits, licence snapshots and source archive hashes must be recorded in the harvest artifact.
