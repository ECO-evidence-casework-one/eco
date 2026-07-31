#!/usr/bin/env python3
"""Fail CI if the native application source gains undeclared network code or secrets."""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
GO_ROOTS = [ROOT / "cmd", ROOT / "internal"]
FORBIDDEN_IMPORTS = {
    "net",
    "net/http",
    "net/url",
    "net/rpc",
    "net/smtp",
}
FORBIDDEN_RUNTIME_MARKERS = (
    "listenandserve(",
    "http.get(",
    "http.post(",
    "dialtcp(",
    "websocket",
)
SECRET_PATTERNS = (
    re.compile(r"(?i)github_pat_[a-z0-9_]{20,}"),
    re.compile(r"(?i)ghp_[a-z0-9]{20,}"),
    re.compile(r"(?i)signpath[_-]?(api[_-]?)?token\s*[:=]\s*['\"][^'\"]+"),
)

failures: list[str] = []
for root in GO_ROOTS:
    for path in sorted(root.rglob("*.go")):
        text = path.read_text(encoding="utf-8")
        rel = path.relative_to(ROOT)
        for imp in FORBIDDEN_IMPORTS:
            if re.search(rf'^[ \t]*"{re.escape(imp)}"[ \t]*$', text, re.MULTILINE):
                failures.append(f"{rel}: forbidden network import {imp!r}")
        low = text.lower()
        for marker in FORBIDDEN_RUNTIME_MARKERS:
            if marker in low:
                failures.append(f"{rel}: forbidden network/server marker {marker!r}")
        for pattern in SECRET_PATTERNS:
            if pattern.search(text):
                failures.append(f"{rel}: possible embedded credential")

if failures:
    print("ECO source-policy check FAILED", file=sys.stderr)
    for failure in failures:
        print(f"- {failure}", file=sys.stderr)
    raise SystemExit(1)

print("ECO source-policy check PASS: no application network imports or embedded credential patterns found.")
