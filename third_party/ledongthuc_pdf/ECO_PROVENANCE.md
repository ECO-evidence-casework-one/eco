# ECO vendored PDF reader provenance

Upstream: `ledongthuc/pdf`
Upstream commit: `b3c860c2375335b0bc6676c430107a553725991d`
Upstream licence: BSD-3-Clause (`LICENSE` retained verbatim)
Upstream module directive at this commit: `go 1.24.1`
ECO compatibility adjustment: local vendored module declares `go 1.23` only; source files are otherwise copied from the exact upstream commit.

Qualification before integration proved the full upstream test suite and a known-text PDF extraction smoke test under Go 1.23.12. ECO's integration adds its own bounds, panic containment, per-page warnings and page-aware SourceSegment metadata around the donor parser.
