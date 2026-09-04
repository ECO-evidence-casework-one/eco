#!/usr/bin/env python3
from pathlib import Path
import argparse, hashlib, json

def sha256_file(path):
    h=hashlib.sha256()
    with open(path,"rb") as f:
        for b in iter(lambda:f.read(1024*1024), b""):
            h.update(b)
    return h.hexdigest()

def main():
    ap=argparse.ArgumentParser()
    ap.add_argument("corpus")
    args=ap.parse_args()
    root=Path(args.corpus)
    m=json.loads((root/"manifest.json").read_text(encoding="utf-8"))
    failures=[]
    for c in m["cases"]:
        p=root/c["path"]
        if not p.is_file():
            failures.append((c["id"],"missing"))
            continue
        if p.stat().st_size != c["bytes"]:
            failures.append((c["id"],"size"))
        if sha256_file(p) != c["sha256"]:
            failures.append((c["id"],"sha256"))
    expected={c["path"] for c in m["cases"]}|{"manifest.json"}
    actual={str(p.relative_to(root)).replace("\\","/") for p in root.rglob("*") if p.is_file()}
    if actual != expected:
        failures.append(("closed_set",{"extra":sorted(actual-expected),"missing":sorted(expected-actual)}))
    if failures:
        print(json.dumps({"status":"FAIL","failures":failures},indent=2))
        return 1
    print(json.dumps({"status":"PASS","cases":len(m["cases"]),"closed_set_files":len(actual)},indent=2))
    return 0

if __name__=="__main__":
    raise SystemExit(main())
