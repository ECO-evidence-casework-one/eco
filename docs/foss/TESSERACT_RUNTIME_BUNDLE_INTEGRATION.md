# Verified Tesseract runtime bundle integration

Date: 2026-09-04

ECO already had a Tesseract adapter, OCR provenance model and executable-only local tool registry. Wave 4 produced a controlled GitHub-only Windows x64 Tesseract runtime, but that runtime also depends on sibling DLLs plus bundled `eng` and `osd` traineddata. Executable-only fingerprinting was therefore not sufficient to claim the complete OCR runtime remained the qualified build.

This integration recognises the exact Wave 4 control receipts (`BUILD_MANIFEST.json`, `RUNTIME_FILE_HASHES.json`, `OCR_SMOKE_RESULT.txt`), verifies all 151 runtime files and rejects missing, changed, extra, non-regular or symlinked entries. It checks the exact qualified Tesseract and tessdata source commits and requires the recorded OCR smoke result `ECO OCR TEST 123`. Registration and each later use re-run the complete inventory verification and the runtime's version/language-data probes.

Bundled language data is supplied to the Tesseract child process with `--tessdata-dir`; ECO does not mutate the process-global `TESSDATA_PREFIX`. Existing executable-only Tesseract registration remains a backwards-compatible fallback, but the verified Wave 4 bundle is preferred automatically when registered.

The Windows Settings page gains a `Locate verified Tesseract runtime` control. The bundle itself is not committed into the ECO source repository; it remains a separately acquired, hash-controlled GitHub Actions build artifact/source package.

## Actual Wave 4 artifact qualification

A separate Windows qualification run used the exact Wave 4 GitHub Actions artifact (`9939304279`) rather than a synthetic runtime. The downloaded artifact matched SHA-256 `d18fec356179df7ddb752d8ba67ef17021ec876cc8a887724b09791199085331`; its inner runtime ZIP matched SHA-256 `26d90693819d1b1ad0b2a35593a5555bb16d6d6f0f1482541ab7eacd531bb8b0`. ECO successfully registered and re-verified that real bundle, executed English OCR through the bundle-aware `--tessdata-dir` path, and then rejected the registered runtime after the test deliberately modified `runtime/bin/tesseract55.dll`. GitHub Actions run `33891211052` completed successfully.
