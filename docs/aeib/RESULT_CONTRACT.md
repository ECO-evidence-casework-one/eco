# AEIB v0.1 Result Contract

Each adapter should emit one result per case with:

- `case_id`
- `engine`
- `engine_version`
- `engine_build_sha256` or commit
- `started_at`
- `finished_at`
- `expected`
- `observed`
- `status`: `PASS`, `FAIL`, `PARTIAL`, or `BLOCKED`
- `exception_class`
- `source_modified`
- `pending_children`
- `notes`

Global run checks should include:

1. closed-set input accounting;
2. source immutability;
3. database/index integrity where applicable;
4. repeat-scan stability where applicable;
5. bounded cancellation/resource behavior;
6. no extraction outside the designated staging root.

A negative security result is valid evidence. Do not tune the benchmark to make a product pass.
