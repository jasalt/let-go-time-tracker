package gtkbridge

import (
	"testing"

	"github.com/nooga/let-go/pkg/api"
	"github.com/nooga/let-go/pkg/vm"
)

func TestInvokeReportsErrors(t *testing.T) {
	ResetCallbackErrorCount()
	ok := vm.MustBox(func() {}).(vm.Fn)
	if !Invoke("test", ok) {
		t.Fatal("successful callback failed")
	}
	lg, err := api.NewLetGo("bridge-test")
	if err != nil {
		t.Fatal(err)
	}
	value, err := lg.Run(`(fn [] (throw (ex-info "sentinel" {})))`)
	if err != nil {
		t.Fatal(err)
	}
	if Invoke("test", value.(vm.Fn)) {
		t.Fatal("failing callback succeeded")
	}
	if got := CallbackErrorCount(); got != 1 {
		t.Fatalf("error count = %d", got)
	}
}
