# Private local Windows preparation — 5 September 2026

Refs #4/#24. Source preparation tooling only: no application feature, public executable upload, release approval or W8 acceptance result.

## Frozen application

Application source stays `8b69a669b003fe30e84f1d344aa7533eb9cd9045` (merged #134). Original main CI `33979546078` passed all four jobs. W1-W7 remain complete within their mapped source/test scope; W8 and other open A-checks remain unaccepted. A helper-only merge does not replace the selected application candidate.

## Completed native preparation evidence

Initial successful standard Windows PowerShell 5.1 qualification `33980535294`, job `101344839890`, compared 237 source-archive files with the exact Git tree, exercised eleven preparation controls, built an independent reference, and ran the user-facing preparer in a fresh child process with fresh downloads/caches. Module verification, all Windows package tests, vet, repeat build and expected fingerprint checks passed. Reusing the occupied output route was refused without changing its executable.

Review then found two problems: the first EXE was placed in candidate before every check passed, and the standalone self-test depended on its host preloading System.IO.Compression. Both were corrected. The builder now prepares the executable, receipt and licence in build-staging and moves that entire directory to candidate only after all checks pass. The self-test loads both assemblies itself.

Final corrected qualification **33981640717** passed:

- 11/11 standalone controls in a clean Windows PowerShell 5.1 process;
- the actual fingerprint-bound command launcher with a synthetic Internet-origin mark and BUILD confirmation, including fresh source/compiler downloads, package tests, vet, repeat builds and exact reference match;
- refusal of a tampered preparer before unblocking it or creating build output;
- a deliberately wrong post-build executable fingerprint: preparation stopped and the candidate directory remained absent.

The corrected preparer is exactly SHA-256 `1ace2fc82630357f09c7e1f9db8e9f7be1ca78f0c97d86c75751336027b0d79e` (UTF-8, LF). The final receipt is `PRIVATE_PREPARATION_REVIEW_FIX_20260905.json`. The source-only launcher template receives exact script and private-manifest fingerprints when the private kit is assembled; it contains no personal recipient details.

The ECO graphical executable was not launched, signed or publicly uploaded. Temporary qualification workflows were removed before final source review.

## Exact identities and recipe distinction

- Application source: `8b69a669b003fe30e84f1d344aa7533eb9cd9045`.
- Source ZIP SHA-256: `d5d56d07140857d4e5ffed91966d32d8510451191959b89235e048015b3c3934`.
- Go 1.23.12 Windows AMD64 ZIP SHA-256: `07c35866cdd864b81bb6f1cfbf25ac7f87ddc3a976ede1bf5112acbb12dfe6dc`.
- Qualified local EXE SHA-256: `d197748f861bd84b00776aa28c8712126091660fcc436e716f28126e05516dcc`; 4,772,352 bytes; NotSigned.

This archive-source recipe uses buildvcs=false and injects the full SourceCommit. It is deliberately not byte-identical to the original VCS-stamped CI executable `e77b9eb380fdadf4a9eeb233d5a946e27fa06497fad5dfd8146ed643a9844815`. Its own Windows tests and expected-fingerprint checks passed. The earlier preparer hash `ef315afd...` is historical and must not be substituted for the corrected helper.

## Failed attempts retained honestly

Run 33980407941 failed on the missing compression assembly. Run 33981390904 passed the corrected standalone tests, launcher, tamper refusal and wrong-hash/candidate-absence assertions, but the intentionally failing child's exit 1 propagated through the runner shell and stopped the overall job before product commit. Final run 33981640717 resets that expected code only after all negative-test assertions succeed. Neither failed workflow is represented as an overall success.

## Private handoff and security boundary

Supply the exact helper, fingerprint manifest, recipient/purpose, unsigned warning, expiry/withdrawal status and the existing NOT RUN acceptance log privately. Preparation needs the network for pinned public source/compiler/dependencies. It uses a new local directory, isolated caches and process-only settings; it asks for no account or elevation and does not open personal workspaces or launch ECO.

The launcher verifies the preparer and manifest fingerprints before removing the Internet-origin mark only from that verified script. RemoteSigned is limited to the spawned process; no permanent policy, Defender or Smart App Control setting is weakened and ECO.exe is not unblocked. A machine-level block must be reported rather than bypassed.

Only after all validation succeeds does candidate/ECO.exe appear. Internal staging files after a failure are not approved candidates and must not be run. Optional OCR/PDF renderer runtimes are not bundled or registered. The next human action is local preparation and return of BUILD_RESULT.txt, not the whole sixteen-case application test. All actual Acer acceptance cases remain NOT RUN. No public distribution, real-evidence use, independent human approval or overall completion percentage follows from this kit.
