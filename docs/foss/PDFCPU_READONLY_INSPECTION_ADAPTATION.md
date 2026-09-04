# pdfcpu read-only PDF inspection for ECO

Date: 2026-09-04

## Upstream source reviewed

- Project: `pdfcpu/pdfcpu`
- Pinned source commit: `105c0c28727afe7f85eb3179d03e9810d8774981`
- Upstream licence: Apache-2.0
- Reviewed contracts:
  - `cmd/pdfcpu/root.go`
  - `cmd/pdfcpu/document.go`
  - `pkg/cli/document_exec.go`
  - `pkg/pdfcpu/info.go`
  - `go.mod`

The Wave-1 donor acquisition audit separately records the exact upstream commit and licence-file hash.

## Why ECO uses the CLI boundary instead of importing the Go library

The pinned pdfcpu source declares Go `1.25.0`. ECO currently declares Go `1.23` and its CI/build baseline is qualified against that line. Importing the current pdfcpu library directly would therefore force a project-wide toolchain change merely to gain PDF inspection.

This slice avoids that unnecessary coupling. pdfcpu remains a separately built local FOSS engine invoked through a narrow read-only process adapter. ECO's own Go dependency graph remains unchanged.

## Exact pinned pdfcpu controls used

The pinned CLI exposes:

- global `--offline` / `-o`: disable HTTP traffic;
- global `--conf`: select or disable the configuration directory;
- `validate --mode relaxed|strict inFile`;
- `info --json inFile`;
- `version`.

ECO invokes every pdfcpu command with `--offline --conf disable`. It never passes modifying commands such as optimize, merge, split, rotate, watermark, stamp, encrypt, decrypt or create.

## ECO inspection boundary

`internal/eco/pdfcpu_adapter.go` implements `RunPDFCPUInspection(...)`.

It requires:

1. an explicit absolute local pdfcpu executable;
2. an explicit absolute verified reading-copy path supplied by ECO;
3. a valid app-owned `SourceReceipt` identifying the preserved encrypted object and its SHA-256.

The adapter then:

1. identifies the local pdfcpu version;
2. performs relaxed validation;
3. only if relaxed validation passes, performs strict validation;
4. only after relaxed validation passes, requests `info --json`;
5. parses one bounded JSON info record into `PDFAssessment`;
6. caps stdout and diagnostics;
7. removes common proxy environment variables in addition to pdfcpu's own `--offline` control;
8. uses no shell, no URL, no stdin and no output PDF path.

`internal/eco/pdfcpu_workflow.go` connects this to the preserved-evidence boundary using `withVerifiedPreservedFile`. The pdfcpu process therefore never receives the user's original source path or ECO's encrypted `.ecoobj`; it receives only a fresh, read-only temporary materialisation whose bytes are verified against the preserved SHA-256 before and after the callback.

A compact authenticated audit event records the engine version, source hash, validation outcomes, PDF version, page count, encryption/signature/form flags and attachment count. The detailed assessment is returned to the caller; this slice does not create a plaintext derivative database or modified PDF.

## Validation wording is deliberately restrained

`relaxed_validation_passed=false` is not automatically translated into “corrupt PDF”. Validation can fail for more than one reason, including inputs the first slice cannot inspect without credentials or features it does not support. ECO preserves the bounded diagnostic and states that the original remains authoritative.

Likewise, passing structural validation says nothing about whether the document's factual contents are true, complete, authentic, legally significant or untampered before ECO received it. The audit event explicitly records `inspection_is_content_truth=false`.

## Information captured after relaxed validation passes

The pinned pdfcpu JSON schema exposes, among other fields:

- PDF version;
- page count;
- title, author, subject, producer and creator metadata;
- creation/modification dates;
- tagged/hybrid/linearized flags;
- cross-reference/object-stream flags;
- watermark, thumbnail, form, signature, append-only, bookmark and names flags;
- encrypted/permissions state;
- attachments.

ECO bounds metadata strings before exposing them in `PDFAssessment`.

## Deliberate first-slice limits

- No owner/user password parameters are accepted yet.
- No broken-link validation is requested, because that would conflict with the offline-only runtime rule.
- No repair/optimization/write operation is performed.
- No pdfcpu binary is bundled by this source-only PR; packaging remains a later controlled step.
