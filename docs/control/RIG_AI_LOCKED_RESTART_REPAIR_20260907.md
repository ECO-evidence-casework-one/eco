# Rig AI locked-restart repair — 7 September 2026

## Observed real Windows failure

A repeated rig setup stopped before rebuilding because the previous failed preview directory could not be renamed. Windows reported that a captured `native-*.stderr.txt` file under the disposable work tree was still in use by another process.

The failed restart happened in the old v3 recovery preamble while moving the whole output root. This is a preparer/restart-control defect, not evidence that Qwen itself failed.

## Repair

The qualified v3 build/model recipe remains unchanged.

A new restart supervisor now:

1. detects completed versus failed/interrupted output;
2. attempts the normal whole-directory archive;
3. if Windows refuses that archive because any old work file is locked, leaves the failed directory untouched;
4. chooses a unique sibling retry directory;
5. moves only the existing Qwen `.part` file, with bounded retry, into a small failed-state seed recognised by v3;
6. invokes the unchanged v3 preparer, which performs its existing controlled partial-file recovery and normal build/model verification;
7. records the actual output path in a sidecar pointer so the one-click CMD reports the correct result directory.

The supervisor does not kill processes, weaken Windows security, delete the locked failed directory, trust a partial model as complete, skip the final Qwen SHA-256 gate, or change the frozen ECO/Qwen executable/model/runtime identities.

## Regression test

The dedicated Windows workflow deliberately opens a synthetic `native-synthetic.stderr.txt` with `FileShare.None`, creates a synthetic failed setup with an 8-byte Qwen `.part`, and requires:

- the locked failed root remains in place;
- a fresh sibling retry root is selected;
- the Qwen partial is moved intact to the retry seed;
- the seed remains recognisable by the qualified v3 preparer;
- the existing v3 native-capture self-test still passes under Windows PowerShell 5.1;
- the one-click launcher contains the supervisor and actual-output-pointer controls.

No executable, model or runtime artifact is uploaded by this workflow.
