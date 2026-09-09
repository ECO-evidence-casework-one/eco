# AEIB v0.1 Case Catalogue

| ID family | Purpose |
|---|---|
| `AEIB-TXT-*` | Encoding/searchability boundaries |
| `AEIB-SIG-*` | Extension/signature mismatch |
| `AEIB-ZIP-*` | Duplicate/nested/corrupt ZIP and path identity |
| `AEIB-TAR-*` | Link metadata and traversal safety |
| `AEIB-EML-*` | Duplicate attachments, nested mail/archive, malformed MIME/encoding |
| `AEIB-OFFICE-*` | Malformed OOXML-like container handling |
| `AEIB-BIN-*` | Unknown deterministic binary input |
| `AEIB-RES-*` | Bounded high-compression/resource behavior |
| `AEIB-MUT-*` | Source mutation/change-detection fixtures |

The generated `manifest.json` is controlling for exact paths, sizes, SHA-256 values, expected behavior and notes for a given seed.
