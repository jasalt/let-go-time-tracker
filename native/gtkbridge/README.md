# GTK bridge boundary

`gtkbridge` is the only handwritten native callback seam for the production
backend. Application state, widget selection, properties, child hierarchy, and
reconciliation are forbidden here.

| Wrapper | Why required | Current let-go limitation | Can disappear? |
| --- | --- | --- | --- |
| `Invoke` | Calls a retained `vm.Fn` and reports errors | gotk4 cannot directly call a let-go function | Yes, if generic callback coercion gains error/lifetime semantics |
| `ConnectClicked` | Converts `vm.Fn` to retained `func()` | Generated method requires concrete callback | Yes |
| `ConnectChanged` | Same for entry changes | Same | Yes |
| `ConnectToggled` | Same for check buttons | Same | Yes |
| `IdleAdd` | Converts a one-shot let-go callback | GLib accepts a reflected callback and retains it | Yes |
| `TimeoutAdd` | Converts a repeating callback and stops it on error | Same, plus bool source-lifetime contract | Yes |

Ordinary gotk4 constructors and methods are exposed by generated/reflective
interop in the host and are not wrapped here. Every callback failure is logged
with context and increments `CallbackErrorCount`; repeating sources are removed
on failure. GTK signal handles are returned so the backend can disconnect them
when handlers change or widgets are destroyed.
