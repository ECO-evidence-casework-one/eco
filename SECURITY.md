# Security policy

## Supported status

ECO is currently an early source-development project. No current build is approved for real or sensitive evidence, ordinary-user distribution or institutional deployment.

## Reporting a vulnerability

Do **not** disclose a suspected security vulnerability in a public issue.

Use GitHub's private vulnerability reporting feature on this repository's **Security** page once it is enabled and confirmed as monitored. Until a tested private route and accountable response owner exist, issue #17 remains a release gate.

Include:

- affected version or exact commit;
- reproducible steps;
- expected and actual behaviour;
- impact assessment;
- a minimal synthetic proof of concept;
- no real personal evidence, private workspace or identifying diagnostic material.

## Security principles and required controls

- fully offline product scope;
- application-preserved originals and fresh verification before downstream use;
- derived data explicitly identified;
- untrusted evidence treated as hostile input;
- no arbitrary code execution by local AI;
- no hidden network fallback;
- authenticated backup validation and independently qualified activation/rollback;
- alias-safe workspace ownership and stale-writer conflict prevention;
- privacy-safe diagnostics;
- release binaries built from public source through a controlled, fail-fast pipeline;
- trusted signatures required before normal end-user distribution;
- actual-build manifest, SBOM, licence and provenance reconciliation;
- accountable security response and continuity ownership before release.

These are required controls, not a certification that current `main` satisfies them. The current release position and outstanding gates are recorded in `CURRENT_STATUS.md` and `docs/control/CURRENT_RELEASE_GATE.md`.

## Coordinated disclosure

No fixed response-time promise is currently made. A public or institutional release requires a named accountable organisation, monitored private route, supported-version policy, advisory process, emergency withdrawal process and continuity plan under issue #17.

Where capacity permits during source development, a valid private report should be acknowledged, investigated with synthetic evidence and followed by a public advisory after a fix or mitigation is available. Real evidence must not be requested or submitted.
