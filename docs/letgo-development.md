# let-go GTK development

## Toolchain

All project commands use the pinned local let-go source revision recorded in
`docs/letgo-baseline.md`:

```sh
scripts/build-toolchain.sh
```

Do not substitute an ambient `lg` binary.

## Tests

```sh
scripts/test.sh
scripts/test-native.sh
examples/gtk-counter/smoke-test.sh
```

## GTK counter and nREPL

```sh
go build -o build/gtk-counter ./examples/gtk-counter
LGTT_NREPL_PORT=7888 ./build/gtk-counter
brepl -p 7888 '@state'
brepl -p 7888 '(glimmer.ratom/swap! state assoc :count 41)'
```

The state mutation can originate on the nREPL connection goroutine. Reactive
invalidation calls `glimmer.backend/schedule!`, which marshals reconciliation
through GLib idle onto GTK's main loop.

To redefine and reload the root without restarting GTK:

```sh
brepl -p 7888 '(defn view [] [:vbox [:label {:label "Reloaded"}]])'
brepl -p 7888 '(reload!)'
```

The root render thunk resolves the current `view` var on each render, so
`reload!` reconciles the redefined component in place. It schedules rather than
directly touching GTK from the REPL thread.

The embedded runtime starts nREPL only when `LGTT_NREPL_PORT` is set. Port `0`
requests an ephemeral port and writes the selected value to `.nrepl-port`, which
editors can discover. CIDER, Calva, Conjure, or any generic nREPL client can
connect; no editor integration is required by the application.

## Required upstream API seam

The embed host needs the same compiler context for application evaluation and
nREPL evaluation. The pinned let-go checkout therefore adds the minimal
`api.LetGo.CompilerContext()` accessor. It exposes no GTK behavior. This is a
generic embedding/nREPL requirement and is a candidate upstream change; the
remaining application does not patch let-go runtime semantics.
