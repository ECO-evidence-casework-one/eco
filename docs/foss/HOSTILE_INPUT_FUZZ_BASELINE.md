# Hostile-input fuzz baseline

Date: 2026-09-04

ECO's FOSS donor Wave 2 acquired and licence-qualified mature fuzzing references including `dvyukov/go-fuzz` (Apache-2.0, commit `e577bee5275c42f9a3ff598ade77fe9806191345`) and Google's `oss-fuzz` (Apache-2.0, commit `b2eee1cf755e4828a81a7949232e1bc5452d92fc`). AFL++ was also acquired for study/external development use, but its AGPL-3.0 licence means ECO does not copy AFL++ code into the application.

For the current Go parser boundaries, ECO uses the Go toolchain's built-in fuzz engine rather than adding another fuzz runtime dependency. The donor projects remain architecture/corpus/harness references; the permanent ECO tests are small native fuzz targets that normal `go test` executes deterministically against seed inputs and dedicated qualification workflows can mutate for bounded periods.

Initial fuzz boundaries:

- ZIP end-of-central-directory preflight plus bounded ZIP inspection;
- EML / MIME readable-message extraction;
- XML text extraction used by Office/OpenDocument container readers;
- file-type byte sniffing.

Inputs are capped to 1 MiB during mutation runs, and each target asserts a bounded-output invariant where the parser returns text. These targets are intended to detect panics, hangs, unbounded expansion and unsafe parser assumptions; they do not replace deterministic regression tests, preserved-original controls, or the main Windows release-integrity gates.

## Qualification result

A separate Go 1.23.12 qualification run (`33893969593`) first passed deterministic seed tests and then fuzzed each boundary for approximately 20 seconds with two workers:

- ZIP preflight/inspection: 114,860 mutated executions, 28 new interesting inputs, PASS;
- EML/MIME reader: 107,571 mutated executions, 80 new interesting inputs, PASS;
- XML/Office readable-text parser: 233,430 mutated executions, 71 new interesting inputs, PASS;
- file-type sniffer: 467,989 mutated executions, 107 new interesting inputs, PASS.

Total: 923,850 mutated executions across the four boundaries with no crash or asserted bound violation.

SQLite corruption fuzzing is intentionally not part of the current ECO baseline because current ECO does not use SQLite. That older reconnaissance item is therefore obsolete unless SQLite is introduced later.
