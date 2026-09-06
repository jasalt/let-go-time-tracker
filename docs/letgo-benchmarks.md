# let-go GTK benchmarks (Milestone 12)

Recorded 2026-09-06 in the Fedora 44 Lima VM (4 GiB RAM), using Xvfb at
1280x720 and `GSK_RENDERER=cairo`. These numbers compare architecture builds in
a repeatable software-rendered environment; they are not physical-desktop frame
presentation measurements.

## Method

For each GUI, a fresh empty HOME/XDG config was used. Startup is process launch
to discovery of a visible window wider than 100 px, polling X11 every 5 ms.
Three launches were sampled. Memory came from `/proc/PID/smaps_rollup` after
five seconds. Cumulative CPU at that point is included only as context because
startup dominates it. Processes were force-terminated between samples because
the reference GTK main loop does not consume SIGTERM while idle.

```sh
Xvfb :77 -screen 0 1280x720x24 -nolisten tcp
# launch with DISPLAY=:77 GDK_BACKEND=x11 GSK_RENDERER=cairo
# poll: xdotool search --pid PID; xdotool getwindowgeometry
# memory: /proc/PID/smaps_rollup
```

## GUI results

| Client                           | Binary bytes | Startup samples (ms) | Median | RSS KiB | PSS KiB | Private KiB |
|----------------------------------|-------------:|----------------------|-------:|--------:|--------:|------------:|
| existing Go/gotk4 `gtt-gtk`      |   24,115,992 | 152, 182, 152        |    152 | 130,308 | 110,877 |      95,404 |
| let-go GTK development (symbols) |   65,419,776 | 172, 196, 197        |    196 | 155,844 | 136,020 |     120,504 |
| let-go GTK release (stripped)    |   42,728,496 | 191, 197, 215        |    197 | 155,908 | 136,252 |     120,860 |
| existing Fyne                    |   33,754,368 | 284, 117, 126        |    126 | 169,968 | 152,764 |     138,016 |

Fyne's first sample had shader/window-system warm-up and is retained rather than
discarded. Shared-library state and the constrained VM make small differences
noise; repeat on target hardware before product decisions.

## Running timer

The deterministic mock provider started a running entry through the real action
path. Over a 10-second interval the GLib one-second source produced exactly 10
reactive ticks. `/proc/PID/stat` advanced one CPU tick at `CLK_TCK=100`, or
**0.10% CPU**. At the end: RSS 151,304 KiB, PSS 130,983 KiB, and private memory
115,332 KiB. No callback errors were logged.

## Reactive update benchmark

`test/reactive_benchmark.lg` mounts a label against the headless backend and
performs immediate reactive mutation/reconciliation. Three independent VM runs:

| Updates | Run 1 ms | Run 2 ms | Run 3 ms | Best ns/update |
| ---: | ---: | ---: | ---: | ---: |
| 100 | 2.654 | 1.488 | 1.327 | 13,274 |
| 1,000 | 28.940 | 14.446 | 14.003 | 14,003 |
| 10,000 | 280.000 | 127.505 | 129.694 | 12,751 |

The linear scaling rules out pathological superlinear reconciliation in this
single-node case. The first run reflects VM/cache warm-up. Coalescing is tested
separately when an event loop is active.

## Interaction and lifetime

- The Glimmer GTK counter delivered at least 1,000 synthetic clicks with no
  callback error and survived three complete launch/close cycles.
- The actual tracker rendered loading, unconfigured error, and deterministic
  mock-provider snapshots under Xvfb.
- A mock running timer scheduled ten one-second reconciliations in ten seconds.
- nREPL mutation and component reload reconciled in the existing window.

## Gate E — release viability

**Proceed, with a footprint caveat.** Startup remains under 220 ms in all let-go
samples and running CPU is negligible. The release adds about 25 MiB PSS and
18.6 MiB on disk over the current Go/gotk4 client, but remains below Fyne's
measured PSS. This is acceptable for the experiment and dramatically below the
multi-process Electron baseline, while preserving a Clojure-first application.
The largest optimization opportunity is build/link reachability: the embedded
runtime and nREPL are both present in the release. A future release tag that
omits nREPL and unused runtime namespaces should reduce binary and memory
without changing application architecture.

## Final native seam classification

| Function group | Class | Disposition |
| --- | --- | --- |
| `gtkbridge.Invoke` and callback error counter | B: generic Go callback limitation | Keep until let-go supplies retained callback coercion and error policy |
| clicked/changed/toggled adapters | B: generic Go callback limitation | Keep; signal handles permit lifecycle cleanup |
| idle/timeout adapters | B: generic Go callback limitation | Keep; they encode one-shot/repeating error semantics |
| application startup and OS-thread lock | A: permanently GTK-specific | Keep in the host composition root, not the bridge |
| ordinary widget constructors/methods exposed by host defs | C: generated interop limitation | Replace with generated bindings when their build integration is stable |
| duplicate callback invocation in demo/product hosts | D: accidental complexity | Removed; both now use `gtkbridge.Invoke` |
| runtime source-path loading | D: accidental complexity | Removed from release via generated source embedding |

The reviewed handwritten bridge is under 60 lines and has no application state,
provider logic, widget tree, properties, or reconciliation decisions.
