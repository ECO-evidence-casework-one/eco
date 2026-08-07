#!/usr/bin/env python3
"""Close MinGW/UCRT64 runtime dependencies for the packaged Qt UI-0 folder.

windeployqt deploys Qt libraries/plugins but not every transitive MSYS2 runtime
DLL. This script walks every staged PE binary once, copies any dependency found
in /ucrt64/bin, and queues the copied DLL for the same check. Windows system
DLLs are intentionally ignored because they do not exist in the UCRT64 map.
"""

from __future__ import annotations

import pathlib
import re
import shutil
import subprocess
import sys

DLL_RE = re.compile(r"^\s*DLL Name:\s*(.+?)\s*$", re.IGNORECASE)


def dependency_names(binary: pathlib.Path) -> list[str]:
    proc = subprocess.run(
        ["objdump", "-p", str(binary)],
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.DEVNULL,
        check=False,
    )
    if proc.returncode != 0:
        return []
    out: list[str] = []
    for line in proc.stdout.splitlines():
        m = DLL_RE.match(line)
        if m:
            out.append(m.group(1).strip())
    return out


def staged_index(root: pathlib.Path) -> dict[str, pathlib.Path]:
    return {
        p.name.lower(): p
        for p in root.rglob("*")
        if p.is_file() and p.suffix.lower() in {".exe", ".dll"}
    }


def main() -> int:
    if len(sys.argv) != 3:
        print("usage: close_runtime.py <package-root> <ucrt64-bin>", file=sys.stderr)
        return 2

    package = pathlib.Path(sys.argv[1]).resolve()
    runtime = pathlib.Path(sys.argv[2]).resolve()
    if not package.is_dir() or not runtime.is_dir():
        print("package or runtime directory missing", file=sys.stderr)
        return 2

    available = {p.name.lower(): p for p in runtime.glob("*.dll") if p.is_file()}
    staged = staged_index(package)
    queue = list(staged.values())
    checked: set[pathlib.Path] = set()

    while queue:
        binary = queue.pop(0)
        if binary in checked:
            continue
        checked.add(binary)
        for dep in dependency_names(binary):
            key = dep.lower()
            if key in staged:
                continue
            src = available.get(key)
            if src is None:
                continue  # Windows/system dependency, deliberately not bundled.
            dst = package / src.name
            print(f"Deploying transitive runtime: {src.name} (required by {binary.relative_to(package)})")
            shutil.copy2(src, dst)
            staged[key] = dst
            queue.append(dst)

    # Second pass: any dependency supplied by UCRT64 must now be present somewhere
    # in the staged package. This catches incomplete closure deterministically.
    missing: list[tuple[pathlib.Path, str]] = []
    for binary in list(staged.values()):
        for dep in dependency_names(binary):
            key = dep.lower()
            if key in available and key not in staged:
                missing.append((binary, dep))

    if missing:
        for binary, dep in missing:
            print(f"ERROR: missing {dep} required by {binary.relative_to(package)}", file=sys.stderr)
        return 1

    print(f"PASS: runtime closure complete; checked {len(checked)} PE files; staged {len(staged)} EXE/DLL files")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
