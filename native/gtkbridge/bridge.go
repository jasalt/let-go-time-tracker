// Package gtkbridge contains only mechanical conversions between retained
// gotk4 callbacks and let-go vm.Fn values. Widget construction and rendering
// policy belong in let-go.
package gtkbridge

import (
	"fmt"
	"os"
	"sync/atomic"

	"github.com/diamondburned/gotk4/pkg/core/glib"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/nooga/let-go/pkg/vm"
)

var callbackErrors atomic.Uint64

// Invoke calls a let-go callback and makes errors observable. It returns false
// so repeating GLib sources are removed after a callback failure.
func Invoke(context string, fn vm.Fn, args ...vm.Value) bool {
	if _, err := fn.Invoke(args); err != nil {
		callbackErrors.Add(1)
		_, _ = fmt.Fprintf(os.Stderr, "gtkbridge: %s callback: %v\n", context, err)
		return false
	}
	return true
}

func CallbackErrorCount() uint64 { return callbackErrors.Load() }
func ResetCallbackErrorCount()   { callbackErrors.Store(0) }

func ConnectClicked(button *gtk.Button, fn vm.Fn) glib.SignalHandle {
	return button.ConnectClicked(func() { _ = Invoke("clicked", fn) })
}

func ConnectChanged(entry *gtk.Entry, fn vm.Fn) glib.SignalHandle {
	return entry.ConnectChanged(func() { _ = Invoke("changed", fn) })
}

func ConnectToggled(button *gtk.CheckButton, fn vm.Fn) glib.SignalHandle {
	return button.ConnectToggled(func() { _ = Invoke("toggled", fn) })
}

func IdleAdd(fn vm.Fn) glib.SourceHandle {
	return glib.IdleAdd(func() { _ = Invoke("idle", fn) })
}

func TimeoutAdd(milliseconds uint, fn vm.Fn) glib.SourceHandle {
	return glib.TimeoutAdd(milliseconds, func() bool { return Invoke("timeout", fn) })
}
