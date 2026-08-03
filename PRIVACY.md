# Privacy and offline-operation policy

ECO is designed to process evidence locally on the user's computer.

## Network behaviour

The current native application source contains no HTTP client, telemetry, analytics, advertising, cloud AI, account system or listening network service.

**This program will not transfer any information to other networked systems unless specifically requested by the user or the person installing or operating it.**

Future functionality must not weaken this rule. Components requiring online operation are outside ECO's product scope.

## User data

ECO's maintainers do not receive, host, inspect or administer users' casework or evidence. A user controls the local workspace and any backup or export they create.

Casework content, including evidence, conversations, settings, workspace names and creation details, is kept in the encrypted workspace. ECO retains only minimal local routing information in plaintext: workspace format, opaque workspace ID, development kind, schema and exact candidate identity; candidate app-state also contains candidate/build identity and opaque hash-chained action audit fields without workspace names or full paths. During an unfinished migration or portable restore, an authenticated plaintext recovery record temporarily contains canonical transaction paths, opaque identities, a random nonce, current phase and start time so recovery can fail closed. Migration also records its build/schema transition; restore records build/candidate/schema and the encrypted-backup SHA-256. Empty lifecycle-lock files contain no workspace data.

## Development reports

Never upload real personal evidence, private correspondence, credentials, vault files or identifying case records to GitHub issues, discussions, pull requests or test fixtures.
