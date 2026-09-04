#!/usr/bin/env python3
import subprocess, sys
from pathlib import Path
here = Path(__file__).resolve().parent
mode = sys.argv[1] if len(sys.argv) > 1 else "validate"
raise SystemExit(subprocess.call([sys.executable, str(here/"aeib"/"generate.py"), mode, "--out", str(here/"corpus")]))
