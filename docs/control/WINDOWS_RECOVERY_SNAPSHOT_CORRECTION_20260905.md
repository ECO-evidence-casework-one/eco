# Windows recovery snapshot measurement correction

5 September 2026. Refs #4 and PR #132. This is test-measurement evidence, not an independent review, crash-recovery certificate or public-release approval.

## Observed failure and control

PR #132 diagnostic head `42db89c6b83c494ee3b6707497dabc7ee575cff6` failed native Windows run `33973606761`, job `101326418805`. The diagnostic compared two snapshots with no restore between them, then compared original and untouched staged workspaces after a boundary-1 rollback.

The no-restore control reported changing modification times for the `derived` and `derived/.work` directories. That diagnostic reported no changed file hashes, sizes, modes or directory entries. It established a measurement discrepancy independent of rollback, not proof that all recovery failures were harmless.

## Source check and correction

Microsoft documents that NTFS attributes returned by directory enumeration can be stale and recommends querying an open handle for current attributes. In the exact Go 1.23.12 source, `(*os.File).Stat` uses `statHandle`, which obtains file information by handle; pathname `os.Stat`/`os.Lstat` can first use `GetFileAttributesEx`. Merely replacing `DirEntry.Info` with pathname `os.Stat` is therefore not the same correction.

The snapshot helper now obtains metadata for regular files and directories through `os.Open` and `File.Stat`, with checked stat and close errors. Links retain non-following entry metadata, and special files are not opened. This helper is test-only; it does not modify production recovery logic or remove acceptance checks.

References checked on 5 September 2026:

- https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-findfirstfilew
- https://github.com/golang/go/blob/go1.23.12/src/os/stat_windows.go

## Checks retained and added

All existing recovery, owner-identity, save/reopen and compatibility assertions remain. The snapshot still compares modification timestamps, modes, sizes, entry presence and regular-file SHA-256 hashes. No directory timestamp field is discarded, no sleep is added, and no Windows test is skipped.

Seven additional measurement controls are included:

1. Repeated read-only snapshots remain identical (ten repetitions for both original and staged synthetic workspaces).
2. Same-length changed bytes remain detectable after restoring the original file modification time.
3. A changed file modification time is detected.
4. A changed directory modification time is detected.
5. A changed file mode is detected.
6. An added entry is detected.
7. A removed entry is detected.

The existing Windows diagnostic remains as a regression test. Passing requires actual equality; it was not changed to tolerate the earlier discrepancy. Final PR/head and post-merge qualification must be read from their actual Actions results, not inferred from this method note.

## Finish-line scope stays fixed

`PRIVATE_TEST_CANDIDATE_FINISH_LINE_20260905.md` retains the same eight A-checks and eight W-packages. W7 includes both handled-error rollback and process-interruption/restart recovery; completing PR #132 alone does not close W7 or #4. W8 remains actual Windows/Acer acceptance. The separate #131 native search UI is already merged; the finish-line baseline predates that merge, and its A5 practical/accessibility acceptance is not claimed by this correction.

No new parser, model, runtime, UI framework, application dependency, binary distribution or new finish-line prerequisite is introduced.
