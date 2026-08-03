# Development workspace lifecycle

ECO development candidates must not treat old local data as though it belongs to a new build. Workspace selection is therefore explicit, candidate-specific and recoverable.

## Identity and compatibility

Every current workspace has:

- a random workspace identity;
- a layman-facing name;
- an exact folder path;
- the build that created it;
- an encrypted workspace format number; and
- a small `workspace.identity.json` file that contains identity and compatibility information, but no evidence, conversation or matter content.

The encrypted record and identity file must agree. A mismatch blocks opening. A build may reopen a workspace only when it understands that format. A newer format is blocked from an older build, and an older unsupported format is left unchanged.

## Candidate-specific application state

Each development candidate derives a separate application-state folder from its build identity, embedded source revision and exact executable SHA-256. Its automatic development workspace is inside that folder. A different committed or ad-hoc binary produces a different folder even when the public milestone label has not changed, so another candidate's last-selected workspace, settings and test records are not inherited. If ECO cannot fingerprint its own executable, startup is blocked instead of risking an ambiguous candidate identity.

The candidate application audit records successful and blocked create, reopen, migration, recovery and reset actions. A previously selected external workspace is recorded for truthfulness but is never automatically reopened at the next launch. Restart opens only that candidate's own default development workspace; an external workspace must be selected again deliberately.

## User-visible flows

First launch creates a new, empty candidate workspace. Creating another workspace requires a new or empty folder. Reopening requires deliberate folder selection and a compatible identity. The interface always displays the workspace name, identity, path, open status and compatibility wording.

Workspace format 1 has one approved migration path to format 2. Migration renames the untouched original to a checkpoint, copies it to a separate staging folder, updates and verifies the staged encrypted state, and activates it only after every preserved evidence object passes integrity checks. A successful checkpoint is retained. An interrupted or unsuccessful migration either verifies the activated workspace or restores the original checkpoint; it never presents unverified old state as migrated data.

Reset requires confirmation of the selected workspace name, identity and path. It first commits empty encrypted metadata, then retires only encrypted object names referenced by that workspace. It does not recursively delete the workspace folder and does not remove unrelated files, source evidence outside the workspace, another workspace or an arbitrary user folder. The reset audit retains the previous audit-head hash without retaining old records or conversations in the visible workspace.

## Current limits

Only the format 1 to format 2 migration is approved. Other older formats and all downgrade attempts are blocked. Successful migration checkpoints are intentionally retained and there is not yet an in-application checkpoint-cleanup control.
