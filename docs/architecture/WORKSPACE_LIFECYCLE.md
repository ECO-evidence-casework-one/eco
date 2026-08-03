# Development workspace lifecycle

ECO development candidates must not treat old local data as though it belongs to a new build. Workspace selection is therefore explicit, candidate-specific and recoverable.

## Identity and compatibility

Every current workspace has:

- a random workspace identity;
- a layman-facing name;
- an exact folder path;
- the build and exact candidate that created it;
- an encrypted workspace format number; and
- a small `workspace.identity.json` routing file containing only its format, opaque workspace ID, development kind, schema and exact candidate identity.

The workspace name, creation time and creating build remain only in the authenticated encrypted record. The encrypted record and routing identity must agree exactly about workspace and candidate identity; absence, damage or mismatch blocks opening. A build may reopen a workspace only when it understands that format. A newer format is blocked from an older build, and an older unsupported format is left unchanged.

## Candidate-specific application state

Each development candidate derives a separate application-state folder from its build identity, embedded source revision and exact executable SHA-256. Its automatic development workspace is inside that folder. A different committed or ad-hoc binary produces a different folder even when the public milestone label has not changed, so another candidate's last-selected workspace, settings and test records are not inherited. If ECO cannot fingerprint its own executable, startup is blocked instead of risking an ambiguous candidate identity.

The plaintext candidate app-state contains the candidate/build identity, opaque workspace IDs, action names, outcomes, timestamps, lay summaries and a hash-chained audit. It does not persist workspace names or full workspace paths. A previously selected external workspace is recorded only by opaque identity and is never automatically reopened at the next launch. Restart opens only a default workspace whose encrypted and routing identities both match the exact candidate; an external workspace must be selected again deliberately.

## User-visible flows

First launch creates a new, empty candidate workspace. Creating another workspace requires a new or empty folder. Reopening requires deliberate folder selection and a compatible identity. The interface always displays the workspace name, identity, path, open status and compatibility wording.

Workspace format 1 has one approved migration path to format 2. Migration uses an HMAC-authenticated recovery record bound to the canonical workspace root, workspace ID, exact candidate/build transition, random nonce, source/destination schemas, phase and random checkpoint/stage identities. Every participant must be a normal direct sibling; symbolic links, junctions and reparse points are rejected before deletion or rename. Migration renames the untouched original to a checkpoint, copies it to a separate staging folder, freshly verifies committed and pending preserved objects, recovers verified pending preservation through the issue #3 state machine, and activates only verified state. A successful checkpoint is retained. An interrupted or unsuccessful migration either verifies the activated workspace or restores the original checkpoint; compensating rollback restores the active path if checkpoint activation fails.

Reset requires confirmation of the selected workspace name, identity and path. It first commits empty encrypted metadata, then retires only encrypted object names referenced by that workspace. It does not recursively delete the workspace folder and does not remove unrelated files, source evidence outside the workspace, another workspace or an arbitrary user folder. The reset audit retains the previous audit-head hash without retaining old records or conversations in the visible workspace.

## Current limits

Only the format 1 to format 2 migration is approved. Other older formats and all downgrade attempts are blocked. Successful migration checkpoints and failed migrated copies retained during rollback are intentionally kept and there is not yet an in-application cleanup control. During an unfinished migration, the authenticated plaintext recovery record necessarily exposes the canonical workspace/checkpoint/stage paths, schema/build transition, opaque identities, nonce, phase and start time; it contains no workspace name, evidence, conversation, matter or setting content.
