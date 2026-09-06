#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)
mkdir -p "$root/build"
"$root/scripts/embed-sources.py"
(cd "$root" && GOMEMLIMIT=1500MiB GOGC=20 go build -p=1 -trimpath -ldflags='-s -w' -o build/gtt-letgo-gtk ./cmd/gtt-letgo-gtk)
printf 'built %s (%s bytes)\n' "$root/build/gtt-letgo-gtk" "$(stat -c %s "$root/build/gtt-letgo-gtk")"
