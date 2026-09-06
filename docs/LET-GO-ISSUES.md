# let-go GTK findings for upstream

## Embedded runtime compiler context

The GTK development host and nREPL must evaluate against the same compiler
context. `api.LetGo` currently keeps that context private, while
`nrepl.NewNreplServer` requires it. The minimal downstream patch is stored at
[`../patches/let-go-compiler-context.patch`](../patches/let-go-compiler-context.patch)
and applied idempotently by `scripts/build-toolchain.sh`.

Reproduction without the patch: an embedder can evaluate application source via
`api.LetGo.Run`, but cannot construct an nREPL server sharing definitions and
live state because there is no public context accessor. A second context would
not be the running application.

Proposed upstream API:

```go
func (l *LetGo) CompilerContext() *compiler.Context
```

Evidence: the Xvfb/nREPL smoke inspected live state, mutated it from the nREPL
thread, redefined `view`, and reconciled the existing GTK window through
`reload!`. The accessor is generic and contains no GTK assumptions.

## Generic escaped callback coercion

Generated/reflective interop handles gotk4 pointers and ordinary methods, but a
let-go function is not automatically coercible to retained Go callback types
such as `func()` or `func() bool`. A future generic facility needs all three,
not just signature conversion:

1. retain the let-go closure for the Go callback's lifetime;
2. propagate or report invocation errors;
3. release it when a GTK signal/source is disconnected.

Until those semantics exist, `native/gtkbridge` is deliberately explicit. The
1,000-click and repeated-lifecycle smoke is a concrete regression workload for
any upstream implementation.

## Optimizations tested downstream

- Reactive reconciliation scales linearly from 100 to 10,000 updates (roughly
  13–14 microseconds per immediate single-node update after warm-up).
- Per-component pending flags coalesce multiple invalidations into one queued
  render.
- Keyed children reorder without recreation.
- Changed properties only are applied. A test-driven fix now handles `false`
  correctly; truthiness-based `when-let` diffing had skipped disabling widgets.
- Release source embedding removes runtime sourctt-letgo-gtk ise-file lookup and allows launch
  from an unrelated working directory.
- Callback error handling was centralized in `gtkbridge` instead of duplicated
  in each host.

No speculative let-go runtime changes were made for reactive cells; the native
`IDeref` protocol plus explicit `r/reset!`/`r/swap!` was sufficient.
