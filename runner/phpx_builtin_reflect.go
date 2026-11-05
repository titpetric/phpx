//go:build phpx

package runner

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxBuiltinFuncInfo describes a function the runtime binds from Go for
// ReflectionFunction: its name, and one parameter per Go parameter. A binding
// carries no parameter names, so they read arg0, arg1 and so on, and a
// variadic Go parameter is a variadic php one. It answers false for a name no
// binding holds.
func (rt *Runtime) phpxBuiltinFuncInfo(f any) any {
	var name string
	switch t := f.(type) {
	case string:
		name = strings.TrimPrefix(t, "\\")
	case *Callable:
		name = t.name
	}
	if name == "" {
		return false
	}
	fn, ok := rt.lookupFunc(name)
	if !ok {
		return false
	}
	return phpxMap("name", name, "params", phpxGoParamInfo(fn), "returnType", nil, "file", false)
}

// phpxGoParamInfo lists the parameters of a Go function in the shape
// phpxParamInfo gives a declaration.
func phpxGoParamInfo(fn any) *model.Array {
	out := model.NewArray()
	t := reflect.TypeOf(fn)
	if t == nil || t.Kind() != reflect.Func {
		return out
	}
	for i := 0; i < t.NumIn(); i++ {
		variadic := t.IsVariadic() && i == t.NumIn()-1
		out.Append(phpxMap(
			"name", "arg"+strconv.Itoa(i), "position", int64(i), "type", nil,
			"hasDefault", variadic, "variadic", variadic, "byRef", false,
			"promoted", false,
		))
	}
	return out
}

// PhpxNamedCallable answers a Closure that remembers the function name it was
// made from, so ReflectionFunction can describe it; false when no function
// has that name.
func (rt *Runtime) PhpxNamedCallable(name string) (any, bool) {
	name = strings.TrimPrefix(name, "\\")
	if _, ok := rt.lookupFunc(name); !ok {
		return nil, false
	}
	return rt.newNamedCallable(name), true
}

// PhpxIsClosure reports whether v is already a Closure value, which
// Closure::fromCallable answers as itself.
func PhpxIsClosure(v any) bool {
	_, ok := v.(*Callable)
	return ok
}

// PhpxCast applies a php cast by its type name, for settype().
func PhpxCast(typ string, v any) any { return helperCast(typ, v) }
