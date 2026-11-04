//go:build phpx

package phpx

import (
	"reflect"

	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxSortFlags)
}

// registerPhpxSortFlags lets the sort family take php's second argument, the
// flags (SORT_STRING, SORT_REGULAR, ...). Upstream's functions take the array
// alone, and the flags are accepted and ignored: the order is the regular one.
func registerPhpxSortFlags(rt *runner.Runtime) {
	for _, name := range []string{"sort", "rsort", "asort", "arsort", "ksort", "krsort"} {
		fn, ok := rt.LookupFunc(name)
		if !ok {
			continue
		}
		ft := reflect.TypeOf(fn)
		if ft == nil || ft.Kind() != reflect.Func || ft.NumIn() != 1 || ft.IsVariadic() {
			continue
		}
		in := []reflect.Type{ft.In(0), reflect.TypeOf([]any(nil))}
		out := make([]reflect.Type, ft.NumOut())
		for i := range out {
			out[i] = ft.Out(i)
		}
		inner := reflect.ValueOf(fn)
		wrapped := reflect.MakeFunc(reflect.FuncOf(in, out, true), func(args []reflect.Value) []reflect.Value {
			return inner.Call(args[:1])
		})
		rt.RegisterFunc(name, wrapped.Interface())
	}
}
