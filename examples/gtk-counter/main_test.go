package main

import (
	"testing"

	"github.com/nooga/let-go/pkg/vm"
	"lgtt/native/gtkbridge"
)

func TestInvokeSuccessAndFailure(t *testing.T) {
	gtkbridge.ResetCallbackErrorCount()
	ok := vm.MustBox(func() {}).(vm.Fn)
	if !invoke("test", ok) {
		t.Fatal("successful callback reported failure")
	}
	lg, err := newRuntime()
	if err != nil {
		t.Fatal(err)
	}
	badValue, err := lg.Run(`(fn [] (throw (ex-info "sentinel" {})))`)
	if err != nil {
		t.Fatal(err)
	}
	if invoke("test", badValue.(vm.Fn)) {
		t.Fatal("failed callback reported success")
	}
	if got := gtkbridge.CallbackErrorCount(); got != 1 {
		t.Fatalf("callback error count = %d, want 1", got)
	}
}

func TestEmbeddedProgramCompiles(t *testing.T) {
	if program == "" {
		t.Fatal("embedded let-go program is empty")
	}
	lg, err := newRuntime()
	if err != nil {
		t.Fatal(err)
	}
	// Compile definitions without entering GTK's blocking run loop.
	if _, err := lg.Run("(do (def state (atom {:count 0})) (swap! state update :count inc) @state)"); err != nil {
		t.Fatal(err)
	}
}
