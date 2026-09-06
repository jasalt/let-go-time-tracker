#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)
lg="$root/build/toolchain/lg"
[[ -x "$lg" ]] || "$root/scripts/build-toolchain.sh"
mkdir -p "$root/build"
"$lg" -source-paths "$root/src:$root/test" -b "$root/build/lgtt-headless-smoke" "$root/test/aot_smoke.lg"
"$root/build/lgtt-headless-smoke"
