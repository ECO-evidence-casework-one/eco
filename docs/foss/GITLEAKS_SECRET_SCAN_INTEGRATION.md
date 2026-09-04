# Gitleaks CI secret-scan integration

Date: 2026-09-04

## Mission fit

This is a GitHub/FOSS-first release-security integration. It adds no ECO runtime feature and no end-user network dependency. Its purpose is to prevent credentials, tokens and similar secrets from being merged into the checked-out ECO source tree.

## Upstream

- Project: `gitleaks/gitleaks`
- Release: `v8.30.1`
- Release published: 2026-03-21
- Licence: MIT
- CI asset: `gitleaks_8.30.1_linux_x64.tar.gz`
- Required SHA-256: `551f6fc83ea457d62a0d98237cbad105af8d557003051f41f3e7ca7b3f2470eb`

The workflow does not use a floating `latest` tag or an unpinned third-party GitHub Action. It downloads the exact upstream release asset directly from GitHub and verifies the asset SHA-256 before extraction/execution.

## ECO boundary

`.github/workflows/ci.yml` adds a `secret-scan` job that:

1. checks out the exact ECO commit being qualified;
2. downloads only the pinned Linux x64 Gitleaks archive;
3. verifies its SHA-256 with `sha256sum --check --strict`;
4. extracts only the `gitleaks` binary into the ephemeral runner temp directory;
5. runs `gitleaks dir --redact --verbose --no-banner .` against the checked-out repository;
6. fails closed on download, checksum, extraction or secret-scan failure.

The Windows build depends on `source-policy`, `secret-scan` and `test-linux`, so packaging cannot proceed when the secret scan is red.

## First qualification finding

The first exact-head run correctly failed on one `generic-api-key` candidate in the historical OKFN partnership assessment. Inspection showed that the candidate was prose — the governance phrase `key recovery, SBOM/provenance ...` — rather than a credential or token.

The finding was not suppressed globally and `generic-api-key` remains enabled. `.gitleaks.toml` contains one rule-specific allowlist entry requiring BOTH:

- the exact historical assessment path; and
- the exact known prose line.

Changing either the file path or the line content causes the exception not to match. This preserves the strict default rule elsewhere.

## Deliberate limitations

- This integration scans the checked-out source tree, not every historical Git commit.
- Any future allowlist change must be justified by a proven false positive and narrowly scoped.
- The scan does not prove that a secret has never existed in repository history, forks, caches, build logs or external systems.
- The scan does not replace GitHub secret protection, credential rotation, code review or release signing.
- No Gitleaks binary is committed into ECO and no scanner is shipped to end users.

## Next release-security lane

After this gate is qualified and merged, the next planned FOSS release-security integration is Sigstore/Cosign plus build provenance/manifest binding. That work must remain separate so secret detection and release signing have independently testable failure modes.
