# Historical repository-provenance records

The files in `docs/provenance/historical/` preserve earlier ECO source-development records for audit and chronology.

They are **not** current repository manifests, current build receipts, release receipts, SBOMs, signing records or evidence that the present `main` tree has passed the historical checks.

## V25 N2 P1 records

The following files preserve records previously stored under generic root filenames:

- `historical/ECO_V25_N2_P1_REPOSITORY_MANIFEST_2026-07-31.csv`
- `historical/ECO_V25_N2_P1_REPOSITORY_PREPARATION_RECEIPT_2026-07-31.json`
- `historical/ECO_V25_N2_P1_SBOM_2026-07-31.spdx.json`
- `historical/ECO_V25_N2_P1_THIRD_PARTY_NOTICES_2026-07-31.md`

They relate to:

- build ID `ECO-V25-20260731-N2-P1`;
- preparation time `2026-07-31T09:20:00Z`;
- historical source-development commit `439863de1cc01e1e08064e86ed08deb9b195e54c`;
- source base release `v25.0.0-n1-p3`.

Later commits changed the tracked tree, tests, workflow and public control documents. The historical manifest, receipt, SBOM and notices therefore cannot be used to verify a later commit or executable.

## Current authority

There is no approved current release manifest, release receipt or exact-build SBOM.

P0 issue #15 controls the future actual-build provenance requirement. Any future authoritative manifest must be generated automatically from one exact clean Git tree and must record the exact commit, tree identity, generation time, generator version, tracked-file count and hash for every tracked file. It must reconcile with the exact SBOM, licence notices, build receipt, executable and signing record.

The generic root files `REPOSITORY_MANIFEST.csv` and `REPOSITORY_PREPARATION_RECEIPT.json` are retained only as explicit pointers to their historical records. They are not current integrity evidence.

The generic root `SBOM.spdx.json` is intentionally absent until an exact-build SBOM exists. Current SBOM authority and requirements are recorded in `../SBOM_STATUS.md`.

The root `THIRD_PARTY_NOTICES.md` now records current source-level notice status only. It is not the final notice bundle for a released executable.
