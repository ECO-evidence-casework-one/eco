# Privacy and offline-operation policy

ECO is designed to process evidence locally on the user's computer.

## Network behaviour

The current native application source contains no HTTP client, telemetry, analytics, advertising, cloud AI, account system or listening network service.

**This program will not transfer any information to other networked systems unless specifically requested by the user or the person installing or operating it.**

This is a current source-level product rule, not independent qualification of a future packaged executable, embedded runtime or model. Issue #14 requires repeatable packet-capture, DNS and firewall-deny evidence for the exact bundled artefact before public offline/network claims are treated as release evidence. Any local inter-process communication must be documented separately from external transmission.

Future functionality must not weaken this rule. Components requiring online operation are outside ECO's product scope.

## User data

ECO's maintainers do not receive, host, inspect or administer users' casework or evidence. A user controls the local workspace and any backup or export they create.

There is no approved support or diagnostic-export workflow for real case material. Do not assume a diagnostic package is privacy-safe merely because ECO operates locally. Issue #14 requires a redacted support export, category preview, manifest, bounded logs and synthetic disclosure testing before such a workflow can be approved.

## Development reports

Never upload real personal evidence, private correspondence, credentials, vault files, diagnostics, screenshots containing case material or identifying case records to GitHub issues, discussions, pull requests or test fixtures.

Use synthetic, non-sensitive examples only.
