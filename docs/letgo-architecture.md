# let-go architecture decisions

## Gate A — callback feasibility

Proceed. The independent GTK counter proves that retained gotk4 closures can
invoke `vm.Fn` values, callback failures can be surfaced, and GLib idle/timeout
sources safely schedule let-go behavior on GTK's main loop. See
[`../examples/gtk-counter/README.md`](../examples/gtk-counter/README.md).

## Gate B — portable Glimmer viability

Proceed. `glimmer.backend`, `glimmer.ratom`, and `glimmer.core` are GTK-free
let-go namespaces. Custom cells implement let-go's native `IDeref`, so `@cell`
is idiomatic. Because core `reset!` and `swap!` only target built-in atoms, the
small explicit `r/reset!` and `r/swap!` API is retained rather than patching the
runtime.

The headless tests prove backend dispatch and scheduling, cursor/reaction watch
cleanup, stateful function components, positional replacement, keyed reorder,
and coalesced reactive rendering. No runtime modifications were required.

## Dependency and thread invariants

```text
gtt.view -> gtt.state/actions -> glimmer.core -> glimmer.ratom
                                  |
                                  v
                            glimmer.backend
                                  |
                                  v
                            glimmer-gtk -> gtkbridge -> gotk4
```

- Domain/application state is printable immutable let-go data and contains no
  GTK handles.
- GTK handles stay in reconciler/backend instances.
- Background actions may update reactive state but never call GTK.
- When the GTK loop is active, `glimmer.backend/schedule!` is the sole path from
  a reactive invalidation to reconciliation.
- Handwritten Go is limited to callback conversion, lifecycle/thread mechanics,
  and conversions that current generated interop cannot express.

## Gate D — application viability

**Go.** Provider-neutral validation/orchestration, duration formatting, secure
configuration loading with legacy fallback, and Clockify/Kimai HTTP request and
mapping code compile and run as let-go. Provider mapping, configuration
roundtrip, input validation, and action transition tests pass without importing
the Go `application`, `domain`, or adapter packages. Ordinary immutable maps and
a map-of-functions provider contract are simpler than reproducing Go
interfaces. The only host dependencies are let-go's built-in `http`, `json`,
`io`, `os`, and `syscall` namespaces.
