# Version 25 N2 P1 implementation record

**Build ID:** `ECO-V25-20260731-N2-P1`  
**Name:** Native Document Vision Foundation Preview 1  
**Release class:** source development snapshot

## Implemented

- document-boundary suggestion with confidence and conservative full-image fallback;
- non-destructive crop view;
- bounded skew-correction estimation;
- arbitrary-angle non-destructive rotation;
- adaptive threshold reading view;
- glare, shadow imbalance, edge-cutoff and probable double-page indicators;
- bounded preview image generation for large photographs;
- tested quadrilateral perspective-correction foundation;
- OCR TSV parser retaining word/line text, confidence, page and normalised coordinates;
- OCR receipt validation;
- bounded OCR identities, text collections, confidence values and nested word structures;
- exact receipt-to-index segment correspondence validation;
- transactional in-memory rollback if the encrypted workspace update cannot be persisted;
- source SHA-256 match gate before OCR can enter the encrypted workspace;
- removal/replacement of prior OCR-derived segments without changing original evidence;
- low-confidence OCR exclusion from Ask ECO retrieval;
- OCR-aware citation labels and support wording;
- exact OCR-region highlight model in the native source preview;
- new core and static native-UI regression tests.

## Not implemented

- production OCR engine invocation;
- OCR model/language packaging;
- interactive crop handles or four-corner perspective editor;
- persistence of derived image renditions;
- PDF page rasterisation;
- generative local model.
