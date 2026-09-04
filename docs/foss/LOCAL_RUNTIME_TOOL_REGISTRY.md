# ECO verified local runtime tool registry

Date: 2026-09-04

## Mission fit

ECO already has source-safe local adapters for several mature FOSS engines, but those workflows historically required a raw executable path on each call. This integration closes that product-plumbing gap without introducing cloud discovery, automatic downloads, a plaintext tool database, or a workspace schema migration.

The user/operator approves a local executable once. ECO records its exact resolved path, SHA-256, byte size, version, upstream project and licence in the existing encrypted authenticated workspace change ledger. Registered workflows then reverify that exact executable before use.

## Approved runtime kinds

| ECO kind | Upstream | Licence | Existing ECO role |
|---|---|---|---|
| `tesseract` | `tesseract-ocr/tesseract` | Apache-2.0 | coordinate-bearing image OCR |
| `docling` | `docling-project/docling` | MIT | local structured document extraction |
| `ocrmypdf` | `ocrmypdf/OCRmyPDF` | MPL-2.0 | sidecar-only scanned/mixed PDF OCR |
| `llama.cpp` | `ggml-org/llama.cpp` | MIT | local GGUF reasoning runtime |
| `pdfcpu` | `pdfcpu/pdfcpu` | Apache-2.0 | offline read-only PDF structure inspection |

Aliases such as `tesseract-ocr`, `llamacpp` and `llama-cli` are normalized to the canonical ECO kind before persistence.

## Registration boundary

`RegisterLocalTool` / `RegisterLocalToolContext`:

1. accepts only a known ECO runtime kind;
2. requires an absolute path resolving to a regular local file;
3. records the file identity before inspection;
4. calculates SHA-256;
5. invokes the engine's already-existing ECO version probe;
6. calculates SHA-256 again and re-stats the file;
7. rejects mutation during registration;
8. persists only a verified registration in the encrypted hash-chained change ledger.

The generic audit map stores size and timestamp as strings to avoid JSON number precision/type ambiguity after encrypted workspace save/reopen cycles.

## Runtime verification boundary

`VerifyRegisteredLocalTool` / `VerifyRegisteredLocalToolContext` fails closed unless the newest registration for that kind still resolves to:

- the same exact path;
- the same byte size;
- the same SHA-256;
- the same engine version;
- the same approved upstream/licence identity.

The file is hashed before and after the version probe and its stable file identity is checked again. If it has been replaced, edited, upgraded, moved or removed, ECO requires re-registration rather than silently trusting the changed binary.

## Registered workflow entry points

The integration adds ordinary calls that no longer require callers to repeatedly supply executable paths:

- `OCRImageWithRegisteredTesseract`
- `ExtractEvidenceWithRegisteredDocling`
- `ExtractEvidenceWithRegisteredOCRmyPDF`
- `InspectEvidencePDFWithRegisteredPDFCPU`
- `AskWithRegisteredLlamaCPP`

Each has a cancellable `...Context` form. These wrappers first verify the registered runtime and then call the already-qualified source-safe adapter/workflow.

Docling's local model/artifact directory remains an explicit local input because it is separate from the Docling executable. llama.cpp's GGUF path remains explicit because the existing adapter independently identifies and hashes the selected model on every run.

## Query surface

- `RegisteredLocalTool(kind)` returns the newest authenticated registration for one kind.
- `RegisteredLocalTools()` returns all active known registrations in stable canonical order.

These APIs are intended for Settings/diagnostics and later UI wiring; callers do not need to understand the encrypted change-ledger representation.

## Persistence and portability

Registrations are stored inside ECO's encrypted workspace manifest and therefore travel inside ECO's existing encrypted portable backup. No separate plaintext registry is introduced.

A restored registration is historical configuration, not proof that the same path exists on the restored computer. The first registered workflow use re-verifies the local path/hash/version and fails closed if the executable is absent or different.

## What this does not do

- It does not download, install or update any runtime.
- It does not approve arbitrary executables or arbitrary upstream projects.
- It does not bundle external engines into `ECO.exe`.
- It does not replace the binary SBOM; external runtimes remain separate from components compiled into ECO itself.
- It does not by itself settle future redistribution/licence-notice obligations for bundled installers.
- It does not treat a registered binary as permanently trusted: every registered workflow re-verifies it.

## Product effect

This turns the existing FOSS adapters from low-level path-based plumbing into a controlled local-runtime layer: approve once, see what ECO is configured to use, and fail closed when that local software changes.
