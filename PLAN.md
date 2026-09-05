# PLAN.md — Glimmer-style GTK4 UI for let-go

## Goal

Reimplement the GTK4 frontend of `jasalt/go-time-tracker` using let-go while keeping authored application and UI code as close to pure Clojure as practical.

The target architecture is:

```text
gtt application/domain code (.lg)
              │
              ▼
reactive application state (.lg)
              │
              ▼
Glimmer-style components / Hiccup (.lg)
              │
              ▼
portable Glimmer reconciler (.lg)
              │
              ▼
GTK backend (.lg)
              │
              ▼
minimal mechanical Go bridge
              │
              ▼
gotk4
              │
              ▼
GTK4 / GLib / GObject
```

The primary design principle is:

> Port Glimmer's Clojure component/reconciler model to let-go, implement GTK widget semantics in Clojure, and use Go only where let-go cannot yet directly satisfy gotk4's callback or host-runtime requirements.

The Go layer must not become a second UI implementation.

---

# Reference projects

Review and continuously compare against:

* Existing application:

  * https://github.com/jasalt/go-time-tracker
  * especially:

    * `ui/gtkui/runner.go`
    * `ui/gtkui/view.go`
    * `application/`
    * `domain/`
    * provider adapters

* let-go:

  * https://github.com/nooga/let-go
  * `examples/`
  * `examples/aot/`
  * Go interop documentation
  * native/AOT tooling
  * nREPL support
  * `IDeref`
  * atoms/watches
  * HTTP and JSON support
  * Go callback interoperability work

* Glimmer portable core:

  * https://github.com/jolt-lang/glimmer
  * especially:

    * `src/glimmer/core.clj`
    * `src/glimmer/backend.clj`
    * `src/glimmer/ratom.clj`
    * headless tests

* GTK4 backend reference:

  * https://github.com/jolt-lang/glimmer-gtk
  * especially:

    * `src/glimmer_gtk/core.clj`
    * `src/glimmer_gtk/widget.clj`
    * `src/glimmer_gtk/ffi.clj`

* Background:

  * https://yogthos.net/posts/2026-08-29-glimmer-ui.html

* gotk4:

  * https://github.com/diamondburned/gotk4

Do not mechanically copy Jolt-specific FFI code. Reuse the architecture and portable Clojure concepts, adapting the host boundary to let-go/Go.

---

# Success criteria

The experiment is successful when all of the following hold.

## Functional

A let-go executable can display a GTK4 time-tracker window supporting at minimum:

* current running timer
* elapsed running time
* today's total
* description input
* project selection
* task/activity selection
* billable toggle
* start timer
* stop timer
* update/save running timer
* refresh
* recent entries

The behavior should match or improve upon the current `gtt-gtk` frontend.

## Architectural

Most application code is `.lg` or portable `.cljc`.

The Go bridge:

* contains no application state;
* contains no Clockify/Kimai logic;
* contains no time-tracker-specific widget hierarchy;
* contains no rendering/reconciliation logic;
* does not decide which widgets should exist;
* does not decide how application state maps to widgets;
* provides only mechanical host operations that cannot reasonably be expressed through current let-go Go interop.

A good Go bridge looks like:

```go
func ConnectClicked(button *gtk.Button, fn vm.Fn)
func ConnectChanged(entry *gtk.Entry, fn vm.Fn)
func IdleAdd(fn vm.Fn)
```

A bad Go bridge looks like:

```go
func BuildTimeTrackerWindow(...)
func RefreshTimerView(...)
func UpdateRunningTimer(...)
```

The latter are explicitly out of scope.

## Reactive

A component may read reactive state using idiomatic Clojure syntax and automatically rerender when relevant state changes.

Desired usage:

```clojure
(defonce app-state
  (r/atom {:loading? true
           :snapshot nil}))

(defn status-view []
  [:label {:label
           (if (:loading? @app-state)
             "Loading..."
             "Ready")}])
```

Updating the state should reconcile only the affected Glimmer component subtree.

## Development workflow

During development:

* the let-go VM remains available;
* nREPL can be used;
* UI state can be inspected from the REPL;
* application state can be changed from the REPL;
* reactive mutations originating outside the GTK thread are marshalled safely to the GTK main loop;
* headless Glimmer tests do not require GTK.

## Release

Produce a standalone native executable using let-go's appropriate native/AOT build path plus gotk4/GTK dependencies.

Document:

* startup time;
* binary size;
* idle RSS/PSS;
* CPU usage while timer is running;
* comparison against existing `gtt-gtk`;
* comparison against `gtt-fyne` where useful.

---

# Non-goals

Do not attempt these before the primary experiment works:

* creating a general-purpose complete gotk4 binding for let-go;
* implementing a general C FFI for let-go;
* replacing gotk4 with raw GTK C bindings;
* implementing every widget supported by GTK4;
* porting all of Jolt Glimmer byte-for-byte;
* supporting Qt, UIKit, Windows native UI, or multiple desktop toolkits;
* implementing a browser renderer;
* making the entire implementation portable to JVM Clojure;
* prematurely optimizing the reconciler;
* upstreaming changes before the experiment establishes a concrete requirement.

Prefer a narrow vertical slice over broad framework work.

---

# Repository organization

Keep reusable framework code clearly separate from application-specific code.

A suitable initial layout is:

```text
letgo/
├── deps.edn
├── src/
│   ├── glimmer/
│   │   ├── backend.lg
│   │   ├── core.lg
│   │   └── ratom.lg
│   │
│   ├── glimmer_gtk/
│   │   ├── core.lg
│   │   ├── widget.lg
│   │   └── props.lg
│   │
│   └── gtt/
│       ├── main.lg
│       ├── app.lg
│       ├── state.lg
│       ├── actions.lg
│       ├── view.lg
│       └── format.lg
│
├── native/
│   └── gtkbridge/
│       ├── bridge.go
│       └── callbacks.go
│
├── test/
│   ├── glimmer/
│   ├── glimmer_gtk/
│   └── gtt/
│
├── scripts/
│   ├── dev
│   ├── test
│   └── build
│
└── README.md
```

This may be adjusted to fit let-go's actual current dependency/AOT conventions after inspecting upstream.

Do not spread let-go-specific code throughout the existing Go packages before the architecture is proven.

---

# Coding principles

## Clojure-first

When logic can reasonably live in `.lg`, put it in `.lg`.

This includes:

* reactive state;
* components;
* Hiccup parsing;
* reconciliation;
* widget property normalization;
* application actions;
* state transitions;
* formatting;
* HTTP API use where practical;
* domain transformations;
* loading states;
* errors;
* event handling.

## Mechanical Go only

Native Go exists only for unavoidable host integration.

Every Go function added should answer:

> Why can this not currently be expressed cleanly through let-go's existing Go interop?

Document the answer when it is not obvious.

## Preserve immutable application data

Prefer immutable maps/vectors/records at application boundaries.

GTK widget handles are host objects and may be opaque/mutable, but they must not leak into domain state.

Good:

```clojure
{:running {:description "Review"
           :billable true
           :elapsed-ms 91234}}
```

Bad:

```clojure
{:running-label some-gtk-label
 :start-button some-gtk-button}
```

## Separate state from effects

A component should primarily describe UI from values.

External effects should be expressed through named actions such as:

```clojure
(actions/start-timer!)
(actions/stop-timer!)
(actions/refresh!)
(actions/select-project! project-id)
```

Do not embed networking directly inside rendering functions.

---

# Milestone 0 — Establish baseline and development branch

## Objective

Capture the current state before changing architecture.

## Tasks

1. Build current:

   * `gtt`
   * `gtt-tui`
   * `gtt-fyne`
   * `gtt-gtk`

2. Verify current GTK behavior manually.

3. Record:

   * startup latency;
   * RSS/PSS after idle;
   * binary size;
   * GTK dependency requirements.

4. Inventory current `ui/gtkui` behavior.

Create a table covering:

```text
feature
current Go implementation
expected let-go implementation
test strategy
```

Include at least:

* initial snapshot load;
* start;
* stop;
* save/update;
* project selection;
* task selection;
* recent entries;
* elapsed display;
* today's total;
* error display;
* refresh.

5. Note existing GTK frontend defects or incomplete behaviors separately rather than reproducing them accidentally.

6. Confirm exact let-go revision being targeted.

7. Confirm exact gotk4 revision.

## Deliverable

```text
docs/letgo-baseline.md
```

## Exit criteria

The current application can be built and its GTK behavior has been recorded sufficiently to identify regressions.

---

# Milestone 1 — Minimal custom let-go + gotk4 host

## Objective

Prove the most important architectural boundary before porting Glimmer.

The required proof is:

> A let-go function can own application behavior while GTK widgets are created and manipulated through gotk4.

## Minimal demo

The `.lg` program should produce:

```text
┌──────────────────────────┐
│ Count: 0                 │
│ [Increment]              │
└──────────────────────────┘
```

Clicking the GTK button should invoke a let-go function that mutates let-go state and updates the label.

Do not implement Glimmer yet.

## Tasks

### 1. Determine minimum gotk4 surface

Start with:

* `gtk.Application`
* `gtk.ApplicationWindow`
* `gtk.Box`
* `gtk.Label`
* `gtk.Button`

### 2. Try direct generated let-go Go interop first

Before writing bridge code, test whether current let-go can expose:

* widget constructors;
* widget instance methods;
* enum values;
* opaque pointers/structs.

Prefer generated interop where it works.

### 3. Implement callback bridge only where required

Likely required:

```go
func ConnectClicked(button *gtk.Button, callback vm.Fn)
```

The Go closure should call the supplied let-go function.

Handle errors explicitly.

Do not silently discard callback failures.

### 4. Prove UI-thread scheduling

Expose the smallest practical mechanism equivalent to:

```text
GLib idle callback
```

Conceptually:

```clojure
(gtk/on-main-thread
  #(reset! state ...))
```

### 5. Prove timer callback

Test a periodic callback suitable for updating an elapsed timer.

### 6. Lifecycle

Verify:

* application launches;
* window closes cleanly;
* callbacks do not disappear because of GC/lifetime problems;
* repeated clicks remain stable;
* callbacks do not leak indefinitely.

## Tests

Automated tests where possible for callback invocation.

Manual smoke test for GTK.

Add a stress test that invokes the callback many times.

## Deliverable

A tiny executable/demo independent of `gtt`.

Suggested:

```text
letgo/examples/gtk-counter/
```

## Exit criteria

A GTK button reliably invokes a let-go function and a let-go-originating change can safely modify GTK state on the GTK main thread.

Do not continue to Glimmer until this works.

---

# Milestone 2 — Port the portable Glimmer backend contract

## Objective

Port `glimmer.backend` with no GTK dependency.

This establishes the boundary between:

```text
reconciler
```

and:

```text
native toolkit
```

## Backend contract

Preserve the conceptual operations from Glimmer:

```clojure
{:name
 :create!
 :apply-props!
 :append-child!
 :remove-child!
 :replace-child!
 :reorder-child!
 :schedule
 :run
 :text->element}
```

Not every operation needs to be mandatory initially, but keep the conceptual seam compatible with Glimmer unless let-go requires a good reason to diverge.

## Tasks

1. Port backend registration.

2. Port operation dispatch.

3. Port loop-running state.

4. Support a headless mock backend.

5. Make backend operations testable without GTK.

## Tests

Test:

* registration;
* missing backend;
* missing operation;
* create dispatch;
* property dispatch;
* append/remove/replace;
* scheduler fallback;
* text conversion.

## Exit criteria

`glimmer.backend` runs entirely under let-go and all tests are headless.

---

# Milestone 3 — Implement let-go-native reactive cells

## Objective

Implement the minimum Reagent/Glimmer-style reactive model required by components.

Do not simply copy Jolt's implementation where let-go provides a cleaner primitive.

## Reactive primitives

Target:

```clojure
(r/atom value)
(r/cursor state [:path])
(r/reaction ...)
```

And idiomatic dereference:

```clojure
@state
```

## Use let-go `IDeref`

Investigate and use let-go's `IDeref` protocol so custom reactive cells participate naturally in `@`.

Avoid globally replacing `clojure.core/deref` unless genuinely necessary.

## Reactive protocol

Provide something conceptually equivalent to:

```clojure
(defprotocol IReactiveCell
  (-value [this])
  (-reset! [this new-value])
  (-add-watch! [this watcher])
  (-remove-watch! [this watcher])
  (-notify-watches! [this]))
```

Adapt names to let-go conventions where useful.

## Important questions

Determine whether:

```clojure
reset!
swap!
```

can be cleanly supported for custom reactive values.

If not, prefer initially:

```clojure
(r/reset! x value)
(r/swap! x f)
```

rather than adding a large runtime patch.

Document any resulting ergonomic difference.

## Dependency tracking

Implement:

```text
component render starts
       ↓
bind current watcher
       ↓
component dereferences reactive cells
       ↓
dependencies register watcher
       ↓
render completes
```

When a dependency changes:

```text
reactive cell changes
       ↓
watcher invoked
       ↓
component scheduled for reconciliation
```

## Watch lifecycle

Prevent stale component watches.

A component that is removed must not keep rerendering against destroyed widgets.

Test mounting/unmounting the same component repeatedly.

## Reactions

Derived state must:

* recompute when dependencies change;
* notify only when value meaningfully changes;
* avoid accumulating duplicate watchers.

## Tests

Port/adapt as much of Glimmer's reactive test suite as practical.

Include:

* basic atom deref;
* reset;
* swap;
* cursor read/write;
* nested cursor;
* reaction;
* dependency change;
* unchanged value does not trigger unnecessary work;
* watch removal;
* component disposal;
* repeated recomputation does not leak watches.

## Exit criteria

Reactive behavior is proven headlessly and uses idiomatic `@`.

---

# Milestone 4 — Port the Glimmer component/reconciler core

## Objective

Render Hiccup-like component values against the headless backend.

GTK is still not involved.

## Supported element forms

Support:

```clojure
[:label {:label "Hello"}]
```

children:

```clojure
[:vbox
 [:label {:label "A"}]
 [:label {:label "B"}]]
```

components:

```clojure
(defn greeting [name]
  [:label {:label (str "Hello " name)}])

[greeting "Alice"]
```

conditional children:

```clojure
(when error
  [:label {:label error}])
```

text:

```clojure
[:box "hello"]
```

## Component lifecycle

Support both conceptual Glimmer/Reagent shapes.

### Form 1

```clojure
(defn component [props]
  [:label ...])
```

### Form 2

State created once:

```clojure
(defn counter []
  (let [n (r/atom 0)]
    (fn []
      [:button
       {:label (str @n)
        :on-click #(r/swap! n inc)}])))
```

If Form 2 substantially complicates the initial port, implement Form 1 first, but retain it as a required milestone before calling the core complete.

## Reconciliation

Implement at minimum:

* create;
* reuse same-tag element;
* apply new props;
* append child;
* remove child;
* replace child;
* recursive child reconciliation.

Then add keyed reconciliation or child reordering only after the simple positional case works.

## Scheduling

When the backend event loop is active:

```text
reactive mutation
→ component watcher
→ backend schedule
→ reconcile on UI thread
```

When headless:

```text
reactive mutation
→ reconcile immediately
```

## Coalescing

Multiple state changes before scheduled reconciliation should ideally produce one render rather than one render per mutation.

Implement a per-component pending flag similar to Glimmer.

## Tests

Use a headless backend whose widgets are ordinary Clojure data.

Test exact operation sequences.

Examples:

```text
create label
update label prop
replace label with button
append child
remove child
reorder keyed child
```

Test render counts so dependency tracking can be verified.

## Exit criteria

A non-GTK test can render a reactive component tree and observe correct incremental reconciliation.

---

# Milestone 5 — Define the minimal GTK bridge API

## Objective

Formalize the native boundary now that the reconciler requirements are known.

Do not expose arbitrary gotk4 APIs just because they exist.

## Bridge categories

The bridge may contain only these categories.

### Application lifecycle

Potential operations:

```text
run application
create application window
present window
quit
```

### Callback conversion

Examples:

```text
clicked
changed
toggled
selection changed
activate
timeout
idle
```

### UI-thread scheduling

Provide a safe one-shot scheduler.

### Missing interop conversions

Only add wrappers for Go API shapes that current let-go Go interop cannot express correctly.

## Prefer generated bindings

For ordinary methods such as:

```text
SetLabel
SetSensitive
Append
Remove
SetChild
SetText
Text
Active
```

prefer direct/generated Go interop.

Do not manually wrap every method.

## Bridge documentation

Create:

```text
letgo/native/gtkbridge/README.md
```

For every native wrapper, document:

```text
wrapper
why required
what upstream let-go limitation requires it
whether it can disappear later
```

## Exit criteria

The bridge has an intentionally small documented public API.

---

# Milestone 6 — Implement `glimmer-gtk` backend in Clojure

## Objective

Implement the GTK-specific backend in `.lg`.

This is where GTK semantics belong.

## Initial widget vocabulary

Implement only what the time tracker requires:

```text
:window
:vbox
:hbox
:label
:entry
:button
:check-button
:dropdown
:scrolled-window
```

Potential later additions:

```text
:separator
:list-box
:list-box-row
:spinner
```

Do not add widgets without a consumer.

## Widget creation

Example conceptual dispatcher:

```clojure
(defn create! [tag props]
  (case tag
    :label  (create-label props)
    :button (create-button props)
    :entry  (create-entry props)
    ...))
```

Keep this in Clojure.

## Properties

Support normalized props such as:

```clojure
{:label "Start"
 :sensitive true
 :visible true
 :hexpand true
 :vexpand false}
```

Entry:

```clojure
{:text "..."
 :placeholder "..."
 :on-change handler}
```

Button:

```clojure
{:label "Start"
 :on-click handler}
```

Check button:

```clojure
{:label "Billable"
 :active true
 :on-change handler}
```

Window:

```clojure
{:title "Go Time Tracker"
 :width 800
 :height 600}
```

## Property diffing

Avoid reconnecting event handlers unnecessarily.

Define clear behavior when handlers change.

Be particularly careful about:

* duplicate signal handlers;
* handler lifecycle;
* closures retaining component state;
* widget destruction.

## Containers

Define explicit GTK semantics for:

* boxes;
* scrolled window;
* application window.

Not all GTK containers expose the same child API.

Hide this difference behind:

```text
append-child!
remove-child!
replace-child!
reorder-child!
```

## Scheduler

Implement backend `:schedule` using the proven GLib main-loop bridge.

## Application run

`glimmer-gtk/run` should own:

```text
GtkApplication
activate callback
ApplicationWindow
root mount
present
main loop
```

but the root widget tree remains defined by Clojure components.

## Tests

Unit-test widget dispatch and property normalization without requiring an interactive desktop where possible.

Add an Xvfb/headless GTK smoke test if practical in CI.

## Exit criteria

A Glimmer-style counter application runs through the actual GTK backend with no application-specific Go UI code.

---

# Milestone 7 — Port the current `gtt-gtk` UI as Glimmer components

## Objective

Reproduce the current time-tracker GTK UI using Clojure components and reactive state.

Do not port networking yet unless necessary.

For this milestone, the existing Go application service may temporarily remain behind a narrow host boundary if that substantially accelerates proving the UI architecture.

However, application-specific host bindings must be considered temporary and clearly isolated.

## Application state

Use a single explicit application state shape.

Example:

```clojure
{:status :loading

 :snapshot
 {:running nil
  :today-ms 0
  :entries []
  :projects []}

 :tasks []

 :form
 {:description ""
  :project-id nil
  :task-id nil
  :billable? false}

 :request
 {:kind nil
  :pending? false}

 :error nil}
```

Refine as needed.

## Components

Split view into components such as:

```text
app
├── timer-form
│   ├── description-field
│   ├── project-select
│   ├── task-select
│   └── billable-toggle
├── timer-status
├── timer-actions
├── error-banner
└── recent-entries
```

Do not create one huge render function.

## Timer form

Implement:

* description;
* project;
* task;
* billable.

Selections should update application state, not widget-local Go state.

## Project/task relationship

When project changes:

```text
project selection
      ↓
state project-id changes
      ↓
task-id reset
      ↓
tasks loaded
      ↓
task dropdown rerenders
```

Do not manually push values directly from one GTK callback into another widget.

## Running timer state

When a timer is running:

```text
Start disabled
Stop enabled
Save enabled
form populated from running entry
```

When not running:

```text
Start enabled
Stop disabled
Save disabled
form reset appropriately
```

This should be derived during render rather than manually synchronized across widget callbacks.

## Recent entries

Render recent entries from immutable application data.

Do not store GTK row handles in application state.

## Errors

Errors should update reactive state.

Example:

```clojure
(swap! app-state assoc :error "...")
```

The UI should render error state declaratively.

## Loading state

Represent loading explicitly.

Avoid relying on "Loading..." widget mutation as implicit state.

## Exit criteria

The let-go/Glimmer UI matches the essential behavior of current `gtt-gtk` using mock or existing application services.

---

# Milestone 8 — Implement real application actions asynchronously

## Objective

Connect the reactive UI to real application effects without blocking GTK.

## Action pattern

Use named effectful actions:

```clojure
(refresh!)
(start!)
(stop!)
(save!)
(select-project!)
(delete-entry!)
(continue-entry!)
```

An action may:

```text
1. update state to pending
2. start background work
3. call provider/application service
4. marshal result to application state
5. Glimmer rerenders automatically
```

## Threading rule

No GTK mutation may happen directly from background work.

Background workers update application state.

Reactive rendering is scheduled onto GTK's main loop.

## Request coordination

Prevent obvious races such as:

```text
refresh A starts
refresh B starts
B finishes
A finishes later and overwrites B
```

Use request IDs/generations if needed.

## Error handling

Every asynchronous effect should produce either:

```clojure
{:ok value}
```

or:

```clojure
{:error error-data}
```

before touching state.

Do not allow worker failures to disappear into logs.

## Exit criteria

The real application can be used interactively without GTK blocking during HTTP requests.

---

# Milestone 9 — Port domain/application layer to let-go

## Objective

Remove the dependency on the Go application service so the time-tracker application's behavior is predominantly Clojure.

Do this only after the UI architecture is proven.

## Port order

Recommended:

```text
domain values
→ formatting
→ application use cases
→ config
→ provider abstraction
→ Clockify
→ Kimai
```

## Domain representation

Prefer ordinary Clojure values:

```clojure
{:id ...
 :description ...
 :start ...
 :end ...
 :billable? ...
 :project-id ...
 :task-id ...}
```

Use records only where they materially improve protocol dispatch or clarity.

## Provider abstraction

Consider a Clojure protocol:

```clojure
(defprotocol TimeProvider
  (current-user [provider])
  (workspaces [provider])
  (current-timer [provider ctx])
  (recent-entries [provider ctx limit])
  (projects [provider ctx])
  (tasks [provider ctx project-id])
  (start-timer [provider ctx input])
  (stop-timer [provider ctx])
  ...)
```

Alternatively, use a map of functions if that is simpler and more idiomatic for let-go.

Do not reproduce Go interfaces for their own sake.

## HTTP

Use let-go's own HTTP/JSON support where it is sufficient.

Avoid introducing Go provider wrappers unless a concrete let-go HTTP limitation is found.

## Configuration

Preserve compatibility with the existing configuration file if practical:

```text
${XDG_CONFIG_HOME:-~/.config}/gtt/config.json
```

Preserve restrictive permissions where supported.

## Exit criteria

The GTK application can run without importing the existing Go `application`, `domain`, or provider adapter packages.

At this point Go should primarily be:

```text
let-go runtime
gotk4
gtkbridge
build glue
```

---

# Milestone 10 — Live nREPL development workflow

## Objective

Make the running GTK application useful as an interactive Clojure development environment.

## Required workflow

Start app with nREPL available.

From REPL:

```clojure
@state
```

should inspect live state.

Mutating:

```clojure
(r/swap! state assoc :error "REPL test")
```

should safely update the running GUI.

## Component reload

Investigate a mechanism similar to Glimmer's live root reload:

```clojure
(ui/reload!)
```

Desired workflow:

```text
edit component
evaluate new definition
call reload!
existing window reconciles in place
```

Do not require restarting GTK for every UI edit.

## Thread safety

REPL operations may arrive from a non-GTK thread.

All resulting GTK reconciliation must be scheduled through the backend.

## Developer documentation

Document editor setup for at least generic nREPL use.

Do not make editor-specific integration a hard dependency.

## Exit criteria

A developer can alter reactive state and redefine a component while the GTK app remains running.

---

# Milestone 11 — Native/AOT release build

## Objective

Produce a releasable binary that includes the required Go dependencies while retaining let-go's native/AOT benefits.

## Investigation

Determine the best current let-go path for:

```text
.lg sources
→ let-go compilation/AOT
→ generated Go
→ gotk4 + gtkbridge linked
→ native executable
```

Do not assume `lg -b` alone can add arbitrary Go dependencies to an already-built interpreter.

Follow current let-go AOT/native-entry examples.

## Requirements

Release build should:

* not require the `lg` executable at runtime;
* contain the application code;
* contain required Go packages;
* dynamically link GTK as appropriate for the platform;
* start directly into the application;
* retain useful source-level diagnostics where practical.

## Development build vs release build

It is acceptable to have:

```text
development:
let-go VM + nREPL + GTK host

release:
AOT/native generated Go executable
```

provided the same application `.lg` sources work in both modes.

## AOT compatibility gate

Run all core/headless tests in both:

* normal let-go VM mode;
* AOT mode where supported.

Do not wait until the final milestone to discover that the reconciler depends on an interpreter-only feature.

## Exit criteria

A native `gtt-letgo-gtk` executable can be built reproducibly.

---

# Milestone 12 — Performance and footprint comparison

## Objective

Determine whether the rewrite retains the project's lightweight-client motivation.

## Compare

At minimum:

```text
gtt-gtk Go/gotk4
gtt-letgo-gtk VM/development build
gtt-letgo-gtk AOT/native build
gtt-fyne
```

Clockify Electron numbers from the existing README may remain context, but rerun them only if useful.

## Measure

### Binary

* executable size;
* stripped size if relevant.

### Startup

Measure:

```text
process start
→ first GTK window visible
```

Use a repeatable methodology.

### Idle memory

Record:

* RSS;
* PSS where available;
* USS where useful.

Measure after startup settles.

### Running timer

Measure memory and CPU with timer updating once per second.

### Interaction

Verify no obvious degradation during:

* typing;
* selecting projects/tasks;
* rendering many entries;
* repeated start/stop;
* repeated refresh.

### Reactive overhead

Create a synthetic benchmark:

```text
100
1,000
10,000
```

reactive state updates against a headless backend.

This is not for micro-optimization yet; it is to catch pathological reconciliation behavior.

## Deliverable

```text
docs/letgo-benchmarks.md
```

## Exit criteria

Performance characteristics are measured, not guessed.

---

# Milestone 13 — Simplify the bridge and evaluate upstream let-go changes

## Objective

Review every remaining Go wrapper after the application works.

## For each bridge function

Classify as:

```text
A. permanently GTK-specific
B. generic Go callback limitation
C. generic let-go interop limitation
D. accidental complexity that can now be removed
```

## Candidate upstream improvements

Possible findings may include:

* generic let-go function → Go `func(...)` argument coercion;
* callback error propagation;
* lifecycle handling for escaped callbacks;
* improved generated bindings for opaque structs;
* direct callback interop in AOT builds;
* custom reactive mutation protocol ergonomics.

Only propose upstream changes backed by this working application.

Do not modify let-go itself merely to imitate Jolt.

## Goal

Ideally the final native layer shrinks toward:

```text
very small GTK lifecycle/callback adapter
```

or disappears for callbacks if let-go's generic Go function coercion becomes sufficient.

---

# Testing strategy

Use three levels.

## 1. Pure headless tests

Largest test layer.

Cover:

```text
ratoms
cursors
reactions
dependency tracking
component lifecycle
Hiccup interpretation
reconciliation
prop diffs
application state transitions
formatting
provider mapping
```

These should not import GTK.

## 2. GTK integration tests

Small focused set.

Cover:

```text
widget creation
prop application
child management
callback invocation
main-thread scheduling
window mount
widget destruction
```

Use Xvfb or an equivalent headless environment if feasible.

## 3. End-to-end smoke tests

Start real application against:

* mock provider;
* optionally a controlled API fixture.

Verify visible state and major actions.

Do not make external Clockify/Kimai availability a requirement for normal CI.

---

# Required test fixtures

Create a deterministic mock time provider containing:

```clojure
{:user ...}

{:projects
 [{:id "project-a" :name "Project A"}
  {:id "project-b" :name "Project B"}]

 :tasks
 {"project-a"
  [{:id "task-a1" :name "Task A1"}]}

 :entries
 [...]

 :running
 ...}
```

Support deterministic action responses.

This allows the GTK UI to be developed without real API credentials.

---

# Error-handling requirements

Errors crossing Go/let-go boundaries must not be silently ignored.

For callbacks:

```text
GTK callback
→ let-go invocation
→ failure
```

must have a defined policy.

During development, failures should at minimum:

* be logged with context;
* update an observable application error channel/state where suitable;
* not arbitrarily panic the entire GTK process unless the error is unrecoverable.

Document callback error semantics in the bridge.

---

# Memory/lifetime requirements

GTK/GObject and let-go/Go GC interactions deserve explicit testing.

Investigate:

* callbacks retained by GTK;
* callbacks removed with widgets;
* closures capturing reactive state;
* destroyed widgets remaining reachable;
* repeated component mount/unmount;
* repeated handler replacement.

Add a stress test that repeatedly creates and destroys a component tree.

Observe memory over many iterations.

A reactive UI that leaks every previous handler is not acceptable even if functional tests pass.

---

# Main-thread rules

Treat GTK thread ownership as a hard invariant.

Only code executing on the GTK main loop may mutate GTK widgets.

Therefore:

```text
HTTP goroutine
nREPL thread
background timer
future async work
```

must never directly call GTK.

They may update reactive state.

The Glimmer backend scheduler is responsible for getting reconciliation onto the GTK loop.

Document this invariant prominently in `glimmer_gtk/core.lg`.

---

# Hiccup API guidelines

Keep the initial API intentionally small and idiomatic.

Example application view:

```clojure
(defn timer-actions [{:keys [running? pending?]}]
  [:hbox {:spacing 6}
   [:button
    {:label "Start"
     :sensitive (and (not running?) (not pending?))
     :on-click actions/start!}]

   [:button
    {:label "Stop"
     :sensitive (and running? (not pending?))
     :on-click actions/stop!}]

   [:button
    {:label "Save changes"
     :sensitive (and running? (not pending?))
     :on-click actions/save!}]])
```

Prefer declarative props over imperative widget calls in application components.

---

# Application-state rule

The application state should be serializable or printable Clojure data wherever practical.

Running:

```clojure
(pp/pprint @state)
```

from nREPL should provide useful information about the application.

Opaque GTK handles must remain inside reconciler/backend instances, not application state.

This is important for:

* REPL inspection;
* tests;
* future persistence;
* agentic development;
* debugging.

---

# Agent workflow

For each milestone:

1. Inspect the relevant upstream/current code before implementation.
2. Write or port tests first where the API is sufficiently clear.
3. Implement the smallest passing vertical slice.
4. Run relevant tests.
5. Run full let-go experiment tests.
6. Manually smoke-test GTK if the milestone touches native UI.
7. Update documentation.
8. Commit one coherent milestone/submilestone at a time.

Do not mix:

```text
runtime interop changes
Glimmer core changes
GTK backend changes
gtt feature changes
```

into one large commit.

---

# Commit guidance

Prefer commits such as:

```text
experiment: add minimal let-go gotk4 host

glimmer: add backend abstraction

glimmer: add IDeref-based reactive atom

glimmer: add headless element reconciliation

gtk: add minimal widget bridge

gtk: implement Glimmer GTK backend

gtt: render timer form through Glimmer

gtt: add asynchronous start and stop actions

gtt: port Clockify provider to let-go HTTP

build: add let-go native GTK executable

docs: record let-go GTK benchmark results
```

Avoid commits such as:

```text
rewrite GTK UI
```

containing many unrelated layers.

---

# Decision gates

Several milestones are explicit stop/go points.

## Gate A — callback feasibility

After Milestone 1 ask:

> Can GTK reliably invoke let-go callbacks without requiring a large or fragile Go runtime layer?

If no, stop and document the blocker before porting Glimmer.

## Gate B — portable Glimmer viability

After Milestone 4 ask:

> Can let-go implement reactive cells and the reconciler cleanly enough that most Glimmer code remains Clojure?

If substantial runtime hacks are required, reassess before GTK work.

## Gate C — GTK backend boundary

After Milestone 6 ask:

> Is widget construction/property logic in Clojure while Go remains mechanical?

If application-specific UI starts accumulating in Go, refactor before proceeding.

## Gate D — application viability

After Milestone 9 ask:

> Can provider/application behavior live naturally in let-go without recreating large Go adapters?

If yes, continue toward an almost completely Clojure application.

## Gate E — release viability

After Milestone 12 ask:

> Are startup, footprint, and runtime behavior acceptable for the lightweight desktop-client goal?

Record the answer even if the experiment is technically successful but operationally unattractive.

---

# Expected final architecture

The desired repository dependency direction is:

```text
gtt.view
    │
    ├──> gtt.state
    ├──> gtt.actions
    └──> glimmer.core
              │
              ├──> glimmer.ratom
              └──> glimmer.backend
                         │
                         ▼
                  glimmer-gtk
                         │
                         ▼
                    gtkbridge
                         │
                         ▼
                       gotk4
                         │
                         ▼
                        GTK4
```

Application effects:

```text
gtt.actions
    │
    ▼
gtt.application
    │
    ▼
provider protocol
   / \
  /   \
Clockify Kimai
```

GTK must not appear below the view layer except through `glimmer-gtk`.

Provider APIs must not appear in Glimmer or GTK code.

---

# Desired source-code balance

The exact numbers are not requirements, but use them as a smell test.

Expected authored application/framework code:

```text
Clojure / let-go     overwhelming majority
Go bridge            small
generated Go         acceptable
gotk4                dependency
GTK                   system dependency
```

If the custom handwritten Go bridge grows into hundreds or thousands of lines of widget-specific application code, revisit the design.

Generated bindings do not count against the Clojure-first goal in the same way as hand-maintained business/UI logic.

---

# Final deliverables

The completed work should leave:

```text
letgo/src/glimmer/*
letgo/src/glimmer_gtk/*
letgo/src/gtt/*
letgo/native/gtkbridge/*
letgo/test/*
```

plus:

```text
docs/letgo-baseline.md
docs/letgo-architecture.md
docs/letgo-development.md
docs/letgo-benchmarks.md
```

and reproducible commands for:

```sh
# run tests
...

# run development GTK app
...

# run with nREPL
...

# build native/AOT release
...

# run GTK smoke tests
...
```

---

# Definition of done

The project is complete when:

* Glimmer's component/reconciler concept runs on let-go.
* The reactive core is written in Clojure.
* A headless backend verifies reconciliation independently of GTK.
* GTK widget semantics are implemented primarily in `.lg`.
* gotk4 callbacks successfully invoke let-go functions.
* Go contains only clearly justified host/mechanical code.
* The time tracker operates through the new reactive GTK frontend.
* Real Clockify/Kimai functionality works through the let-go application layer, unless explicitly documented as a separate follow-up.
* nREPL can inspect and mutate the live application.
* REPL-driven reactive changes are safely scheduled onto the GTK main loop.
* A standalone native/AOT executable can be produced.
* VM and AOT paths are tested where applicable.
* Memory/startup/binary measurements are recorded.
* Remaining Go wrappers are reviewed for possible removal or upstream let-go improvements.
* The architecture remains understandable as a Clojure application with a small native GTK host seam rather than a Go GUI embedding a scripting language.

The most important outcome is not merely a working GTK window. It is demonstrating that **let-go can act as the primary implementation language for a native reactive desktop application, while Go/gotk4 remains an implementation detail at the foreign-runtime boundary.**
