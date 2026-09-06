# Follow-up: upstream callback-coercion issue

Tracks an issue filed against [nooga/let-go](https://github.com/nooga/let-go) describing
a gap noted in [LET-GO-ISSUES.md](LET-GO-ISSUES.md) under "Generic escaped callback
coercion." This doc is the downstream-side tracking note: what was filed, what it does
and doesn't cover, and what changes here if/when it's fixed.

**Upstream issue:** _link/id to be filled in once filed — see `tmp/ISSUE.md` at the
repo root for the drafted text._

## What the issue reports

`Func.Unbox()` (`pkg/vm/func.go`) is let-go's only existing generic path for coercing a
let-go `fn` into a plain Go callback type (`func()`, `func() bool`, etc. — the shapes
GLib/gotk4 signal and source APIs take). Two bugs, both reproduced against `pkg/api`
alone with no GTK dependency:

1. The `reflect.MakeFunc` proxy always returns exactly one value, so it panics
   (`reflect: wrong return count from function created by MakeFunc`) for any target
   signature with a different output arity — `func()` (zero outputs) is the common
   signal-handler shape, so this panics on the case we actually need.
2. `out, _ := f.Run()` inside the proxy discards the error, so a let-go-side exception
   thrown during a callback invocation is silently dropped instead of reaching the host.

## What it deliberately does not cover

The issue scopes out callback *lifecycle* (retain while a signal is connected, release
on disconnect) as a separate, smaller-audience concern. Today that property holds for
free here because `native/gtkbridge` uses ordinary hand-written Go closures, not
`Unbox`: gotk4 pins the closure (and the `vm.Fn` it captures) only while the signal
stays connected, and drops it on `HandlerDisconnect` (see `disconnect!` in
`examples/gtk-counter/main.go` and `cmd/gtt-letgo-gtk/main.go`), letting normal Go GC
reclaim it. A future generic upstream mechanism isn't guaranteed to preserve that if it
needs its own retention registry to work around bugs 1–2 above.

## Status here

No blocker: `native/gtkbridge` already routes every signal/source connection through
its own `Invoke` helper (`native/gtkbridge/bridge.go`) instead of `Unbox`, precisely to
sidestep both bugs. This doc exists so we don't lose track of the upstream report, not
because anything downstream is currently broken.

## If/when the upstream issue is fixed

- Revisit whether `native/gtkbridge`'s hand-written per-signal-shape adapters
  (`ConnectClicked`, `ConnectChanged`, `ConnectToggled`, `IdleAdd`, `TimeoutAdd`) can be
  collapsed onto a generic `Unbox`-based coercion instead, per-shape adapter code being
  the maintenance cost called out in `LET-GO-ISSUES.md`.
- Before relying on any generic replacement, rerun the existing 1,000-click and
  repeated-lifecycle stress tests (`examples/gtk-counter`) against it — those are the
  regression workload for exactly the retain/release property the upstream issue
  doesn't promise to preserve.
- Keep `native/gtkbridge`'s explicit error-to-stderr + counter behavior
  (`CallbackErrorCount`/`ResetCallbackErrorCount`) as the acceptance bar for whatever
  error-propagation upstream ships, since that's what our error/thread/lifetime policy
  (`examples/gtk-counter/README.md`) is written against.
