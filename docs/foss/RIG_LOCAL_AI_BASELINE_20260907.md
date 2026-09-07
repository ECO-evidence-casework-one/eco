# ECO rig local-AI baseline — 2026-09-07

## Purpose

This record controls the first rig-hosted local generative-AI preview after PR #139 connected the application-facing Ask ECO route to ECO's grounded llama.cpp workflow and PR #141 made real Qwen/fallback state visible in the native Ask control.

The goal is compatibility first: reproduce the Qwen model family that previously ran on the low-spec laptop, but use the current ECO grounding/verification architecture and current source rather than restoring an older application build.

## GitHub-first donor search/result

### Runtime — ADOPT (bounded)

- Upstream: `ggml-org/llama.cpp`
- Licence: MIT
- Release tag: `b10259`
- Source commit: `1269cb1ff1598751f846241be90083ae9ad036fb`
- Windows baseline asset: `llama-b10259-bin-win-cpu-x64.zip`
- GitHub-published asset SHA-256: `6613d8d56263233ef800fb8f8135231adb5eb851b40558281190242b0b20556b`
- Use: local CPU-only `llama-cli.exe` plus its release DLLs.
- Boundary: no llama-server, RPC, remote model URL or network inference path is used by ECO.

CPU is deliberately the first rig gate because it changes only the host machine, not both machine and accelerator backend at once. GPU/Vulkan/CUDA qualification is a later bounded optimization slice after the working baseline and hardware receipt exist.

### Build toolchain — QUALIFY

- Distribution: `actions/go-versions` GitHub release `1.23.12-16792118003`
- Asset: `go-1.23.12-win32-x64.zip`
- GitHub-published asset SHA-256: `c27b02f15d4ceb89fbce6ffe2a28df3dd293608cf79e9f12839f672863622845`
- Purpose: portable build toolchain only; it is not bundled into the final preview application directory.

### Model weights — authoritative exception after GitHub-first search

The model weights are not sourced from an unofficial GitHub mirror. The authoritative publisher is Qwen's official Hugging Face repository:

- Publisher/repository: `Qwen/Qwen2.5-1.5B-Instruct-GGUF`
- Licence reported by publisher: Apache-2.0
- File: `qwen2.5-1.5b-instruct-q4_k_m.gguf`
- Published SHA-256: `6a1a2eb6d15622bf3c96857206351ba97e1af16c30d7a74ee38970e434e9407e`
- Size class: about 1.12 GB

This is the same controlled model identity selected for the earlier laptop baseline. Using the publisher's authoritative weights is safer than adopting a third-party GitHub re-upload merely to keep every byte on one hosting platform.

## ECO application/source identity

The rig preparer is pinned to merged application source commit:

`b03ec2358dbf437deb922ad1cbb96d4e5c6faedb`

That source includes:

- PR #139 — application-facing Ask ECO route to verified local llama.cpp/Qwen with deterministic fallback;
- PR #141 — truthful `QWEN READY`, `QWEN RUNNING`, `QWEN CHECKED`, unavailable/rejected and source-fallback presentation through the native read-only Ask control.

The private rig preparer intentionally uses a reproducible archive-source recipe: it downloads the exact GitHub source archive, disables automatic VCS stamping with `-buildvcs=false`, uses `-trimpath`, clears the Go build ID and injects the full source commit through ECO's `SourceCommit` linker value.

An independent fresh Windows qualification first observed, then a later run reproduced exactly, this archive-source candidate identity:

- archive-source `ECO.exe` SHA-256: `eb6159cb0406a0d1b7195285f03848048026726e6abe3c73b9f7a4f52a9fbee3`
- size: `4,892,672` bytes
- source tests/vet before build: PASS

GitHub Actions run `34099322645` / #491 then passed the full package gate on that pinned source/identity:

- Linux tests/vet: PASS
- source policy: PASS
- secret scan: PASS
- rig-preparer self-test: PASS
- ordinary Windows deterministic build/tests: PASS
- independent archive-source ECO build reproduction: PASS
- pinned GitHub llama.cpp runtime route: PASS
- SBOM reconciliation: PASS
- private signing/tamper-rejection rehearsal: PASS

The archive-source executable is controlled by its own recipe/fingerprint and is not represented as byte-identical to a Git-checkout Actions executable.

## One-click preparer gates

`scripts/prepare-rig-ai-preview-v3.ps1` must, in order:

1. collect a non-invasive hardware receipt;
2. obtain the exact merged ECO source commit from GitHub;
3. obtain and verify the pinned GitHub-hosted Go toolchain;
4. verify Go module dependencies, run tests and vet;
5. build the archive-source ECO recipe twice and require SHA-256 `eb6159cb0406a0d1b7195285f03848048026726e6abe3c73b9f7a4f52a9fbee3`;
6. obtain and verify the exact llama.cpp GitHub release archive;
7. obtain and verify the official Qwen model SHA-256;
8. run a real CPU-only/offline Qwen generation smoke test;
9. create a process-only launcher that points current ECO at the verified runtime/model and redirects `LOCALAPPDATA` to isolated `PreviewUserData`;
10. launch only after all preceding gates pass.

`START_RIG_AI_PREVIEW.cmd` prefers `E:\ECO_RIG_AI_PREVIEW` when the E: drive is available, so model/build data does not default to the Windows system drive on the rig. It uses a local sibling folder if E: is unavailable.

No administrator elevation, registry mutation, Defender/Smart App Control change, cloud AI, server/RPC inference or real-evidence permission is part of this slice.

## Acceptance boundary

CI proves the preparer, exact archive-source ECO build recipe and GitHub runtime route. CI intentionally does not download the roughly 1.12 GB model on every run.

The feature is **not yet accepted as working on the user's rig**. That requires both:

1. the rig preparer reporting `Real offline Qwen generation: PASS`; and
2. the current ECO Ask screen demonstrating `QWEN READY` and a successful grounded model answer as `QWEN CHECKED` (or truthfully reporting the fallback/rejection state instead).

The wider modern/accessibility work remains independently gated under Issue #7 and Product Lab PR #142. Keyboard Tab/Shift+Tab remediation is separately isolated in draft PR #143 pending physical acceptance, despite its CI passing.
