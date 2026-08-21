#!/usr/bin/env python3
from pathlib import Path
import hashlib, json, subprocess, sys, tempfile

ROOT=Path(__file__).resolve().parents[2]
GEN=ROOT/"tools"/"aeib"/"generate.py"
VER=ROOT/"tools"/"aeib"/"verify.py"

def tree_hash(root):
    h=hashlib.sha256()
    for p in sorted(x for x in root.rglob("*") if x.is_file()):
        rel=str(p.relative_to(root)).replace("\\","/")
        h.update(rel.encode()+b"\0"+p.read_bytes()+b"\0")
    return h.hexdigest()

with tempfile.TemporaryDirectory() as td:
    a=Path(td)/"a"; b=Path(td)/"b"; c=Path(td)/"c"
    subprocess.check_call([sys.executable,str(GEN),"--output",str(a),"--seed","20260821"])
    subprocess.check_call([sys.executable,str(GEN),"--output",str(b),"--seed","20260821"])
    subprocess.check_call([sys.executable,str(GEN),"--output",str(c),"--seed","20260822"])
    subprocess.check_call([sys.executable,str(VER),str(a)])
    subprocess.check_call([sys.executable,str(VER),str(b)])
    ha,hb,hc=tree_hash(a),tree_hash(b),tree_hash(c)
    assert ha==hb, (ha,hb)
    assert ha!=hc, "different seed should change at least deterministic garbage case"
    ma=json.loads((a/"manifest.json").read_text())
    assert ma["safety"]["synthetic_only"] is True
    assert ma["safety"]["live_malware"] is False
    assert ma["safety"]["personal_data"] is False
    assert len(ma["cases"]) == 22
    print(json.dumps({
        "status":"PASS",
        "same_seed_tree_sha256":ha,
        "different_seed_tree_sha256":hc,
        "cases":len(ma["cases"])
    },indent=2))
