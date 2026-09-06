#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)
lg="$root/build/toolchain/lg"
[[ -x "$lg" ]] || "$root/scripts/build-toolchain.sh"
files=(
  test/glimmer/backend_test.lg
  test/glimmer/ratom_test.lg
  test/glimmer/core_test.lg
  test/glimmer_gtk/core_test.lg
  test/gtt/view_test.lg
  test/gtt/actions_test.lg
  test/gtt/provider_test.lg
  test/gtt/config_test.lg
  test/gtt/application_test.lg
)
for file in "${files[@]}"; do
  echo ">> $file"
  output=$("$lg" -source-paths "$root/src:$root/test" "$root/$file")
  printf '%s\n' "$output"
  summary=$(printf '%s\n' "$output" | tail -n 1)
  if [[ ! "$summary" =~ Fail:\ 0\ Error:\ 0$ ]]; then
    echo "test failure in $file" >&2
    exit 1
  fi
done
