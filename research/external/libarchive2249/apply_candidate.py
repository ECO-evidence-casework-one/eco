#!/usr/bin/env python3
from pathlib import Path
import sys

root = Path(sys.argv[1])
rar5 = root / "libarchive/archive_read_support_format_rar5.c"
test_c = root / "libarchive/test/test_read_format_rar5.c"
makefile = root / "Makefile.am"
fixture = root / "libarchive/test/test_read_format_rar5_contains_zip.rar.uu"

src = rar5.read_text(encoding="utf-8")
old_sig = '''\tif(!memcmp(h, signature, sizeof(rar5_signature_xor)))\n\t\treturn 30;'''
new_sig = '''\tif(!memcmp(h, signature, sizeof(rar5_signature_xor)))\n\t\treturn 64;'''
if src.count(old_sig) != 1:
    raise SystemExit(f"expected exactly one standard-signature bid site, got {src.count(old_sig)}")
src = src.replace(old_sig, new_sig)

old_order = '''\tif(best_bid > 30)\n\t\treturn -1;\n\n\tmy_bid = bid_standard(a);\n\tif(my_bid > -1) {\n\t\treturn my_bid;\n\t}\n\tmy_bid = bid_sfx(a);'''
new_order = '''\tmy_bid = bid_standard(a);\n\tif(my_bid > -1) {\n\t\treturn my_bid;\n\t}\n\n\t/* Keep the lower-confidence SFX scan at its existing 30-bit bid. */\n\tif(best_bid > 30)\n\t\treturn -1;\n\n\tmy_bid = bid_sfx(a);'''
if src.count(old_order) != 1:
    raise SystemExit(f"expected exactly one rar5_bid ordering block, got {src.count(old_order)}")
src = src.replace(old_order, new_order)
rar5.write_text(src, encoding="utf-8")

marker = "DEFINE_TEST(test_read_format_rar5_contains_zip)"
if marker not in test_c.read_text(encoding="utf-8"):
    with test_c.open("a", encoding="utf-8") as f:
        f.write('''\n\n/* Regression for #2249: an uncompressed ZIP stored inside RAR5 must not\n * cause the seekable ZIP bidder to win over the exact RAR5 signature. */\nDEFINE_TEST(test_read_format_rar5_contains_zip)\n{\n\tPROLOGUE("test_read_format_rar5_contains_zip.rar");\n\n\tassertA(0 == archive_read_next_header(a, &ae));\n\tassertEqualInt(ARCHIVE_FORMAT_RAR_V5, archive_format(a));\n\tassertEqualString("test.zip", archive_entry_pathname(ae));\n\n\tEPILOGUE();\n}\n''')

from make_fixture import uu_file
fixture.write_text(uu_file("test_read_format_rar5_contains_zip.rar"), encoding="ascii")

mf = makefile.read_text(encoding="utf-8")
needle = "\tlibarchive/test/test_read_format_rar5_stored.rar.uu \\\n"
addition = needle + "\tlibarchive/test/test_read_format_rar5_contains_zip.rar.uu \\\n"
if "test_read_format_rar5_contains_zip.rar.uu" not in mf:
    if mf.count(needle) != 1:
        raise SystemExit(f"expected one Makefile RAR5 stored fixture line, got {mf.count(needle)}")
    mf = mf.replace(needle, addition)
    makefile.write_text(mf, encoding="utf-8")

print("candidate patch applied")
