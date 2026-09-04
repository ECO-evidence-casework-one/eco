# llama.cpp local grounded adapter for ECO

Date: 2026-09-04

## Upstream source reviewed

- Project: `ggml-org/llama.cpp`
- Pinned source commit: `0ef4d560e12c1a46470265c1abd31dd47c777d23`
- Upstream licence: MIT
- Primary reviewed CLI contract: `tools/cli/README.md`
- Executable family reviewed: `tools/cli`

The Wave-1 acquisition audit separately records the exact upstream commit and licence-file hash.

## Exact pinned CLI controls ECO uses

The pinned `llama-cli` contract supports:

- `--model FNAME` for a local model path;
- `--file FNAME` for a local prompt file;
- `--json-schema-file FILE` to constrain generation to a supplied JSON schema;
- `--offline` to prevent model/network retrieval;
- `--simple-io` for subprocess/limited-console compatibility;
- `--no-display-prompt` so model-facing evidence text is not echoed into stdout;
- deterministic sampling controls including `--seed`, `--temp`, `--top-k`, `--top-p`, and `--min-p`;
- bounded generation via `--n-predict` and explicit context size via `--ctx-size`.

The same pinned CLI also exposes network-capable mechanisms such as model URLs, Hugging Face repositories, RPC servers and a server-base URL. ECO's adapter deliberately never passes those arguments.

## ECO execution boundary

`internal/eco/llamacpp_adapter.go`:

1. Requires an explicit absolute local `llama-cli` executable path.
2. Requires an explicit absolute local regular `.gguf` model file.
3. Fingerprints the GGUF model before generation and checks that the file did not change during the run.
4. Filters inherited `LLAMA_ARG_*` variables so environment variables cannot inject RPC, model URL, Hugging Face or other llama.cpp arguments.
5. Removes common model/token/proxy environment variables and forces llama.cpp/Hugging Face offline settings.
6. Writes the prompt and JSON schema to a private temporary directory rather than putting evidence text on the command line.
7. Captures stdout/stderr into bounded buffers and rejects overflow.
8. Parses stdout as one strict JSON object with unknown fields and trailing output rejected.
9. Uses no shell invocation and no local HTTP server.

## Grounded release boundary

`internal/eco/llamacpp_workflow.go` connects the adapter to the Ethos-inspired grounding layer already merged into ECO.

The local model is **not** trusted to author the released answer. It receives only the bounded model-facing `GroundingContext`, and returns a schema-constrained `GroundingEmission` containing a draft plus source claims.

ECO then:

1. resolves every claim only through the exact app-owned grounding context shown to that model run;
2. rejects invented evidence/segment IDs;
3. requires quote/value text to exist in the shown source text;
4. re-verifies the preserved encrypted source objects;
5. rejects the whole result when any claim fails;
6. discards the model's free-form draft for release purposes;
7. deterministically renders the trusted `QuestionRecord.Answer` only from successfully grounded citation text;
8. writes the engine version, model name/hash, context ID and `model_draft_released=false` into ECO's authenticated change chain.

A rejected emission creates no trusted `QuestionRecord`; only a rejection event is added to the audit chain.

## Deliberate first-slice limits

- CPU-only invocation (`--device none`, zero GPU layers) is used for predictable low-spec operation. Resource-aware GPU selection can be added later behind explicit controls.
- Prompt size is capped at 64 KiB and generated output at 128 KiB.
- Generation is capped at 2048 tokens with a 16,384-token context request.
- Only local GGUF files are accepted. Model downloading/installing is outside this runtime path.
- A green result proves textual grounding, not source truth, completeness, relevance, legal correctness or sound reasoning.

## Why the free-form draft is not released

Citation grounding alone cannot prove that every sentence in a model-written paragraph is covered by the claims it emitted. A model could otherwise write one unsupported sentence while attaching a different valid citation. This first integration therefore uses the model for source selection/reasoning but reconstructs the released answer from verified source wording. Later natural-language synthesis should be added only with a stronger deterministic coverage control.
