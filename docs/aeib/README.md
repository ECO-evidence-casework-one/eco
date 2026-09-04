# AEIB — Adversarial Evidence Ingestion Benchmark

AEIB is a safe, deterministic synthetic benchmark for software that ingests untrusted documents, archives, email, and evidence-like corpora.

## v0.1 goals

AEIB tests whether an ingestion system can account for hostile or malformed input without silently losing provenance, corrupting state, following unsafe paths, or modifying source data.

The initial corpus contains 22 synthetic cases spanning encoding, signature mismatch, duplicate archive members, nested/corrupt archives, unsafe path metadata, malformed email/MIME, duplicate attachments, nested email/archive structures, malformed OOXML-like containers, deterministic binary garbage, a safe high-compression fixture, and source-mutation fixtures.

## Safety

AEIB contains **no live malware, credentials, personal data, persistence, command-and-control code, or undisclosed third-party exploit payloads**.

## Quick start

```bash
python tools/aeib/generate.py --output corpus --seed 20260821
python tools/aeib/verify.py corpus
python tools/aeib/test_determinism.py
```

## Interpretation

AEIB does not declare a product secure. It supplies reproducible hostile fixtures and expected security behavior. Each application needs an adapter that maps its own observed results to the result contract.

## Origin

The benchmark design is derived from repeated synthetic hostile-input testing performed during development of local evidence-ingestion tooling. It is intentionally product-neutral so other document/evidence systems can use it.

AEIB code in this repository is governed by the repository's existing open-source licence unless a later standalone release states otherwise.
