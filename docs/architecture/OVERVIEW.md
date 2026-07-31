# Native architecture overview

The current Version 25 native branch is a Go Windows GUI application with an internal evidence core.

Current responsibilities include:

- Windows desktop interface;
- encrypted local workspace;
- streaming file import and hashing;
- type-signature detection and quarantine;
- evidence extraction where supported;
- image preview and assessment;
- deterministic local source-backed retrieval;
- backup, restore and integrity controls.

Planned architecture separates high-risk image, OCR, document and AI processing into restricted local workers. The final application must not depend on a browser, localhost, Python installation, cloud service or online API.
