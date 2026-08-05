#!/usr/bin/env bash
set -euo pipefail

output_dir="${1:-${RUNNER_TEMP:-/tmp}}"
mkdir -p "$output_dir"
base64_path="$output_dir/m118-repair.b64"
patch_path="$output_dir/m118-issue65.patch"

hash_file() {
  sha256sum "$1" | sed 's/^\\//' | cut -d' ' -f1
}

expected=(
  2033ac2314d96fcd56a1cad1a0fba87e48d0cb8d039f59a226f331376e701ac7
  e40f5515361b370c809a99b90f1008bb8a74d39cdabf5ae5da40ec9a55293d4f
  dda168c4fedb305bb0ff882aff761ea8ac94c4d4df60c74871fed77c5d46b70e
  52e1e0aa954f4590ddf6f8665d8b8ca1582bb3ab32c2bb5cdb74b609b61de88a
  66d75dd6d825524cf044f986f82db13e2fe1e012d7ccad4029883d411a563ccf
)

: > "$base64_path"
for i in 0 1 2 3 4; do
  src="repairs/issue65/repair.v2.part0${i}"
  clean="$output_dir/repair-part-${i}"
  tr -d '\r\n\t ' < "$src" > "$clean"
  got="$(hash_file "$clean")"
  echo "PART_${i}_SHA256=${got}"
  test "$got" = "${expected[$i]}"
  cat "$clean" >> "$base64_path"
done

test "$(wc -c < "$base64_path" | tr -d ' ')" = "14696"
base64 -d "$base64_path" | gzip -d > "$patch_path"
patch_hash="$(hash_file "$patch_path")"
echo "DECODED_PATCH_SHA256=${patch_hash}"
test "$patch_hash" = "68ad7817a0069e5c17c0ab770146800dc0213bfb9cdc77a7d85bd88c771359cd"

git apply --check "$patch_path"
git apply "$patch_path"
test "$(hash_file internal/runtime/turnorchestrator/orchestrator.go)" = "acefa7452f7115facb3d097055eb00f4efe45c7bababc1ab888d07743bcbc7f4"
test "$(hash_file internal/runtime/turnorchestrator/types.go)" = "8e11bb2945bea42be22a63a5f08c4c2956c204297d694fbf5826c8d8b0e5f9b3"

echo "REPAIR_APPLIED_FROM_EXACT_PARTS=true"
