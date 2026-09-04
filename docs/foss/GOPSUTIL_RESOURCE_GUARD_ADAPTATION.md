# gopsutil local resource guard adaptation

Date: 2026-09-04
Upstream: `shirou/gopsutil`
Pinned release: `v4.25.3`
Licence: BSD-3-Clause
ECO integration: direct Go library, local resource sampling only

## Why this pin

The current gopsutil development head inspected during this integration declares Go 1.24, while ECO's qualified module and Windows/Linux build gates remain on Go 1.23. ECO therefore does not follow upstream `master` for this dependency.

`v4.25.3` was selected because its published `go.mod` explicitly declares `go 1.23`. ECO pins that version in `go.mod` and commits `go.sum` verification material rather than silently floating to a newer release.

This is a compatibility pin, not a claim that v4.25.3 is permanently preferred. A later gopsutil upgrade requires the normal dependency/toolchain qualification again.

## Licence boundary

gopsutil is distributed under BSD-3-Clause. ECO must preserve the applicable copyright, licence conditions and disclaimer in the release notices/SBOM work for any binary that redistributes this dependency.

No endorsement by gopsutil authors or contributors is claimed.

## Imported capability

This slice imports only:

- `github.com/shirou/gopsutil/v4/cpu`
- `github.com/shirou/gopsutil/v4/mem`
- `github.com/shirou/gopsutil/v4/disk`

ECO uses those packages for local machine observations only:

- combined CPU-use sampling;
- total/available RAM and used percentage;
- free/total space for a caller-selected local working disk.

The first attached runtime is ECO's default local `llama.cpp` execution path. A GGUF model is fingerprinted first, then ECO samples local resources before starting model generation.

## Deliberately not imported

This slice does not use gopsutil's process enumeration, connections/network statistics, users, host inventory or remote-observability patterns. It does not transmit, upload or export telemetry and it does not add a server or network client to ECO.

The source-policy rule remains unchanged: ECO application code is not allowed to add network client/server primitives merely to report operational telemetry.

## Resource decision boundary

Resource monitoring is a guardrail, not a scheduler or hardware-health oracle.

Default hard launch blocks are deliberately narrow:

- available RAM below 256 MiB; or
- RAM usage at or above 98%; or
- free space on the relevant working disk below 256 MiB.

Default warnings are raised for:

- available RAM below 1 GiB;
- RAM usage at or above 92%;
- working-disk free space below 1 GiB;
- combined CPU usage at or above 95%.

High CPU use never blocks a launch by itself in this slice. CPU load can change quickly and delaying a local engine solely because of one sample would create false failures.

For `llama.cpp`, ECO also compares available RAM with the GGUF file size plus 512 MiB of conservative headroom. That comparison is warning-only. A GGUF file's byte size is not a reliable prediction of the runtime working set because actual memory use depends on model format, context, caches and runtime settings.

## Telemetry failure rule

Failure to obtain one ordinary resource metric becomes an explicit warning rather than a blanket application outage. Context cancellation/deadline errors still propagate normally.

This distinction prevents an observability defect from masquerading as evidence corruption or preventing unrelated work. A known critical sampled resource condition may block an engine; an unavailable metric does not invent a critical condition.

## llama.cpp audit integration

A successful grounded local-AI audit record now carries a compact resource preflight summary: resource level, blocked state, warning count and only the sampled CPU/RAM/disk numeric values needed to explain the decision.

If critical resources block `llama.cpp` before generation, ECO creates a separate authenticated `local-ai-resource-blocked` change record with `generation_started=false`. No trusted `QuestionRecord` is created.

The audit record does not prove that the machine was healthy, that a model would have succeeded, or that the resource sample remained unchanged after capture.

## Tests

The integration tests cover:

- critical RAM fail-closed behavior;
- critical working-disk fail-closed behavior;
- high-CPU warning without CPU-only blocking;
- warning-only GGUF memory headroom;
- metric-unavailable degradation;
- regular-file to parent-disk path resolution and relative-path rejection;
- live CPU/RAM/disk sampling on the qualified CI platforms;
- context and sample-interval bounds;
- propagation of resource warnings through the grounded local-AI result/audit boundary;
- critical-resource refusal without creation of a trusted question record.

## Non-claims and future use

This slice does not claim to enforce per-process CPU/RAM quotas, prevent out-of-memory termination, predict model performance, diagnose hardware faults or provide security monitoring.

The generic `CheckLocalEngineResources` boundary can later be reused by Docling/OCR pipelines where the operating policy is appropriate. That reuse should be deliberate per engine rather than silently wrapping every subprocess with identical assumptions.
