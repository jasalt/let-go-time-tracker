# let-go GTK baseline (Milestone 0)

Recorded 2026-09-06 on the Fedora/Lima development sandbox. This document is a
regression oracle for the experiment; it describes the existing Go frontend,
not the desired implementation.

## Revisions

| Component | Revision |
| --- | --- |
| `go-time-tracker` behavior reference (`../gtt`) | `bec6d70b97bcfe68a335870600bcef07403d315c` |
| let-go source/toolchain target (`../tmp/let-go`) | `091111861106b48b008fcf7ac3c4ed451410821b` |
| gotk4 Go module | `github.com/diamondburned/gotk4/pkg v0.4.0`, module sum `h1:6b4G1IRAqyQGG0HrNF/0C2L1KfbRPWffU2BZdQefsgs=` |
| Go | `go1.26.7-X:nodwarf5 linux/amd64` |
| OS | Fedora 44, Linux `7.1.10-200.fc44.x86_64`, amd64 |

The let-go revision above is the exact local checkout used by the project. Do
not substitute an ambient `lg` executable.

## Reproducible build results

The reference checkout was built from a clean `bin/` directory:

```sh
cd ../gtt
make clean all
```

All four targets built successfully. A cold-ish build after module downloads
had been populated took 3:06.06 wall time and reached 546,692 KiB maximum build
RSS. gotk4 emitted generated-C warnings about a declaration of `free`; these did
not fail the build.

| Target     | Command                                   |  Unstripped size |
|------------|-------------------------------------------|-----------------:|
| `gtt`      | `go build -o bin/gtt ./cmd/gtt`           |  9,591,871 bytes |
| `gtt-tui`  | `go build -o bin/gtt-tui ./cmd/gtt-tui`   | 12,113,337 bytes |
| `gtt-fyne` | `go build -o bin/gtt-fyne ./cmd/gtt-fyne` | 33,754,368 bytes |
| `gtt-gtk`  | `go build -o bin/gtt-gtk ./cmd/gtt-gtk`   | 24,115,992 bytes |

Verification on the same tree:

```text
go test ./...       PASS
go vet ./...        PASS
go test -race ./... TIMED OUT after 10 minutes while compiling desktop/cgo packages
```

The ordinary suite covers the Go core, providers, CLI, TUI, and Fyne. There are
no tests in `ui/gtkui`.

The Fyne build additionally required Fedora development packages that were not
initially installed: `mesa-libGL-devel`, `libXcursor-devel`, `libXrandr-devel`,
`libXi-devel`, `libXinerama-devel`, and `libXxf86vm-devel` (plus their resolved
dependencies).

## GTK dependencies

Build requirements documented and verified on Fedora:

```sh
sudo dnf install gtk4-devel gobject-introspection-devel
```

The measured system had GTK 4.22.4 and GObject Introspection 1.86.0. CGO and a C
compiler are required. The executable dynamically links GTK/GLib and their
normal transitive desktop stack, including `libgtk-4`, `libgio-2.0`,
`libgobject-2.0`, `libglib-2.0`, Pango, Cairo, Graphene, GDK Pixbuf, X11, and
Wayland client libraries. Runtime systems therefore need GTK4 and dependencies;
they do not need GTK development headers.

## GTK launch and measurements

The baseline was launched in a fresh, intentionally unconfigured home on Xvfb:

```sh
Xvfb :97 -screen 0 1280x720x24 -nolisten tcp &
HOME=/tmp/lgtt-m0-home \
XDG_CONFIG_HOME=/tmp/lgtt-m0-home/.config \
DISPLAY=:97 GDK_BACKEND=x11 ../gtt/bin/gtt-gtk
```

A script recorded monotonic wall-clock time immediately before process launch
and polled `xdotool search --pid PID --name 'Go Time Tracker'` every 10 ms.
Window discovery took **190 ms**. This is process-to-mapped-window latency under
Xvfb/software rendering, not a frame-presented measurement on physical hardware.
After three seconds idle, `/proc/PID/smaps_rollup` reported:

| Metric        |       Value |
|---------------|------------:|
| RSS           | 224,152 KiB |
| PSS           | 203,711 KiB |
| anonymous PSS | 103,716 KiB |
| file PSS      |  94,111 KiB |
| swap          |       0 KiB |

The high value includes GTK software/Vulkan rendering under Xvfb and should not
be compared blindly with a hardware-composited desktop. The process was still
settling at the short sample (`ps` showed 21.5% cumulative CPU); a running-timer
CPU figure cannot be measured without provider credentials or a mock provider.
Milestone 12 must use repeated, settled measurements and the same compositor for
all compared clients.

Visual smoke testing confirmed an 800x600 window containing description input,
project and task dropdowns, billable checkbox, elapsed and today labels,
Start/Stop/Save/Refresh buttons, status text, and the recent-entry scroll area.
With no configuration, the asynchronous initial load displays:

```text
gtt is not configured; run `gtt configure`
```

The application stayed responsive and closed cleanly. Portal, DRI3, and Vulkan
warnings were expected in the minimal Xvfb session; llvmpipe rendered the UI.

## Current GTK implementation and target inventory

The current implementation is only `ui/gtkui/runner.go` and
`ui/gtkui/view.go`. `Runner` owns `GtkApplication`, starts goroutines for API
work, and marshals completion with `glib.IdleAdd`. `View` owns mutable widget
handles and translates snapshots to imperative widget updates.

| Feature | Current Go implementation | Expected let-go implementation | Regression strategy |
| --- | --- | --- | --- |
| Initial snapshot load | `activate.load` calls `UseCases.Snapshot(ctx, 30)` in a goroutine, then `IdleAdd` calls `SetSnapshot` | `gtt.actions/refresh!` updates explicit loading/request state; reactive root renders result on GTK scheduler | Reducer/action test with deterministic mock provider; GTK smoke verifies loading then loaded state |
| Start | Click reads widget values via `View.Input`, calls `UseCases.Start`, then reloads | Named async `start!` reads serializable form state and refreshes state | Headless action tests for payload, pending, success/error, races; GTK click smoke |
| Stop | Click calls `UseCases.Stop`, then reloads | Named async `stop!` with pending/error state | Headless success/error tests and GTK click smoke |
| Save/update | Button exists and sensitivity changes, but **no click handler is connected** | Named async `save!` calls running-entry update | Action contract test and end-to-end mock-provider edit/save test |
| Project selection | Snapshot fills project model and ID side array; selection always reset to index 0 | Project ID lives in app state; selection updates it, resets task ID, and starts generation-safe task load | Reducer tests and stale-response test; GTK selection smoke |
| Task selection | Model is reset to only “No task”; **tasks are never loaded** | Tasks loaded asynchronously for selected project and rendered from state | Mock-provider task fixture tests and GTK dropdown smoke |
| Recent entries | `SetSnapshot` creates labels from description and elapsed value | Component renders immutable entry maps, later with continue/delete actions | Headless component tree assertions; GTK fixture screenshot/smoke |
| Elapsed display | Set only when a snapshot arrives | Derived from running start and current tick state; one-second scheduler | Fake-clock formatting/reactivity test and timed GTK smoke |
| Today total | Formatted from `snapshot.Today` only when snapshot arrives | Derived/rendered snapshot total plus running progression according to application contract | Fake-clock boundary tests and timed GTK smoke |
| Error display | Snapshot/start/stop error text is assigned directly to status label | Errors are structured/printable app state rendered by an error component | Action failure tests and visible GTK error smoke |
| Refresh | Button invokes the same asynchronous load closure | Named generation-safe `refresh!` action | Out-of-order response test and repeated-click GTK stress |
| Running form population | Running description and billable copied into widgets; project/task are not selected | Render form state initialized from running entry | Headless state transition test and GTK running fixture smoke |
| Button enablement | `SetSnapshot` toggles Start/Stop/Save based on running state | Declarative props derived from `running?` and `pending?` | Component prop assertions for all state combinations |
| Main-thread handoff | Completion callbacks use `glib.IdleAdd` | Glimmer GTK backend `schedule`; background code never touches GTK | Scheduler integration test and thread/stress smoke |

## Defects and incomplete behavior to avoid reproducing

These are observed directly from source and, where applicable, from the
unconfigured Xvfb smoke:

1. **Save is inert:** `view.save` has no `ConnectClicked` call.
2. **Project/task workflow is incomplete:** no selection-change handler exists,
   no `UseCases.Tasks` request is made, and the task model always contains only
   “No task”.
3. **Running project/task selection is lost:** every snapshot selects index 0
   regardless of the running entry.
4. **Recent entries do not update visibly after refresh:** `SetSnapshot` assigns
   a new box to `v.entries`, but the scrolled window retains the original child;
   rows are appended to the detached replacement box.
5. **No live ticking:** elapsed and today labels change only on snapshot loads;
   there is no one-second timer.
6. **Failed initial load leaves misleading controls:** widgets begin sensitive,
   and the error path only changes status text, so Stop and Save remain enabled
   in the unconfigured smoke despite no running timer.
7. **No pending-state protection or stale-response sequencing:** repeated
   refresh/action clicks can overlap, and older loads can overwrite newer ones.
8. **No lifecycle cancellation owned by the view:** goroutines use the runner
   context, but closing the window does not explicitly cancel outstanding work.
9. **Start/stop errors overwrite general status only:** errors are not retained
   as structured state, and a later status update erases them.
10. **Recent-entry actions are absent:** GTK has no continue/delete controls,
    unlike the shared application contract and richer frontends.
11. **GTK has no automated tests:** widget state, callback wiring, and thread
    ownership can regress unnoticed.

## Baseline conclusion

The current GTK target is buildable and its minimal shell launches reliably,
but it is substantially behind the TUI/Fyne behavior. The let-go port should
use the shared application contracts as its behavior oracle and treat the GTK
code primarily as a widget/layout baseline. Gate A should first prove callbacks,
main-loop scheduling, timers, and lifetime handling rather than porting the
current imperative defects.
