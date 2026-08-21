#!/usr/bin/env python3
import argparse
import binascii
from pathlib import Path

UU_LINES = [
    "M4F%R(1H'`0`SDK7E\"@$%!@`%`0&`@`\"RNQFY)`(#\"]<!!-<!(/[V5YV````(",
    "M=&5S=\"YZ:7`*`P+2A*ZXL,':`5!+`P04``@`\"`\"47M)8```````````5````",
    "M\"``@`'1E<W0N='AT550-``=XUW%F>-=Q9G'7<69U>`L``00`````!``````+",
    "MR<@L5@\"BM*+\\7(62U.(2O9**$@!02P<(@&&2\"!4````5````4$L!`A0#%``(",
    "M``@`E%[26(!AD@@5````%0````@`(````````````+:!`````'1E<W0N='AT",
    "M550-``=XUW%F>-=Q9G'7<69U>`L``00`````!`````!02P4&``````$``0!6",
    "M````:P``````?V;7$\"0\"`PN5``25`\"\"`89((@```\"&9I;&4N='AT\"@,\"NY$M",
    "BLK#!V@%4:&ES(&ES(&9R;VT@=&5S=\"YT>'0==U91`P4$````",
]


def decode_fixture() -> bytes:
    return b"".join(binascii.a2b_uu(line.encode("ascii")) for line in UU_LINES)


def uu_file(filename: str) -> str:
    return "begin 644 " + filename + "\n" + "\n".join(UU_LINES) + "\n`\nend\n"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("output", type=Path)
    ap.add_argument("--uu", action="store_true")
    args = ap.parse_args()
    args.output.parent.mkdir(parents=True, exist_ok=True)
    if args.uu:
        args.output.write_text(uu_file(args.output.name.removesuffix(".uu")), encoding="ascii")
    else:
        data = decode_fixture()
        assert data.startswith(b"Rar!\x1a\x07\x01\x00")
        assert b"PK\x03\x04" in data and b"PK\x05\x06" in data
        args.output.write_bytes(data)
        print(f"wrote {len(data)} bytes to {args.output}")


if __name__ == "__main__":
    main()
