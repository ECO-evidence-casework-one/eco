# Current ECO project status

**Status date:** 1 August 2026  
**Public repository:** `ECO-evidence-casework-one/eco`  
**Current public source milestone:** `ECO-V25-20260731-N2-P1`  
**Current public commit reviewed:** `3bf2f78c31f0341d0d0be8eb011d8e0c473f35bb`

## Public source position

The source currently published on `main` remains the Version 25 N2 P1 native document-vision foundation.

It is an early source-development snapshot. It is not an approved stable release and it does not include an approved bundled OCR engine or generative local language model.

## Later local development

Later native Windows candidates have been developed and independently inspected outside `main`.

### V32 M1 P1

`ECO-V32-20260801-M1-P1` was independently reproducible, but it was held before Windows execution.

It was not promoted because the supplied evidence was not sufficient to approve:

- the exact llama.cpp runtime and every redistributed runtime file;
- execution-time runtime verification;
- the final self-contained model-and-runtime executable;
- current V32 source-manifest and SBOM records;
- complete V32 end-to-end release evidence.

V32 is not a public release and must not be used with real evidence.

### V34

Version 34 is under active development.

It has not yet entered independent intake. It is not a source milestone, release candidate or public release. It must receive a new build identity and pass the independent source, runtime, packaging, Windows, accessibility, offline and synthetic-evidence gates.

## Release position

There is currently no approved signed end-user release.

An official Windows release requires:

- exact public source and immutable commit;
- complete corresponding source;
- accurate manifest and SPDX SBOM;
- pinned and verified runtime and model components;
- reproducible build evidence;
- independent Acer N4500 testing;
- zero-network evidence;
- accessibility evidence;
- trusted Authenticode signing;
- no unresolved critical or high-priority defect.

## Evidence restrictions

Use synthetic information only.

Do not upload real evidence, private correspondence, credentials, vaults, personal case records or unapproved executables to this repository.

## Status authority

This document records project status. `VERSION` continues to identify the latest source milestone approved for `main`; it must not be advanced merely because a later private candidate exists.
