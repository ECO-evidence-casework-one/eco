#!/usr/bin/env python3
from __future__ import annotations
import argparse, base64, hashlib, io, json, random, shutil, tarfile, warnings, zipfile
from pathlib import Path

FIXED_DT = (2026, 1, 1, 0, 0, 0)
VERSION = "0.1.0"
MAX_FIXTURE_BYTES = 2 * 1024 * 1024

def sha256(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()

def write(path: Path, data: bytes) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_bytes(data)

def zip_bytes(entries, compression=zipfile.ZIP_DEFLATED):
    bio = io.BytesIO()
    with warnings.catch_warnings():
        warnings.simplefilter("ignore", UserWarning)
        with zipfile.ZipFile(bio, "w") as z:
            for name, data in entries:
                info = zipfile.ZipInfo(name, FIXED_DT)
                info.compress_type = compression
                info.create_system = 3
                info.external_attr = 0o100644 << 16
                z.writestr(info, data)
    return bio.getvalue()

def corrupt_first_zip_payload(data: bytes) -> bytes:
    out = bytearray(data)
    i = out.find(b"PK\x03\x04")
    if i < 0:
        return bytes(out)
    name_len = int.from_bytes(out[i+26:i+28], "little")
    extra_len = int.from_bytes(out[i+28:i+30], "little")
    comp_size = int.from_bytes(out[i+18:i+22], "little")
    start = i + 30 + name_len + extra_len
    if comp_size > 4 and start + comp_size <= len(out):
        out[start + comp_size // 2] ^= 0x01
    return bytes(out)

def minimal_docx(text: str) -> bytes:
    content_types = b'''<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>'''
    rels = b'''<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>'''
    doc = f'''<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
<w:body><w:p><w:r><w:t>{text}</w:t></w:r></w:p></w:body></w:document>'''.encode()
    return zip_bytes([
        ("[Content_Types].xml", content_types),
        ("_rels/.rels", rels),
        ("word/document.xml", doc),
    ])

def minimal_pdf(text: str) -> bytes:
    stream = f"BT /F1 12 Tf 72 720 Td ({text}) Tj ET".encode("latin-1", "replace")
    return b"".join([
        b"%PDF-1.4\n",
        b"1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj\n",
        b"2 0 obj << /Type /Pages /Kids [3 0 R] /Count 1 >> endobj\n",
        b"3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >> endobj\n",
        b"4 0 obj << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> endobj\n",
        b"5 0 obj << /Length " + str(len(stream)).encode() + b" >> stream\n" + stream + b"\nendstream endobj\n",
        b"trailer << /Root 1 0 R >>\n%%EOF\n",
    ])

def _headers(subject: str) -> bytes:
    return (
        "From: synthetic-sender@example.invalid\r\n"
        "To: synthetic-recipient@example.invalid\r\n"
        f"Subject: {subject}\r\n"
        "Date: Thu, 01 Jan 2026 00:00:00 +0000\r\n"
        "MIME-Version: 1.0\r\n"
    ).encode("ascii")

def eml_bytes(subject: str, body: str, attachments=()):
    if not attachments:
        return _headers(subject) + b"Content-Type: text/plain; charset=utf-8\r\n\r\n" + body.encode("utf-8") + b"\r\n"
    boundary = "AEIB-" + hashlib.sha256(subject.encode("utf-8")).hexdigest()[:16]
    out = bytearray()
    out += _headers(subject)
    out += f'Content-Type: multipart/mixed; boundary="{boundary}"\r\n\r\n'.encode("ascii")
    out += f"--{boundary}\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n{body}\r\n".encode("utf-8")
    for name, data, maintype, subtype in attachments:
        out += f"--{boundary}\r\n".encode("ascii")
        if maintype == "message" and subtype == "rfc822":
            out += b"Content-Type: message/rfc822\r\n"
            out += f'Content-Disposition: attachment; filename="{name}"\r\n\r\n'.encode("utf-8")
            out += data + b"\r\n"
        else:
            out += f"Content-Type: {maintype}/{subtype}\r\n".encode("ascii")
            out += b"Content-Transfer-Encoding: base64\r\n"
            out += f'Content-Disposition: attachment; filename="{name}"\r\n\r\n'.encode("utf-8")
            encoded = base64.b64encode(data)
            for i in range(0, len(encoded), 76):
                out += encoded[i:i+76] + b"\r\n"
    out += f"--{boundary}--\r\n".encode("ascii")
    return bytes(out)

def tar_bytes():
    bio = io.BytesIO()
    with tarfile.open(fileobj=bio, mode="w") as tf:
        payload = b"safe payload\n"
        info = tarfile.TarInfo("normal.txt")
        info.size = len(payload); info.mtime = 0
        tf.addfile(info, io.BytesIO(payload))
        traversal = b"must never be written outside staging\n"
        info = tarfile.TarInfo("../AEIB_ESCAPE_SENTINEL.txt")
        info.size = len(traversal); info.mtime = 0
        tf.addfile(info, io.BytesIO(traversal))
        link = tarfile.TarInfo("symlink-outside")
        link.type = tarfile.SYMTYPE; link.linkname = "../outside"; link.mtime = 0
        tf.addfile(link)
    return bio.getvalue()

def build(out: Path, seed: int = 20260821):
    if out.exists():
        shutil.rmtree(out)
    out.mkdir(parents=True)
    rng = random.Random(seed)
    fixtures = []

    def add(fid, rel, data, family, expected, note=""):
        if len(data) > MAX_FIXTURE_BYTES:
            raise RuntimeError(f"{fid} exceeds safety cap")
        write(out / rel, data)
        fixtures.append({
            "fixture_id": fid, "path": rel.replace("\\", "/"), "family": family,
            "expected_secure_handling": expected, "note": note,
            "bytes": len(data), "sha256": sha256(data),
        })

    add("TXT-UTF8", "text/utf8.txt", "AEIB UTF-8 café — synthetic\n".encode(), "text", "accepted")
    add("TXT-CP1252", "text/cp1252.txt", "AEIB café – cp1252\r\n".encode("cp1252"), "text", "accepted")
    add("TXT-UTF16LE-NOBOM", "text/utf16le_no_bom.txt", "AEIB UTF16 LE no BOM".encode("utf-16le"), "text", "accepted-or-explicit-encoding-uncertain")
    add("TXT-EXTENSIONLESS", "text/extensionless", b"extensionless synthetic content\n", "text", "accepted-or-unsupported")
    add("TXT-INVALID-UTF8", "text/invalid_utf8.txt", b"prefix\xed\xa0\x80suffix\n", "text", "accepted-with-safe-replacement-or-rejected")

    add("PATH-PERCENT", "paths/percent%25name.txt", b"percent filename\n", "path", "accounted")
    add("PATH-NEWLINE", "paths/line\nbreak.txt", b"newline filename\n", "path", "accounted-with-safe-display")
    add("PATH-NFC", "paths/café.txt", b"NFC\n", "path", "accounted")
    add("PATH-NFD", "paths/café.txt", b"NFD\n", "path", "accounted-with-distinct-or-normalized-identity")

    docx = minimal_docx("AEIB synthetic DOCX")
    add("OOXML-DOCX-VALID", "office/valid.docx", docx, "office", "accepted")
    add("OOXML-DOCX-TRUNCATED", "office/truncated.docx", docx[:max(40, len(docx)//3)], "office", "corrupt-or-rejected")

    pdf = minimal_pdf("AEIB synthetic PDF")
    add("PDF-VALID", "pdf/valid.pdf", pdf, "pdf", "accepted")
    add("PDF-TRUNCATED", "pdf/truncated.pdf", pdf[:-20], "pdf", "corrupt-or-rejected")
    add("PDF-EXT-MISMATCH", "pdf/not_really_pdf.pdf", b"MZ synthetic extension mismatch only; not executable\n", "signature", "signature-mismatch-or-unsupported")

    add("ZIP-NORMAL", "archives/normal.zip", zip_bytes([("a.txt", b"A\n"), ("b.txt", b"B\n")]), "archive", "accepted")
    add("ZIP-DUPLICATE-NAMES", "archives/duplicate_names.zip", zip_bytes([("dup.txt", b"first\n"), ("dup.txt", b"second\n")]), "archive", "accepted-with-distinct-member-identities")
    nested_inner = zip_bytes([("inner.txt", b"nested\n")])
    add("ZIP-NESTED", "archives/nested.zip", zip_bytes([("inner.zip", nested_inner), ("outer.txt", b"outer\n")]), "archive", "accepted-with-bounded-recursion")
    add("ZIP-TRAVERSAL-NAMES", "archives/traversal_names.zip", zip_bytes([
        ("../AEIB_ESCAPE_SENTINEL.txt", b"must never escape staging\n"),
        ("/absolute-like.txt", b"must not use absolute destination\n"),
        ("C:/device-like.txt", b"must not use drive path\n"),
        ("safe.txt", b"safe\n")
    ]), "archive", "reject-or-quarantine-unsafe-members")
    crc_source = zip_bytes([("crc.txt", b"CRC integrity sentinel " * 20)])
    add("ZIP-CORRUPT-CRC", "archives/corrupt_crc.zip", corrupt_first_zip_payload(crc_source), "archive", "corrupt-or-rejected")
    add("ZIP-HIGH-RATIO-SAFE", "archives/high_ratio_safe.zip", zip_bytes([("compressible.txt", b"0"*(512*1024))]), "resource", "bounded-by-policy", "512 KiB uncompressed safety-capped stress fixture")
    add("TAR-TRAVERSAL-SYMLINK", "archives/traversal_symlink.tar", tar_bytes(), "archive", "reject-or-quarantine-unsafe-members")

    add("EML-NORMAL", "email/normal.eml", eml_bytes("AEIB normal", "synthetic body"), "email", "accepted")
    add("EML-DUPLICATE-ATTACHMENT-NAMES", "email/duplicate_attachments.eml", eml_bytes("AEIB duplicate attachments", "body", [
        ("same.txt", b"one", "text", "plain"), ("same.txt", b"two", "text", "plain"),
    ]), "email", "accepted-with-distinct-attachment-identities")
    add("EML-NESTED", "email/nested_email.eml", eml_bytes("AEIB nested email", "outer", [
        ("forwarded.eml", eml_bytes("inner", "inner body"), "message", "rfc822")
    ]), "email", "accepted-with-bounded-recursion")
    add("EML-INVALID-BASE64", "email/invalid_base64.eml", (
        b"From: a@example.invalid\r\nTo: b@example.invalid\r\nSubject: bad b64\r\n"
        b"MIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=x\r\n\r\n"
        b"--x\r\nContent-Type: application/octet-stream\r\nContent-Transfer-Encoding: base64\r\n"
        b"Content-Disposition: attachment; filename=\"bad.bin\"\r\n\r\n%%%NOT-BASE64%%%\r\n--x--\r\n"
    ), "email", "malformed-but-accounted")
    add("EML-UNKNOWN-CHARSET", "email/unknown_charset.eml", (
        b"From: a@example.invalid\r\nTo: b@example.invalid\r\nSubject: bad charset\r\n"
        b"Content-Type: text/plain; charset=x-aeib-unknown\r\n\r\nsynthetic body\r\n"
    ), "email", "safe-replacement-or-explicit-unsupported")
    add("EML-INVALID-UTF8-SURROGATE-BYTES", "email/invalid_utf8_surrogate.eml", (
        b"From: a@example.invalid\r\nTo: b@example.invalid\r\nSubject: invalid utf8\r\n"
        b"Content-Type: text/plain; charset=utf-8\r\n\r\nprefix\xed\xa0\x80suffix\r\n"
    ), "email", "malformed-but-accounted")

    mbox = b""
    for i in range(3):
        mbox += f"From synthetic{i}@example.invalid Thu Jan  1 00:00:0{i} 2026\n".encode()
        mbox += eml_bytes(f"AEIB mbox {i}", f"message {i}").replace(b"\r\n", b"\n") + b"\n"
    add("MBOX-THREE", "email/three_messages.mbox", mbox, "email", "accepted")

    add("BIN-GARBAGE", "binary/deterministic_garbage.bin", bytes(rng.randrange(0,256) for _ in range(4096)), "binary", "unsupported-or-accounted")
    add("MUTATION-A", "mutation/source_A.txt", b"AEIB mutation source version A\n", "mutation", "accepted")
    add("MUTATION-B", "mutation/source_B_same_length.txt", b"AEIB mutation source version B\n", "mutation", "accepted-and-distinct-hash")

    manifest = {
        "schema": "aeib-manifest-v0.1", "version": VERSION, "seed": seed,
        "fixture_count": len(fixtures), "max_fixture_bytes": MAX_FIXTURE_BYTES,
        "safety": "synthetic-only-no-execution-no-network",
        "fixtures": sorted(fixtures, key=lambda x: x["fixture_id"]),
    }
    (out/"MANIFEST.json").write_text(json.dumps(manifest, indent=2, ensure_ascii=False)+"\n", encoding="utf-8")
    return manifest

def validate(out: Path):
    manifest = json.loads((out/"MANIFEST.json").read_text(encoding="utf-8"))
    failures = []
    for f in manifest["fixtures"]:
        p = out / f["path"]
        if not p.exists():
            failures.append((f["fixture_id"], "missing")); continue
        data = p.read_bytes()
        if len(data) != f["bytes"]: failures.append((f["fixture_id"], "size"))
        if sha256(data) != f["sha256"]: failures.append((f["fixture_id"], "sha256"))
        if len(data) > manifest["max_fixture_bytes"]: failures.append((f["fixture_id"], "safety-cap"))
    return failures

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("mode", choices=["generate","validate"])
    ap.add_argument("--out", default="corpus")
    ap.add_argument("--seed", type=int, default=20260821)
    args = ap.parse_args()
    out = Path(args.out)
    if args.mode == "generate":
        m = build(out, args.seed)
        print(json.dumps({"version":VERSION,"fixtures":m["fixture_count"],"seed":args.seed}, indent=2))
        return 0
    failures = validate(out)
    print(json.dumps({"failures":failures,"pass":not failures}, indent=2))
    return 1 if failures else 0

if __name__ == "__main__":
    raise SystemExit(main())
