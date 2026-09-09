# Expected security invariants

1. Every top-level source is inventoried or explicitly rejected.
2. Source bytes are not modified.
3. Duplicate archive members remain distinguishable.
4. Unsafe archive paths never escape staging.
5. Archive links/symlinks are not blindly followed.
6. Malformed input does not become an untracked generic failure.
7. Partial/cancelled work is not reported as complete.
8. Repeat scans are stable.
9. Changed source bytes invalidate stale derived output.
10. Database/search-index integrity remains clean.
11. Resource ceilings fail closed.
12. Unsupported/encrypted/corrupt content is accounted for truthfully.
