# AEIB v0.1 — Adversarial Evidence Ingestion Benchmark

**Status:** private prototype / defensive benchmark  
**Scope:** document, email and archive ingestion robustness  
**Safety:** synthetic data only; no malware, credentials, personal evidence or third-party exploit payloads.

AEIB is a deterministic hostile-input corpus for applications that ingest evidence-like files.

> A bad source file is allowed. An unaccounted-for source file is not.

## v0.1 goals

Test whether an ingestion engine can account for hostile or malformed material without silently skipping
sources, mutating source bytes, confusing duplicate members, extracting unsafe paths, reporting interrupted
work as complete, reusing stale derived data, or corrupting its own index/state.

## Quick start

    python RUN_AEIB.py generate
    python RUN_AEIB.py validate
    python -m unittest discover -s tests -v

AEIB v0.1 contains no live malware and does not execute corpus files.
