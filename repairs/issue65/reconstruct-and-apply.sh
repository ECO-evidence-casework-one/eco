#!/usr/bin/env bash
set -euo pipefail

output_dir="${1:-${RUNNER_TEMP:-/tmp}}"
mkdir -p "$output_dir"
base64_path="$output_dir/m118-repair-v3.b64"
patch_path="$output_dir/m118-issue65-v3.patch"

hash_file() {
  sha256sum "$1" | sed 's/^\\//' | cut -d' ' -f1
}

expected=(
  6fb88aa8da85c304efb08419114e232ecbea1e558fc50a066e2f9d5e7167a778
  414ae4aa2de93d267a61594a0b032644f09d4f03b021e3dfc1ef3e794a73a3de
  fa41393c7dc898573e2a4cfd0670aa781c63ce683ff8ef034539111c594c4c1f
  c1749f783118764edbe76be425803910c04982e3c55a10d2b1ba1244205c1e67
  2b062a1a97c53b0257581b65f329f042f71a5ad90fed0fb1daf3cefbe2f50ec4
)

: > "$base64_path"
for i in 0 1 2 3 4; do
  src="repairs/issue65/repair.v3.part0${i}"
  clean="$output_dir/repair-v3-part-${i}"
  tr -d '\r\n\t ' < "$src" > "$clean"
  got="$(hash_file "$clean")"
  echo "PART_${i}_SHA256=${got}"
  test "$got" = "${expected[$i]}"
  cat "$clean" >> "$base64_path"
done

test "$(wc -c < "$base64_path" | tr -d ' ')" = "16716"
base64 -d "$base64_path" | gzip -d > "$patch_path"
patch_hash="$(hash_file "$patch_path")"
echo "DECODED_PATCH_SHA256=${patch_hash}"
test "$patch_hash" = "48dba9b09150568d85c1acf0b0b3b05d77e92e32733c8140590236fb360fef2e"

git apply --check "$patch_path"
git apply "$patch_path"
test "$(hash_file internal/runtime/turnorchestrator/orchestrator.go)" = "1b563040a3571a2d9185f888c97c5d0709ac5dcd2f4307a21679808eb0af57de"
test "$(hash_file internal/runtime/turnorchestrator/types.go)" = "20c68bdb7a0222df7c6e3f20302bc9056666d5f25b4f7cf0e6f5f2bbb12d473d"
test "$(hash_file internal/runtime/turnorchestrator/process_lease.go)" = "7fa7246cccf66d2c2eb2b956f0016080e11e49d75dc7be33525bc0e9402a6b5b"

echo "REPAIR_V3_APPLIED_FROM_EXACT_PARTS=true"
