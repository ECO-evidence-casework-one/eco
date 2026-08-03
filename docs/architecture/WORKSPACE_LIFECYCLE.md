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

Inspection and opening run as one object-bound lifecycle transaction. ECO holds a cross-process lock for the logical workspace path and retains the exact root, key, encrypted database, routing identity and objects directory while it authenticates them and completes the first verified lifecycle write. Later metadata writes reacquire the same lifecycle lock, re-authenticate the immutable workspace identity and replace metadata relative to the retained root. A replacement root, key, database, identity or objects directory is rejected without writing either the authentic or substituted workspace.

Workspace format 1 has one approved migration path to format 2. Migration uses an HMAC-authenticated recovery record bound to the canonical workspace root, workspace ID, exact candidate/build transition, random nonce, source/destination schemas, phase and random checkpoint/stage identities. Every participant must be a normal direct sibling; symbolic links, junctions and reparse points are rejected before deletion or rename. Migration renames the untouched original to a checkpoint, copies it to a separate staging folder, freshly verifies committed and pending preserved objects, recovers verified pending preservation through the issue #3 state machine, and activates only verified state. A successful checkpoint is retained. An interrupted or unsuccessful migration either verifies the activated workspace or restores the original checkpoint; compensating rollback restores the active path if checkpoint activation fails.

Portable restore uses a separate HMAC-authenticated recovery record bound to the canonical root, original and restored workspace identities, build/candidate/schema, backup SHA-256, random nonce, exact checkpoint/stage/failed roles and the `prepared`, `staged`, `original-moved`, `activated` and `recovered` phases. The active root and stage are renamed only through retained object handles/descriptors with no-replace semantics. Staging cleanup walks and removes only the authenticated retained tree. Startup can verify an activated restore or restore the original checkpoint after interruption at either rename boundary.

On Windows these controls use `CreateFileW` with reparse-point opening and delete/rename sharing denied, handle-derived volume/file IDs, `SetFileInformationByHandle` and handle-relative `NtSetInformationFile`. On Linux/amd64 they use `openat` with `O_NOFOLLOW`, descriptor-derived inode/device/mount IDs, `flock`, `renameat2(RENAME_NOREPLACE)` and `unlinkat`. Platforms without equivalent object-bound primitives fail closed for workspace opening and destructive lifecycle changes.

Reset requires confirmation of the selected workspace name, identity and path. It first commits empty encrypted metadata, then retires only encrypted object names referenced by that workspace. It does not recursively delete the workspace folder and does not remove unrelated files, source evidence outside the workspace, another workspace or an arbitrary user folder. The reset audit retains the previous audit-head hash without retaining old records or conversations in the visible workspace.

## Current limits

Only the format 1 to format 2 migration is approved. Other older formats and all downgrade attempts are blocked. Successful migration and portable-restore checkpoints, plus failed copies retained during rollback, are intentionally kept and there is not yet an in-application cleanup control. During an unfinished migration or portable restore, the authenticated plaintext recovery record necessarily exposes canonical root/checkpoint/stage paths, opaque identities, nonce, phase and start time; migration also exposes its schema/build transition and restore exposes build/candidate/schema plus the encrypted-backup SHA-256. These records contain no workspace name, evidence, conversation, matter or setting content. Empty sibling lifecycle-lock files expose no path or workspace content beyond their filesystem location.
