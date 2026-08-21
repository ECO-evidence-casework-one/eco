# Security and safe-use rules

Use AEIB only on software/systems you own, operate, or are explicitly authorized to assess.

The corpus is synthetic. It contains malformed files and hostile archive metadata.
Do not bulk-extract traversal/symlink fixtures with an unsafe extractor.

AEIB performs no network activity, does not execute corpus content, uses only Python's standard library,
caps fixture sizes, and writes only below the requested corpus directory.
