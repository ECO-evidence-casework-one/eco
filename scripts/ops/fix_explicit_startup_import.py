from pathlib import Path

p = Path("cmd/eco/main_other.go")
text = p.read_text(encoding="utf-8")
old = '\t"os"\n\t"path/filepath"\n'
if text.count(old) != 1:
    raise SystemExit(f"expected one main_other import anchor, found {text.count(old)}")
p.write_text(text.replace(old, '\t"os"\n', 1), encoding="utf-8")
