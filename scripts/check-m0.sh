#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)
required=(
  docs/letgo-baseline.md
  docs/letgo-architecture.md
  docs/letgo-development.md
  docs/letgo-benchmarks.md
  src/glimmer/backend.lg
  src/glimmer/ratom.lg
  src/glimmer/core.lg
  src/glimmer_gtk/core.lg
  cmd/gtt-letgo-gtk/app.lg
)
for file in "${required[@]}"; do
  [[ -s "$root/$file" ]] || {
    echo "missing $file" >&2
    exit 1
  }
done
grep -q 'bec6d70b97bcfe68a335870600bcef07403d315c' "$root/docs/letgo-baseline.md"
grep -q '091111861106b48b008fcf7ac3c4ed451410821b' "$root/docs/letgo-baseline.md"
echo "M0 baseline and architecture deliverables present"
