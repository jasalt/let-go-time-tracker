# let-go GTK counter (Gate A)

This is the Milestone 1 host-boundary proof. The authored behavior and state live
in [`app.lg`](app.lg): it creates the widget hierarchy, owns the count/timer
atoms, handles clicks, and decides label content. [`main.go`](main.go) is a
mechanical gotk4 host.

## Run and test

```sh
go build -o build/gtk-counter ./examples/gtk-counter
GDK_BACKEND=x11 ./build/gtk-counter

go test ./examples/gtk-counter
examples/gtk-counter/smoke-test.sh
```

The smoke test uses Xvfb and xdotool. It proves the periodic callback advances,
queues a render through `glib.IdleAdd`, delivers at least 1,000 button clicks,
checks that no callback errors were logged, and launches/closes three independent
application lifecycles.

## Interop finding

Current let-go can generate wrappers for ordinary Go constructors, values, and
instance methods, and its reflective embed API can round-trip gotk4 pointers.
That is suitable for broad generated bindings later. Directly passing a let-go
function where gotk4 requires a concrete Go callback such as `func()` is not a
normal value conversion, however. gotk4 signal registration also retains the Go
closure after the call returns. The proof therefore uses explicit wrappers that
accept `vm.Fn`, retain it through a Go closure, invoke it, and report invocation
errors.

The host functions in this spike intentionally include a few typed widget
operations only to prove pointer/method interop before generated bindings are
introduced. They are not the final bridge API. Milestone 5 must remove ordinary
widget operations from the handwritten seam and retain only callback conversion,
main-loop scheduling, lifecycle operations that generated interop cannot express,
and documented conversions.

## Error, thread, and lifetime policy

- Every callback passes through `invoke`; errors are written to stderr with the
  callback kind and increment an observable process-local counter.
- A failed repeating timer callback returns false, removing that GLib source.
- `runtime.LockOSThread` pins GTK startup/main-loop ownership.
- Click and timeout callbacks run on GTK's main loop. `on-idle!` proves a
  one-shot scheduling path for work originating elsewhere.
- gotk4 signal/source closures retain the `vm.Fn`; the 1,000-click test proves
  callbacks remain callable after initial evaluation and garbage-collection
  opportunities.
- Closing each window lets `GtkApplication.Run` return cleanly; three fresh
  lifecycle cycles pass under Xvfb.

## Gate A decision

**Go.** GTK reliably invokes let-go functions without application state or UI
policy in Go. gotk4 pointers round-trip through the embed boundary, main-loop and
timer callbacks work, callback failures have explicit semantics, and stress and
lifecycle smoke tests pass. The required native seam is small enough to proceed
to the portable Glimmer backend. The principal optimization is to generate
ordinary gotk4 constructor/method bindings and reserve handwritten Go for
escaped callback coercion and scheduling. The counter now mounts through
`glimmer.core` and `glimmer-gtk.core`, rather than constructing its tree
imperatively.

## Gate C decision

**Go.** The actual Xvfb counter uses the GTK backend's Clojure `case` dispatch,
property diffing, child semantics, callback handler replacement, and scheduler.
The same 1,000-click and lifecycle stress passes. Handwritten `gtkbridge`
contains no widget hierarchy or rendering decisions; the embed host exposes
mechanical gotk4 methods pending generated bindings.
