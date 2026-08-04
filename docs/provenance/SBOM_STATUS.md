# SBOM status

ECO does not currently have an authoritative SBOM for a released executable.

The root `SBOM.spdx.json` is retained only as an explicitly labelled **historical V25 N2 P1 source-level record**. Its document name, namespace and comments state that it is not authoritative for current `main`, a current build, an executable, signing or release.

The exact former root bytes are preserved unchanged at:

`docs/provenance/historical/ECO_V25_N2_P1_SBOM_2026-07-31.spdx.json`

The historical document uses `filesAnalyzed: false`, does not identify an exact current commit or executable hash, and does not inspect PE imports, resources, sections, overlay data, embedded payload or runtime-created files.

P0 issue #15 controls the future authoritative SBOM. Before release, the SBOM must:

- identify the exact source commit, build and executable;
- include the executable hash and exact packaged or redistributed components;
- distinguish redistributed components from operating-system dependencies and build-only tools;
- reconcile one-to-one with the content manifest, licence notices, corresponding source, build receipt and signing record;
- pass independent SPDX validation and actual-binary inspection;
- contain no stale build identity, placeholder source location or unexplained component.

The presence of the historical root SPDX document must not be treated as satisfaction of issue #15. No application, evidence-use, signing or release gate is opened by this status record.
