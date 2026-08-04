# V39 M1.18 test summary

**Local validation date:** 4 August 2026  
**Package:** `internal/runtime/turnorchestrator`  
**Base intended for branch:** canonical `main` `f997e1049f8c24ed04848127ec26d55ee784b6f4`

## Source inventory

| File | Bytes | SHA-256 |
|---|---:|---|
| `errors.go` | 183 | `f2c6fb9055c06f2393a6f82eae90145b3fff22b495dbd4cba2bd287192d805fe` |
| `orchestrator.go` | 14,100 | `ee7e0c0861e7cd2af5bd5ceb66e4228c5e2725b8898698653755b7826b91b1f1` |
| `types.go` | 6,678 | `3f216fc6d5b08e6f6766155a298cfbcc6f713bbdedb951b81f30246db351caf8` |
| `helpers_test.go` | 4,071 | `8bd93e22063cd468d241dd9abcb38813e5b2e79163a748ca730577e90abf4d50` |
| `orchestrator_failure_test.go` | 11,089 | `f76b8a86c29ae970d321fd4ee424ec0181d6fbcc0d85a16af98d2a62afaf0dc7` |
| `orchestrator_policy_test.go` | 2,612 | `a49e3f8b7e80eca0174c3b25cd25e819f2184dbf9ab61782b21bf3b15b9aed75` |
| `orchestrator_success_test.go` | 7,243 | `6e9c93136da82e76a26ec6143e21091ae7877452de856eb5ae944a84e313179d` |

The table records the pre-publication local source identity before the final documentation addition. Git/GitHub blob and commit identities remain controlling after publication.

## Validation completed

- `gofmt`: PASS;
- focused ordinary tests: PASS;
- focused race detector: PASS;
- focused race detector repeated 10 times: PASS;
- focused ordinary suite repeated 100 times: PASS;
- `go vet`: PASS;
- current repository source-policy rules applied to the package: PASS;
- statement coverage: **89.6%**;
- Windows amd64 test-package cross-compilation: PASS;
- temporary Windows test binary size: `4,484,608` bytes;
- temporary Windows test binary SHA-256: `9c67611ec109108d64d69df66002399cce1670be257f24513c8fd784b6c379b9`;
- temporary Windows test binary deleted after identity recording: YES.

The temporary test executable was not uploaded, distributed or treated as a user application.

## Tested behaviours

- accepted turn follows compile → route → admit → generate → release → verify → erase;
- accepted output is verifier-produced, not raw generated text;
- deterministic fallback for bounded dependency failures;
- invalid request blocks all downstream dependencies;
- context, route and generation identity validation;
- stream turn/run binding;
- strict chunk sequencing;
- per-chunk and total-output limits;
- cancellation and timeout without fallback text;
- late-chunk suppression;
- verifier rejection and verifier failure;
- dependency panic containment;
- lease release exactly once;
- lease-ID panic handling with cleanup;
- release failure suppresses output;
- erasure failure suppresses accepted and fallback output;
- all orchestrator-owned buffers reach the eraser and are cleared;
- deterministic fallback declaration enforcement;
- concurrent turn isolation;
- cleanup deadline propagation;
- request-timeout policy cap;
- receipt digest reconstruction;
- receipt contains no raw content or direct content fingerprints.

## Qualification boundary

This local validation proves only the self-contained package at the recorded source identity. It does not prove compatibility with the complete GitHub repository until exact-head GitHub Actions pass. It does not prove application integration, real worker enforcement, model quality, evidence accuracy, accessibility, packaging or release readiness.
