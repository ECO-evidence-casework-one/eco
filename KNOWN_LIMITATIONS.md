# Known limitations

## Current N2 P1 source milestone

- No trusted code-signed end-user binary is available.
- A bundled OCR engine is not yet included.
- The coordinate-bearing OCR receipt parser and vault gate accept validated worker output, but no production OCR worker is enabled.
- Perspective correction is implemented as a tested transformation foundation; an interactive four-corner editor is not yet present.
- Auto-crop and deskew are conservative reading-view suggestions, not changes to original evidence.
- Crop and deskew suggestions can be wrong and require visual review.
- Handwriting recognition is not implemented.
- Native PDF page rendering is not yet implemented.
- A generative local language model is not included.
- Ask ECO remains a deterministic source-backed retrieval engine.
- Direct Windows Narrator, NVDA, high-DPI and long-duration execution evidence remains incomplete.
- The current encrypted workspace format remains an early development format and must not hold irreplaceable evidence.
- Windows application version-resource metadata and the final installer/uninstaller are not yet complete.

## Safety position

No current source milestone should be described as an approved stable release. Do not disable Smart App Control or other Windows protections to run an unsigned ECO artifact.
