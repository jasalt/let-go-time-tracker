#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "$0")/.." && pwd)
letgo=$(cd "$root/../tmp/let-go" && pwd)
patch="$root/patches/let-go-compiler-context.patch"
if ! grep -q 'func (l \*LetGo) CompilerContext' "$letgo/pkg/api/api.go"; then
  git -C "$letgo" apply "$patch"
fi
mkdir -p "$root/build/toolchain"
(cd "$letgo" && go build -o "$root/build/toolchain/lg" .)
printf 'built %s from %s\n' "$root/build/toolchain/lg" "$(git -C "$letgo" rev-parse HEAD)"
