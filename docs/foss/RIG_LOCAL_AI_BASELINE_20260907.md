# ECO rig local-AI baseline — 2026-09-07

## Purpose

This record controls the first rig-hosted local generative-AI preview after PR #139 connected the application-facing Ask ECO route to ECO's grounded llama.cpp workflow.

The goal is compatibility first: reproduce the model family that previously ran on the low-spec laptop, but use the current ECO grounding/verification architecture and current main source rather than restoring an older application build.

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

The preparer is pinned to merged application source commit:

`24170e4505cfaadf97d37bc172ebf3097af7de55`

Post-merge GitHub Actions run #471 qualified a **Git-checkout** Windows artifact from that source:

- Actions `ECO.exe` SHA-256: `190d06468a9cf282c1837ee08bf854f20724d42e70fc1e0bcb44479a015f36ba`
- Size: `4,880,384` bytes
- Linux tests/vet: PASS
- source policy: PASS
- secret scan: PASS
- Windows deterministic rebuild/tests: PASS
- SBOM reconciliation: PASS
- private signing/tamper rehearsal: PASS

The private rig preparer intentionally uses a different reproducible recipe: it downloads the exact GitHub source archive, disables automatic VCS stamping with `-buildvcs=false`, uses `-trimpath`, clears the Go build ID and injects the full source commit through ECO's `SourceCommit` linker value.

That archive-source recipe is **source-equivalent but not byte-identical** to the Git-checkout Actions build. CI run #478 demonstrated the resulting archive-source candidate identity:

- archive-source `ECO.exe` SHA-256: `8ca12dafdd78182d0984aafebed2b7ed0894b3471a45f7d1e0602b67ac382426`
- size: `4,880,384` bytes
- source tests/vet before build: PASS

This separate identity is deliberate. The preparer must reproduce the archive-source hash twice on an independent fresh Windows runner before the package is accepted.

## One-click preparer gates

`scripts/prepare-rig-ai-preview-v3.ps1` must, in order:

1. collect a non-invasive hardware receipt;
2. obtain the exact merged ECO source commit from GitHub;
3. obtain and verify the pinned GitHub-hosted Go toolchain;
4. verify Go module dependencies, run tests and vet;
5. build the archive-source ECO recipe twice and require its independently qualified SHA-256;
6. obtain and verify the exact llama.cpp GitHub release archive;
7. obtain and verify the official Qwen model SHA-256;
8. run a real CPU-only/offline Qwen generation smoke test;
9. create a process-only launcher that points current ECO at the verified runtime/model and redirects `LOCALAPPDATA` to isolated `PreviewUserData`;
10. launch only after all preceding gates pass.

No administrator elevation, registry mutation, Defender/Smart App Control change, cloud AI, server/RPC inference or real-evidence permission is part of this slice.

## Acceptance boundary

CI can prove the preparer, exact archive-source ECO build recipe and GitHub runtime route. CI intentionally does not download the 1.12 GB model on every run.

The feature is **not** accepted as working on the rig until the preparer reports a real local Qwen generation PASS on that rig and the current ECO UI successfully completes an Ask ECO turn using the configured local-AI path. UI status/engine visibility remains a separate product-visible follow-up; silent deterministic fallback must not be mistaken for proof that Qwen ran.
