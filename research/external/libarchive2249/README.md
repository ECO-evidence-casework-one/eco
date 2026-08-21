# libarchive #2249 clean-room validation

This research lane validates a possible upstream contribution for libarchive issue #2249: a RAR5 archive containing an uncompressed ZIP can be auto-detected as the inner ZIP because the exact RAR5 signature currently bids below the seekable ZIP bidder.

## Frozen upstream

`libarchive/libarchive@2e5d7befd63c02200e2ad6295b8ff95de3cf922d`

That SHA was the live `master` head when this validation lane was prepared on 21 August 2026.

## Candidate change

- raise the exact eight-byte RAR5 standard-signature bid from 30 to 64;
- perform that exact-signature check before the existing `best_bid > 30` early return;
- keep the lower-confidence SFX scan at bid 30;
- add a regression using the synthetic uuencoded fixture supplied in issue #2249;
- require the selected format to be `ARCHIVE_FORMAT_RAR_V5` and the outer entry to be `test.zip`.

## Controls

The GitHub Actions validation job first builds the unmodified frozen upstream and requires the issue fixture to be identified as ZIP. It then applies the candidate, rebuilds, requires RAR detection, and runs the new libarchive regression test.

No upstream repository is modified by this lane. No binary artifact is uploaded. No CVE/security-severity claim is made; this is a format-identification correctness contribution requested by the upstream maintainer.
