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
- Only the recoverable workspace-format 1 to format 2 migration is currently approved; other older formats and downgrade attempts are blocked.
- Successful workspace-migration checkpoints are retained for rollback and do not yet have an in-application cleanup control.
- A minimal plaintext workspace routing identity exposes format, opaque workspace ID, development kind, schema and exact candidate identity. Candidate app-state exposes candidate/build identity plus opaque, hash-chained action audit fields, but no workspace names or full paths.
- Development workspaces from before exact candidate binding have no trustworthy candidate identity and are blocked instead of being attributed by guesswork.
- An unfinished migration temporarily retains an authenticated plaintext recovery record containing canonical migration paths, opaque identities, build/schema transition, nonce, phase and start time. It contains no evidence, conversation, matter, workspace name or setting content.
- Failed migrated copies retained during compensating rollback do not yet have an in-application cleanup control.
- Successful portable restores retain the original workspace checkpoint for rollback; restore checkpoints and failed restore copies do not yet have an in-application cleanup control.
- An unfinished portable restore temporarily retains an authenticated plaintext recovery record containing canonical restore paths, opaque identities, build/candidate/schema, encrypted-backup SHA-256, nonce, phase and start time. It contains no casework content or workspace name.
- Object-bound workspace opening, reset, migration and portable-restore mutation are implemented for Windows and Linux/amd64. Other platforms fail closed for these operations until equivalent primitives are implemented.
- Windows application version-resource metadata and the final installer/uninstaller are not yet complete.

## Safety position

No current source milestone should be described as an approved stable release. Do not disable Smart App Control or other Windows protections to run an unsigned ECO artifact.
