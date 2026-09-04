# AEIB v0.1 Status

**State:** research draft; not merged into `main`; not a product release.

Current scope:

- 22 safe deterministic synthetic fixture families/cases;
- standard-library generator;
- closed-set SHA/size verifier;
- same-seed determinism gate;
- different-seed divergence gate;
- dedicated GitHub Actions research qualification with no artifact upload.

Current non-claims:

- AEIB does not certify software as secure;
- AEIB does not yet represent the full historical Dossier Forge hostile campaign;
- no external project adoption is claimed;
- no maintainer endorsement is claimed;
- no CVE/GHSA or third-party vulnerability disclosure is claimed.

Next research step after exact-head CI qualification is to define a minimal adapter interface and run AEIB against an owned ingestion implementation, preserving PASS, FAIL, PARTIAL and BLOCKED results without tuning the benchmark to the target.
