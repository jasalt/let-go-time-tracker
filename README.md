# lgtt

A let-go-first Glimmer-style GTK4 time tracker, using a small mechanical Go /
gotk4 host boundary.

## Build and test

```sh
scripts/build-toolchain.sh
scripts/test.sh
scripts/test-native.sh
examples/gtk-counter/smoke-test.sh
scripts/build-aot.sh
```

The release command creates `build/gtt-letgo-gtk`, a stripped standalone binary
that embeds application sources and does not require `lg` or a source checkout
at runtime. GTK4 remains a dynamic system dependency.

```sh
./build/gtt-letgo-gtk
LGTT_NREPL_PORT=7888 ./build/gtt-letgo-gtk
```

Configure Clockify or Kimai using the compatible
`${XDG_CONFIG_HOME:-~/.config}/gtt/config.json` format described by the Go
reference application. For deterministic development, a configuration
containing `{"provider":"mock"}` selects the built-in mock provider.

## Architecture

- `src/glimmer/` — GTK-free backend contract, reactive cells, and reconciler;
- `src/glimmer_gtk/` — GTK widget/property/container semantics in let-go;
- `src/gtt/` — serializable state, components, asynchronous actions,
  configuration, and provider adapters;
- `native/gtkbridge/` — retained callback and GLib scheduling conversion only;
- `cmd/gtt-letgo-gtk/` — mechanical gotk4 composition host and generated source
  embedding;
- `examples/gtk-counter/` — Gate A/Gate C callback and backend stress proof.

See:

- [`docs/letgo-baseline.md`](docs/letgo-baseline.md)
- [`docs/letgo-architecture.md`](docs/letgo-architecture.md)
- [`docs/letgo-development.md`](docs/letgo-development.md)
- [`docs/letgo-benchmarks.md`](docs/letgo-benchmarks.md)
- [`docs/LET-GO-ISSUES.md`](docs/LET-GO-ISSUES.md)

No real provider credentials or reset credits are used by automated tests.

## Coding agent sessions

- From PLAN.md to c7692a8 https://pi.dev/session/#ebe762f3695752c2afe8e9efe645aeb8
