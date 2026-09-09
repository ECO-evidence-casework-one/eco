# AEIB v0.1 — Initial Qualification Receipt R1

**Date:** 21 August 2026

## Local controlled qualification before public staging

The initial AEIB v0.1 generator was deliberately rejected when its same-seed corpus was not byte-deterministic. Root cause was automatically generated MIME multipart boundaries in Python's email serialization.

The generator was corrected to use explicit deterministic MIME boundaries and the entire qualification was rerun.

### Passing corrected result

- generated cases: **22**
- closed-set corpus files including manifest: **23**
- manifest size/SHA verification: **PASS**
- same-seed byte-level tree determinism: **PASS**
- different-seed divergence: **PASS**
- safety assertions (`synthetic_only`, no live malware, no personal data): **PASS**
- private development ZIP integrity: **PASS**

Same-seed tree SHA-256 observed in the controlled local run:

`7b594905acce9cd7b9eff3e30283a67756df136c4e83bd33fb1a81f94121293c`

Different-seed tree SHA-256:

`136c733fbfaf6663cf9126dc9581f43ac209b6c2c39b41b2a1dcd73a79c2e407`

Private development pack SHA-256:

`bef7e0485e4e891fc695d0e484b508228a7e0ce88198605a5461bc55193b1dcc`

## Public evidence rule

The local result is supporting evidence only. The controlling public qualification for this branch should be the GitHub Actions run of `.github/workflows/aeib.yml` against the exact PR checkout.

No generated hostile corpus, executable, model or runtime artifact should be uploaded by the research workflow.
