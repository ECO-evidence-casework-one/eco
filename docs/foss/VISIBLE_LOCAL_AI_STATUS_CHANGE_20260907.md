# Visible local-AI status — implementation boundary

This change does not add a model, network path, server, remote API or new third-party runtime.

It reuses ECO's existing local llama.cpp/Qwen configuration and grounded-answer workflow from merged PR #139. The new backend status function verifies the same configured runtime/model identities and probes llama.cpp with `--offline --version`; it does not register tools or alter workspace state merely to report readiness.

The Windows presentation layer uses the existing standard read-only Ask `EDIT` control. No new donor code is copied into the application in this slice.

Release/acceptance remains blocked until real rig Qwen generation and the wider accessibility/UI work are complete.
