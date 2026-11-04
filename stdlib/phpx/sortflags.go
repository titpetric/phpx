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

func init() {
	runner.RegisterBinding(registerPhpxExtraArgs)
}

// registerPhpxExtraArgs lets the functions below take the trailing arguments php
// gives them, which upstream's signatures stop short of: preg_replace's $limit
// and $count are accepted and ignored.
func registerPhpxExtraArgs(rt *runner.Runtime) {
	for name, keep := range map[string]int{"preg_replace": 3} {
		fn, ok := rt.LookupFunc(name)
		if !ok {
			continue
		}
		ft := reflect.TypeOf(fn)
		if ft == nil || ft.Kind() != reflect.Func || ft.NumIn() != keep || ft.IsVariadic() {
			continue
		}
		in := make([]reflect.Type, 0, keep+1)
		for i := 0; i < keep; i++ {
			in = append(in, ft.In(i))
		}
		in = append(in, reflect.TypeOf([]any(nil)))
		out := make([]reflect.Type, ft.NumOut())
		for i := range out {
			out[i] = ft.Out(i)
		}
		inner := reflect.ValueOf(fn)
		rt.RegisterFunc(name, reflect.MakeFunc(reflect.FuncOf(in, out, true), func(args []reflect.Value) []reflect.Value {
			return inner.Call(args[:keep])
		}).Interface())
	}
}
