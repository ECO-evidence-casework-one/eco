# AEIB v0.1 Historical Method Note

AEIB is informed by earlier synthetic hostile-input work carried out while developing local evidence-ingestion tooling.

That prior work established useful methodological rules that AEIB retains:

- deterministic seed-controlled corpora;
- fresh-process hostile execution for certification runs;
- explicit time/resource ceilings;
- scan the same corpus more than once to test stability;
- treat every source object as accounted for rather than silently skipped;
- preserve source bytes unchanged;
- stop and repair when a hostile seed exposes a real defect;
- replay the exact failing seed before restarting the wider campaign;
- invalidate old green results after material scanner changes.

Historical examples included malformed document/archive/email inputs, duplicate member names, nested containers, source mutation during processing, crash/recovery state, Unicode/encoding failures and database/index integrity checks.

This note is provenance, not a claim that the current AEIB v0.1 corpus reproduces every historical test. AEIB v0.1 begins with a smaller product-neutral 22-case foundation and expands only under controlled review.
