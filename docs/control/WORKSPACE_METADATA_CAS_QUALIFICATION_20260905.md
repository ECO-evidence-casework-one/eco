# Workspace metadata CAS qualification — 5 September 2026

Pre-PR controlled integration run `33961733775` completed successfully before this record was added.

The run applied the bounded CAS patch and required all of the following before committing product changes:

- source/offline policy PASS;
- `go test ./...` PASS;
- `go vet ./...` PASS;
- Windows/amd64 `internal/eco` test-binary cross-compile PASS;
- Windows/amd64 `cmd/eco` build PASS;
- `git diff --check` PASS.

The temporary patch script and apply workflow were then removed. Normal pull-request CI remains required before merge.
