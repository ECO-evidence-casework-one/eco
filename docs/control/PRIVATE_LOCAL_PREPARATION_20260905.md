# Private local Windows preparation — 5 September 2026

Refs #4/#24. This adds source preparation tooling, not an application feature, executable upload, public release or W8 acceptance result.

## Frozen application and qualified recipe

Application source remains `8b69a669b003fe30e84f1d344aa7533eb9cd9045` (merged #134). Its original main CI is run `33979546078` with all four jobs passed. W1-W7 remain complete within their mapped source/test scope; W8 and the other open A-checks remain unaccepted. A helper-only merge does not move the selected application candidate.

Standard Windows runner qualification `33980535294`, job `101344839890`, passed under Windows PowerShell 5.1. It compared 237 downloaded source files with an archive of the exact Git tree, ran eleven preparation controls, built an independent reference executable, then ran the actual preparer in a fresh child process with fresh source/compiler downloads and isolated caches. Go module verification, all Windows package tests, vet, two deterministic builds and the independent expected fingerprint check passed. An occupied output-route retry was refused and its existing executable hash stayed unchanged. The ECO graphical executable was not launched.

No executable, DLL, compiler, model or runnable ZIP was uploaded as an Actions artifact. The temporary qualification workflow was removed before this source-tooling PR. The remaining self-test script expects System.IO.Compression and System.IO.Compression.FileSystem to be loaded by its test host, as in the qualifier; the user-facing preparer loads both itself.

## Exact reference identities

- Source commit: `8b69a669b003fe30e84f1d344aa7533eb9cd9045`.
- Source archive SHA-256: `d5d56d07140857d4e5ffed91966d32d8510451191959b89235e048015b3c3934`.
- Go 1.23.12 Windows AMD64 archive SHA-256: `07c35866cdd864b81bb6f1cfbf25ac7f87ddc3a976ede1bf5112acbb12dfe6dc`.
- Local archive-recipe executable SHA-256: `d197748f861bd84b00776aa28c8712126091660fcc436e716f28126e05516dcc`; 4,772,352 bytes; Authenticode status NotSigned.
- Tested preparer byte SHA-256: `ef315afd101b42ec4afd85d25bc78d5be3565935ba8c3769183b6f65a45ad866` (original Windows/mixed line endings).
- Normalized Git preparer blob: `a60018612690f3ced6e1f215314d7e9a4c38eeb1`.

The local source-archive recipe explicitly uses `-buildvcs=false` and injects the full SourceCommit. It is deliberately NOT byte-identical to the original VCS-stamped main-CI executable `e77b9eb380fdadf4a9eeb233d5a946e27fa06497fad5dfd8146ed643a9844815`. It was separately qualified as above. Do not silently substitute these fingerprints.

The first preparation attempt `33980407941` failed because the PowerShell 5.1 self-test host had not explicitly loaded the compression assembly. No qualified preparer commit was made by that failed run. The successful run corrected assembly loading and native output routing before tests and committed only the qualified preparer correction.

## Handoff conditions

A private kit must provide the exact preparer, the fingerprint manifest, explicit recipient/purpose, unsigned warning and expiry/withdrawal status, together with the existing NOT RUN acceptance log. Personal recipient details and real local paths belong in the private handoff, not this public repository.

Preparation needs network access for the pinned public source/compiler/Go dependencies. The program remains a separate offline application; a networked development build is not proof of final runtime network behaviour. The preparer creates a new local directory, uses local caches and process-only environment settings, requests no account or elevation, and does not launch ECO or open personal workspaces. It refuses existing output paths and unexpected hashes. No system security settings or application-control bypass are supplied.

The result is a private unsigned core candidate for later controlled synthetic testing. Optional OCR/PDF renderer runtimes are not bundled or registered. First return the local BUILD_RESULT.txt; only then prepare the actual-machine launch/runtime steps. A successful hosted rehearsal is not the user's Acer result. The sixteen-case acceptance checklist is unchanged.

A private command wrapper, if supplied, must verify exact script/manifest fingerprints before execution and keep failures visible. Any file unblocking must target only that verified script, never a whole directory or ECO.exe. Permanent policy, Defender and Smart App Control must not be weakened to force execution. A blocked machine must report the exact message instead.
