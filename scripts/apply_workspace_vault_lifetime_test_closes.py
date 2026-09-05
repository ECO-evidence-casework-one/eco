from pathlib import Path


def replace_once(text: str, old: str, new: str, label: str) -> str:
    count = text.count(old)
    if count != 1:
        raise SystemExit(f"{label}: expected exactly one anchor, found {count}")
    return text.replace(old, new, 1)


local_path = Path("internal/eco/local_tools_test.go")
local = local_path.read_text(encoding="utf-8")
local = replace_once(
    local,
    '''\tif strings.Contains(string(raw), toolPath) || strings.Contains(string(raw), "5.5.1-test") || strings.Contains(string(raw), registered.SHA256) {\n\t\tt.Fatal("encrypted workspace leaks local tool registration plaintext")\n\t}\n\n\treopened, err := OpenVault(vaultRoot)\n''',
    '''\tif strings.Contains(string(raw), toolPath) || strings.Contains(string(raw), "5.5.1-test") || strings.Contains(string(raw), registered.SHA256) {\n\t\tt.Fatal("encrypted workspace leaks local tool registration plaintext")\n\t}\n\tif err := v.Close(); err != nil {\n\t\tt.Fatal(err)\n\t}\n\n\treopened, err := OpenVault(vaultRoot)\n''',
    "local-tool reopen close",
)
local_path.write_text(local, encoding="utf-8")

occurrence_path = Path("internal/eco/occurrence_test.go")
occurrence = occurrence_path.read_text(encoding="utf-8")
occurrence = replace_once(
    occurrence,
    '''\tif _, duplicate, err := v.ImportFile(pathB, nil); err != nil || !duplicate {\n\t\tt.Fatalf("duplicate import err=%v duplicate=%v", err, duplicate)\n\t}\n\n\treopened, err := OpenVault(vaultRoot)\n''',
    '''\tif _, duplicate, err := v.ImportFile(pathB, nil); err != nil || !duplicate {\n\t\tt.Fatalf("duplicate import err=%v duplicate=%v", err, duplicate)\n\t}\n\tif err := v.Close(); err != nil {\n\t\tt.Fatal(err)\n\t}\n\n\treopened, err := OpenVault(vaultRoot)\n''',
    "occurrence reopen close",
)
occurrence_path.write_text(occurrence, encoding="utf-8")

preservation_path = Path("internal/eco/preservation_test.go")
preservation = preservation_path.read_text(encoding="utf-8")
needle = '''\n\treopened, err := OpenVault(root)\n'''
count = preservation.count(needle)
if count != 4:
    raise SystemExit(f"preservation restart closes: expected 4 reopen anchors, found {count}")
preservation = preservation.replace(
    needle,
    '''\n\tif err := vault.Close(); err != nil {\n\t\tt.Fatal(err)\n\t}\n\n\treopened, err := OpenVault(root)\n''',
)
preservation_path.write_text(preservation, encoding="utf-8")
