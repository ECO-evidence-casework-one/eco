#!/usr/bin/env python3
"""
AEIB v0.1 — Adversarial Evidence Ingestion Benchmark
Safe deterministic synthetic corpus generator.

No live malware. No credentials. No personal data. No third-party exploit payloads.
Python standard library only.
"""
from __future__ import annotations
from pathlib import Path
from email.message import EmailMessage
from email.policy import SMTP
import argparse, hashlib, io, json, random, shutil, tarfile, zipfile, warnings

VERSION = "0.1.0"
DEFAULT_SEED = 20260821

def sha256_bytes(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()

def write(path: Path, data: bytes):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_bytes(data)

def deterministic_zip(entries):
    out = io.BytesIO()
    with warnings.catch_warnings():
        warnings.simplefilter("ignore", UserWarning)
        with zipfile.ZipFile(out, "w", compression=zipfile.ZIP_DEFLATED) as z:
            for name, data in entries:
                info = zipfile.ZipInfo(name)
                info.date_time = (1980, 1, 1, 0, 0, 0)
                info.compress_type = zipfile.ZIP_DEFLATED
                info.external_attr = 0o600 << 16
                z.writestr(info, data)
    return out.getvalue()

def deterministic_tar(entries):
    out = io.BytesIO()
    with tarfile.open(fileobj=out, mode="w") as t:
        for name, data, kind in entries:
            ti = tarfile.TarInfo(name)
            ti.mtime = 0
            ti.uid = ti.gid = 0
            ti.uname = ti.gname = ""
            if kind == "file":
                ti.size = len(data)
                t.addfile(ti, io.BytesIO(data))
            elif kind == "symlink":
                ti.type = tarfile.SYMTYPE
                ti.linkname = data.decode("utf-8")
                t.addfile(ti)
    return out.getvalue()

def eml_bytes(subject, body, attachments=None, boundary="AEIB-BOUNDARY"):
    msg = EmailMessage(policy=SMTP)
    msg["From"] = "synthetic-sender@example.invalid"
    msg["To"] = "synthetic-recipient@example.invalid"
    msg["Subject"] = subject
    msg["Message-ID"] = "<aeib-v0-1@example.invalid>"
    msg["Date"] = "Fri, 21 Aug 2026 12:00:00 +0000"
    msg.set_content(body)
    for name, payload, maintype, subtype in attachments or []:
        msg.add_attachment(payload, maintype=maintype, subtype=subtype, filename=name)
    if msg.is_multipart():
        msg.set_boundary(boundary)
    return msg.as_bytes(policy=SMTP)

def make_corpus(dest: Path, seed: int):
    if dest.exists():
        shutil.rmtree(dest)
    dest.mkdir(parents=True)
    rng = random.Random(seed)
    cases = []

    def add(case_id, family, rel, data, expected, notes=""):
        p = dest / rel
        write(p, data)
        cases.append({
            "id": case_id, "family": family, "path": rel.replace("\\","/"),
            "sha256": sha256_bytes(data), "bytes": len(data),
            "expected": expected, "notes": notes
        })

    add("AEIB-TXT-001","encoding","text/utf8.txt",
        "AEIB synthetic UTF-8 — café — Ελληνικά — 日本語\n".encode(),
        "inventory_and_extract")
    add("AEIB-TXT-002","encoding","text/utf16le_no_bom.txt",
        "AEIB UTF16LE without BOM searchable sentinel".encode("utf-16le"),
        "inventory_and_extract_or_explicit_unknown_encoding")
    add("AEIB-TXT-003","encoding","text/windows1252.txt",
        "AEIB café £ “quoted”".encode("cp1252"),
        "inventory_and_extract_or_explicit_unknown_encoding")

    add("AEIB-SIG-001","signature","signature/not_really_pdf.pdf",
        b"AEIB synthetic text with misleading .pdf extension\n",
        "inventory_and_detect_mismatch")

    add("AEIB-ZIP-001","archive","archive/duplicate_names.zip",
        deterministic_zip([("dup.txt",b"first"),("dup.txt",b"second")]),
        "inventory_all_occurrences_without_collision")

    inner = deterministic_zip([("inner.txt",b"nested synthetic evidence")])
    add("AEIB-ZIP-002","archive","archive/nested.zip",
        deterministic_zip([("inner.zip",inner),("root.txt",b"root")]),
        "inventory_nested_with_provenance")

    add("AEIB-ZIP-003","path_safety","archive/traversal_name.zip",
        deterministic_zip([("../outside.txt",b"harmless traversal sentinel")]),
        "reject_or_quarantine_unsafe_member_without_escape")

    add("AEIB-ZIP-004","path_identity","archive/odd_names.zip",
        deterministic_zip([("line\nbreak.txt",b"newline"),("100%25.txt",b"percent")]),
        "inventory_with_unambiguous_occurrence_identity")

    valid = deterministic_zip([("ok.txt",b"ok")])
    add("AEIB-ZIP-005","malformed","archive/truncated.zip",
        valid[:-9], "inventory_as_corrupt_without_crash_or_silent_skip")

    add("AEIB-TAR-001","path_safety","archive/symlink_traversal.tar",
        deterministic_tar([
            ("safe.txt",b"safe","file"),
            ("link",b"../../outside","symlink")
        ]), "inventory_link_metadata_but_never_follow_outside_root")

    dup_eml = eml_bytes("AEIB duplicate attachment names","Synthetic message.",[
        ("same.txt",b"one","text","plain"),("same.txt",b"two","text","plain")
    ], boundary="AEIB-BOUNDARY-DUP")
    add("AEIB-EML-001","email","email/duplicate_attachments.eml",dup_eml,
        "inventory_both_attachments_with_distinct_occurrence_identity")

    zip_payload = deterministic_zip([("inside.txt",b"attached archive")])
    zip_eml = eml_bytes("AEIB ZIP attached","Synthetic archive attachment.",[
        ("evidence.zip",zip_payload,"application","zip")
    ], boundary="AEIB-BOUNDARY-ZIP")
    add("AEIB-EML-002","email","email/zip_attachment.eml",zip_eml,
        "inventory_attachment_and_nested_members")

    forwarded = eml_bytes("forwarded synthetic","inner body", boundary="AEIB-FWD-INNER")
    outer = eml_bytes("outer synthetic","outer body",[
        ("forwarded.eml",forwarded,"message","rfc822")
    ], boundary="AEIB-FWD-OUTER")
    add("AEIB-EML-003","email","email/forwarded_eml.eml",outer,
        "inventory_nested_message_with_provenance")

    malformed = (
        b"From: synthetic@example.invalid\r\nTo: x@example.invalid\r\n"
        b"Subject: malformed synthetic MIME\r\nMIME-Version: 1.0\r\n"
        b"Content-Type: application/octet-stream\r\n"
        b"Content-Transfer-Encoding: base64\r\n\r\n"
        b"%%%NOT-BASE64%%%\r\n"
    )
    add("AEIB-EML-004","malformed","email/malformed_base64.eml",malformed,
        "inventory_as_malformed_without_generic_untracked_failure")

    invalid_header = (
        b"From: synthetic@example.invalid\r\nTo: x@example.invalid\r\n"
        b"Subject: AEIB invalid bytes \xed\xa0\x80\r\n\r\nbody\r\n"
    )
    add("AEIB-EML-005","encoding","email/invalid_header_bytes.eml",invalid_header,
        "sanitize_or_escape_invalid_text_at_database_log_boundaries")

    for ext, cid in [("docx","AEIB-OFFICE-001"),("xlsx","AEIB-OFFICE-002"),("pptx","AEIB-OFFICE-003")]:
        data = deterministic_zip([
            ("[Content_Types].xml",b'<?xml version="1.0"?><Types></Types>'),
            ("broken.xml",b"<root><unterminated>")
        ])
        add(cid,"office",f"office/malformed.{ext}",data,
            "inventory_container_and_report_parse_failure_without_crash")

    garbage = bytes(rng.randrange(0,256) for _ in range(4096))
    add("AEIB-BIN-001","malformed","binary/deterministic_garbage.bin",garbage,
        "inventory_unknown_binary_without_crash")

    bombish = b"A" * (2 * 1024 * 1024)
    add("AEIB-RES-001","resource","resource/high_ratio.zip",
        deterministic_zip([("2mb.txt",bombish)]),
        "enforce_documented_expansion_limits_or_account_safely",
        "Safe 2 MiB payload; deliberately modest, not a destructive zip bomb.")

    add("AEIB-MUT-001","state","mutation/source_A.txt",
        b"VERSION-A-" + b"A"*4096, "detect_if_source_changes_during_read")
    add("AEIB-MUT-002","state","mutation/source_B.txt",
        b"VERSION-B-" + b"B"*4096, "replacement_bytes_for_mutation_test")

    manifest = {
        "schema": 1,
        "benchmark": "AEIB",
        "version": VERSION,
        "seed": seed,
        "safety": {
            "synthetic_only": True,
            "live_malware": False,
            "credentials": False,
            "personal_data": False,
            "third_party_exploit_payloads": False
        },
        "cases": sorted(cases, key=lambda x: x["id"])
    }
    (dest / "manifest.json").write_text(
        json.dumps(manifest, indent=2, sort_keys=True) + "\n",
        encoding="utf-8"
    )
    return manifest

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--output", required=True)
    ap.add_argument("--seed", type=int, default=DEFAULT_SEED)
    args = ap.parse_args()
    manifest = make_corpus(Path(args.output), args.seed)
    print(f"AEIB {VERSION}: generated {len(manifest['cases'])} cases at {args.output}")

if __name__ == "__main__":
    main()
