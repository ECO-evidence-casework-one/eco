# OCRmyPDF sidecar-only adaptation

Date: 2026-09-04
Upstream: `ocrmypdf/OCRmyPDF`
Pinned commit: `196072acef056aa11b63e60564d98762bd38f638`
Upstream version at pin: 17.11.0
Licence: MPL-2.0
ECO integration: local external process adapter; no OCRmyPDF source code copied into ECO

## Why OCRmyPDF is useful after Tesseract and Docling

ECO already has a direct local Tesseract image adapter and a local Docling document-reading adapter. OCRmyPDF fills a narrower remaining gap: orchestrating page rasterisation and OCR across scanned or mixed-content PDFs while preserving page order.

This integration does not replace those readers. OCRmyPDF's own sidecar documentation states that the sidecar contains OCR text produced on pages where OCR was performed; text already present in a PDF is not copied into the sidecar. Therefore treating the sidecar as the complete document text would silently omit material.

ECO stores OCRmyPDF output only as supplemental page-bound source segments.

## Pinned upstream contract

The pinned OCRmyPDF CLI supports the controls used by this slice:

- `--output-type none` to produce no output PDF when only a sidecar is wanted;
- `--sidecar FILE` for Tesseract-recognised text;
- `--mode skip` to leave pages that already contain text out of OCR;
- `--ocr-engine tesseract`;
- `--rasterizer pypdfium`;
- `--jobs 1`;
- `--tesseract-timeout`;
- `--max-ocr-image-mpixels`;
- `--language`.

The pinned package requires Python 3.11 or newer and lists pypdfium2 as a core dependency.

OCRmyPDF's sidecar merger writes a form-feed between page positions. ECO uses those separators to preserve page numbers even when an individual page contributes no OCR text.

## ECO execution profile

The first qualified slice uses:

- output type: `none`;
- sidecar: temporary file inside a private temporary directory;
- OCR mode: `skip`;
- OCR engine: `tesseract`;
- rasterizer: `pypdfium`;
- jobs: `1`;
- Tesseract timeout: 30 seconds per OCR operation;
- maximum OCR image size: 25 megapixels;
- language: explicit caller-selected installed Tesseract language, initially expected to be `eng` for the packaged configuration;
- progress bar: disabled.

ECO does not pass plugin, force-OCR, redo-OCR, rotate, deskew, cleaning, background-removal or PDF optimisation/conversion options in this slice.

Using one worker plus a bounded OCR image size is deliberately conservative for the low-spec Windows target. It is a resource guard, not a guarantee of total process memory consumption.

## Source and output boundary

OCRmyPDF never receives the user's original source path or ECO's encrypted `.ecoobj` file.

ECO:

1. locates a committed verified PDF preservation record;
2. decrypts the preserved object into ECO's private `derived/.work` verified reading copy;
3. verifies that copy against the preserved SHA-256;
4. makes the reading copy read-only;
5. runs OCRmyPDF against that temporary copy;
6. requires sidecar-only output;
7. rejects the run if an output PDF unexpectedly appears;
8. bounds and reads the UTF-8 sidecar;
9. re-hashes the reading copy to prove OCRmyPDF did not mutate it;
10. re-verifies the encrypted preserved object immediately before committing derived segments;
11. deletes temporary reading/sidecar/work files.

No derived searchable PDF is retained by this integration.

## Segment semantics

The sidecar is split on OCRmyPDF's page form-feeds before ECO's ordinary text normalisation.

Every accepted segment receives:

- a distinct `SEG-OCRPDF-*` ID;
- a global ordinal within that OCRmyPDF result;
- the exact page position from the sidecar;
- origin `ocrmypdf`;
- the preserved object filename;
- the preserved SHA-256.

Sidecar text does not provide ECO with trustworthy word boxes or confidence scores. This integration therefore sets no region and no confidence. A runner that attempts to inject coordinates or confidence into this sidecar path is rejected.

OCR text is a recognition suggestion, not verified transcription or content truth. User-facing/reasoning layers must retain that distinction.

## Idempotence and coexistence

A rerun removes only older segments whose origin is `ocrmypdf`, then inserts the new OCRmyPDF result.

It does not replace:

- `EvidenceItem.ExtractedText`;
- native extraction segments;
- Docling segments;
- direct coordinate-bearing Tesseract OCR segments;
- stronger existing status/readability when one already exists.

If a PDF had no readable text before and OCRmyPDF returns useful sidecar segments, ECO may mark the evidence readable and identify the status as OCRmyPDF OCR suggestions.

## Resource guard

The adapter uses ECO's generic local resource-pressure preflight before launching OCRmyPDF. Critical RAM/working-disk pressure may block the engine; high CPU remains advisory-only under the shared policy.

The resulting resource assessment is carried into the authenticated OCRmyPDF audit event.

## Offline/process boundary

ECO invokes a caller-selected absolute local executable directly with `exec.CommandContext`; no shell is used.

Common proxy environment variables are removed and `NO_PROXY=*` is set for the child process. No network or cloud option is used by this adapter.

This is a runtime policy boundary, not a claim that every arbitrary third-party Python environment is intrinsically incapable of networking. ECO's packaged runtime must still pin and audit the actual OCRmyPDF/Python dependency set before public release.

## Licence boundary

OCRmyPDF is MPL-2.0. This slice communicates with OCRmyPDF as a separate local executable and copies no OCRmyPDF source into ECO.

If ECO later redistributes OCRmyPDF or a Python environment containing it, release packaging must preserve the MPL notices and source-code obligations for covered OCRmyPDF files, plus the licences/notices for its dependencies. That packaging/licence work is separate from this source adapter and must be completed before redistribution.

## Tests

The integration tests cover:

- sidecar-only/bounded CLI arguments;
- absence of force/redo/preprocessing/plugin options;
- proxy-environment isolation;
- page-position preservation across blank/skipped page slots;
- distinct page-aware segment IDs;
- source SHA/object binding;
- no invented geometry or confidence;
- bounded diagnostic capture;
- verified ECO reading-copy use;
- preservation of existing native/Docling text;
- rerun replacement of only old OCRmyPDF segments;
- non-PDF rejection before runner invocation;
- forged source-binding rejection;
- authenticated audit claims that no output PDF is retained and existing reading is not replaced.

## Non-claims

This slice does not claim that:

- OCRmyPDF sidecar text is complete;
- OCR text is correct;
- a page with no sidecar text is blank;
- a 30-second timeout is appropriate for every document;
- 25 megapixels or one worker guarantees a specific memory ceiling;
- OCRmyPDF validates legal/evidential meaning;
- the packaged Windows runtime is complete yet.

Actual OCRmyPDF/Tesseract/pypdfium installation and packaged-runtime qualification remain a later Windows packaging gate. No user installation is required for repository-level CI in this slice.
