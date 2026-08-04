# SBOM status

ECO does not currently have an authoritative SBOM for a released executable.

The former root `SBOM.spdx.json` described the historical V25 N2 P1 source milestone. It used `filesAnalyzed: false`, did not identify an exact current commit or executable hash, and did not inspect PE imports, resources, sections, overlay data, embedded payload or runtime-created files.

That exact historical SPDX document is preserved at:

`docs/provenance/historical/ECO_V25_N2_P1_SBOM_2026-07-31.spdx.json`

The generic root `SBOM.spdx.json` is intentionally absent until an exact-build SBOM exists. Its absence must not be interpreted as approval to distribute a binary without an SBOM.

P0 issue #15 controls the future authoritative SBOM. Before release, the SBOM must:

- identify the exact source commit, build and executable;
- include the executable hash and exact packaged or redistributed components;
- distinguish redistributed components from operating-system dependencies and build-only tools;
- reconcile one-to-one with the content manifest, licence notices, corresponding source, build receipt and signing record;
- pass independent SPDX validation and actual-binary inspection;
- contain no stale build identity, placeholder source location or unexplained component.

No application, evidence-use, signing or release gate is opened by this status record.
