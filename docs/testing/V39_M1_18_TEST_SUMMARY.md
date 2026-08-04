# V39 M1.18 test summary

**Local validation date:** 4 August 2026  
**Package:** `internal/runtime/turnorchestrator`  
**Current-main branch base:** `8cf4da3c1a1bc196280080d147ea5ab833111866`

The package was first published on a branch from `f997e1049f8c24ed04848127ec26d55ee784b6f4`. While that work was underway, `main` advanced through documentation/control-only PRs #55, #57 and #60. The exact hardened M1.18 source/test blobs were reconciled onto the current-main base above without selecting older application, workflow or release-gate files.

## Source inventory

| File | Bytes | SHA-256 |
|---|---:|---|
| `errors.go` | 183 | `f2c6fb9055c06f2393a6f82eae90145b3fff22b495dbd4cba2bd287192d805fe` |
| `orchestrator.go` | 15,926 | `c9854eafbdd45c536ec9ff2a25f9e7e3e724e99cf00c8699c59c01340fd0df5b` |
| `types.go` | 6,839 | `15068464b285e11db63d01827b25b1684876c17c9d61acd7ec59b044a59fd4d4` |
| `helpers_test.go` | 4,087 | `65a58a0a6a7d441a32af791c68d1c5ffb26b8177309b42b02b975225e76348d7` |
| `orchestrator_failure_test.go` | 17,707 | `07cc272f1780159e921fb80a087924f227cb349295c69560f7a31b8f45759fee` |
| `orchestrator_policy_test.go` | 3,769 | `2a70a490017a4efc802c45f076c001c98776dfc356203e72037837020082e97c` |
| `orchestrator_success_test.go` | 7,808 | `b91911f7f7876dd58c7b195653df6cf57035b1446742d08186ed26264e412c4e` |

The table records the post-red-team local source identity. Git/GitHub blob and commit identities remain controlling after publication.

## Validation completed

- `gofmt`: PASS;
- focused ordinary tests: PASS;
- focused race detector: PASS;
- focused race detector repeated 10 times: PASS;
- focused ordinary suite repeated 100 times: PASS;
- `go vet`: PASS;
- current repository source-policy rules applied to the package: PASS;
- statement coverage: **91.4%**;
- Windows amd64 test-package cross-compilation: PASS;
- temporary Windows test binary size: `4,535,808` bytes;
- temporary Windows test binary SHA-256: `e788826d4d002a2cee5b6ae40712a1137debcd28bcb535f54eb94dbb48a59793`;
- temporary Windows test binary deleted after identity recording: YES.

The temporary test executable was not uploaded, distributed or treated as a user application.

## Tested behaviours

- accepted turn follows compile → route → admit → generate → release → verify → erase;
- accepted output is verifier-produced, not raw generated text;
- deterministic fallback for bounded dependency failures;
- invalid request blocks all downstream dependencies;
- context, route and generation identity validation;
- cryptographically random per-run identities and repeated-input uniqueness;
- stream turn/run binding;
- strict chunk sequencing;
- per-chunk and total-output limits;
- cancellation and timeout without generated, verified or fallback text;
- cancellation detected during fallback, verification and final erasure suppresses output;
- fallback buffers returned alongside errors are zeroed;
- late-chunk suppression;
- verifier rejection and verifier failure;
- dependency panic containment, including clock failures;
- nil parent-context handling;
- lease release exactly once;
- lease-ID panic handling with cleanup;
- release failure suppresses output;
- erasure failure suppresses accepted and fallback output;
- partial and invalid compiler prompts reach the eraser;
- all orchestrator-owned prompt, runtime, generated, verifier-output and verification-copy buffers reach the eraser and are cleared;
- fallback byte buffers are zeroed after the returned string copy is created;
- deterministic fallback declaration enforcement;
- concurrent turn isolation;
- cleanup deadline propagation;
- request-timeout policy cap;
- receipt digest reconstruction;
- receipt contains no raw content or direct content fingerprints;
- unapproved free-text generation finish reasons cannot enter the receipt.

## Qualification boundary

This local validation proves only the self-contained package at the recorded source identity. It does not prove compatibility with the complete GitHub repository until exact-head GitHub Actions pass. It does not prove application integration, real worker enforcement, model quality, evidence accuracy, accessibility, packaging or release readiness.
