#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)
lg="$root/build/toolchain/lg"
[[ -x "$lg" ]] || "$root/scripts/build-toolchain.sh"
case "${1:-health}" in
health)
  "$lg" -source-paths "$root/src:$root/test" -e \
    '(require (quote glimmer.backend)) (require (quote gtt.state)) (println {:status :ok :runtime :let-go})'
  ;;
*)
  echo "usage: $0 health" >&2
  exit 2
  ;;
esac
