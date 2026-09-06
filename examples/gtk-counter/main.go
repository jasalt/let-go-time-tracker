package main

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/nooga/let-go/pkg/api"
	"github.com/nooga/let-go/pkg/compiler"
	"github.com/nooga/let-go/pkg/nrepl"
	"github.com/nooga/let-go/pkg/vm"
	"lgtt/native/gtkbridge"
)

//go:embed app.lg
var program string

func invoke(context string, fn vm.Fn) bool {
	return gtkbridge.Invoke(context, fn)
}

func installHost(lg *api.LetGo) error {
	defs := map[string]any{
		"gtk-app": func(id string) *gtk.Application {
			return gtk.NewApplication(id, gio.ApplicationFlags(0))
		},
		"gtk-window": func(app *gtk.Application) *gtk.ApplicationWindow {
			return gtk.NewApplicationWindow(app)
		},
		"gtk-vbox": func(spacing int) *gtk.Box {
			return gtk.NewBox(gtk.OrientationVertical, spacing)
		},
		"gtk-label":          gtk.NewLabel,
		"gtk-button":         gtk.NewButtonWithLabel,
		"window-title!":      func(w *gtk.ApplicationWindow, s string) { w.SetTitle(s) },
		"window-size!":       func(w *gtk.ApplicationWindow, x, y int) { w.SetDefaultSize(x, y) },
		"window-child!":      func(w *gtk.ApplicationWindow, b *gtk.Box) { w.SetChild(b) },
		"window-present!":    func(w *gtk.ApplicationWindow) { w.Present() },
		"box-append-label!":  func(b *gtk.Box, child *gtk.Label) { b.Append(child) },
		"box-append-button!": func(b *gtk.Box, child *gtk.Button) { b.Append(child) },
		"box-append-any!": func(b *gtk.Box, child vm.Value) {
			b.Append(child.Unbox().(gtk.Widgetter))
		},
		"box-remove-any!": func(b *gtk.Box, child vm.Value) {
			b.Remove(child.Unbox().(gtk.Widgetter))
		},
		"window-child-any!": func(w *gtk.ApplicationWindow, child vm.Value) {
			w.SetChild(child.Unbox().(gtk.Widgetter))
		},
		"widget-sensitive!": func(widget vm.Value, value bool) {
			widget.Unbox().(interface{ SetSensitive(bool) }).SetSensitive(value)
		},
		"widget-visible!": func(widget vm.Value, value bool) {
			widget.Unbox().(interface{ SetVisible(bool) }).SetVisible(value)
		},
		"widget-hexpand!": func(widget vm.Value, value bool) {
			widget.Unbox().(interface{ SetHExpand(bool) }).SetHExpand(value)
		},
		"widget-vexpand!": func(widget vm.Value, value bool) {
			widget.Unbox().(interface{ SetVExpand(bool) }).SetVExpand(value)
		},
		"button-label!": func(button *gtk.Button, text string) { button.SetLabel(text) },
		"disconnect!": func(object vm.Value, signal glib.SignalHandle) {
			object.Unbox().(interface{ HandlerDisconnect(glib.SignalHandle) }).HandlerDisconnect(signal)
		},
		"label-text!": func(label *gtk.Label, text string) { label.SetText(text) },
		"on-activate!": func(app *gtk.Application, fn vm.Fn) {
			app.ConnectActivate(func() { invoke("activate", fn) })
		},
		"on-click!": func(button *gtk.Button, fn vm.Fn) {
			button.ConnectClicked(func() { invoke("clicked", fn) })
		},
		"on-idle!": func(fn vm.Fn) {
			glib.IdleAdd(func() { invoke("idle", fn) })
		},
		"every-ms!": func(ms int, fn vm.Fn) {
			glib.TimeoutAdd(uint(ms), func() bool { return invoke("timeout", fn) })
		},
		"app-run!": func(app *gtk.Application) int { return app.Run([]string{"gtk-counter"}) },
	}
	for name, value := range defs {
		if err := lg.Def(name, value); err != nil {
			return fmt.Errorf("define %s: %w", name, err)
		}
	}
	return nil
}

func newRuntime() (*api.LetGo, error) {
	lg, err := api.NewLetGo("gtk-counter")
	if err != nil {
		return nil, err
	}
	lg.SetLoadPath([]string{"src", "."})
	if err := installHost(lg); err != nil {
		return nil, err
	}
	return lg, nil
}

func run() error {
	lg, err := newRuntime()
	if err != nil {
		return err
	}
	for _, namespace := range []string{"glimmer.backend", "glimmer.ratom", "glimmer.core", "glimmer-gtk.core"} {
		if _, err := lg.Run("(require '" + namespace + ")"); err != nil {
			return fmt.Errorf("require %s: %w", namespace, err)
		}
	}
	var server *nrepl.NreplServer
	if configuredPort := os.Getenv("LGTT_NREPL_PORT"); configuredPort != "" {
		port, err := strconv.Atoi(configuredPort)
		if err != nil {
			return fmt.Errorf("invalid LGTT_NREPL_PORT: %w", err)
		}
		contextProvider := any(lg).(interface{ CompilerContext() *compiler.Context })
		server = nrepl.NewNreplServer(contextProvider.CompilerContext())
		if err := server.Start(port); err != nil {
			return fmt.Errorf("start nREPL: %w", err)
		}
		defer server.Stop()
	}
	_, err = lg.Run(program)
	return err
}

func main() {
	runtime.LockOSThread()
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gtk-counter:", err)
		os.Exit(1)
	}
}
