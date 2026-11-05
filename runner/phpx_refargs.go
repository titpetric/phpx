//go:build phpx

package runner

import (
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxBindRefArgs binds the parameters a call passes by reference to the
// cells of the caller's variables, and answers the arguments as plain values.
// A *model.RefArg for a parameter taken by value is its value.
func phpxBindRefArgs(decl *model.FuncDecl, args []any, scope *Scope) []any {
	if !model.PhpxAnyByRef.Load() {
		return args
	}
	var out []any
	for i, a := range args {
		ra, ok := a.(*model.RefArg)
		if !ok {
			continue
		}
		if out == nil {
			out = append([]any(nil), args...)
		}
		out[i] = ra.Value
		if i < len(decl.Params) && decl.Params[i].ByRef && ra.Make != nil {
			scope.bindRef(decl.Params[i].Name, ra.Make())
		}
	}
	if out == nil {
		return args
	}
	scope.args = out
	return out
}

// phpxInvokeArgs strips the reference wrappers off the arguments of `$f(...)`
// when $f is a Go function or the name of a builtin; a closure, a declared
// function name, an array callable and an invokable object bind their own
// parameters and read a wrapper themselves.
func (rt *Runtime) phpxInvokeArgs(callee any, args []any) []any {
	if !model.PhpxAnyByRef.Load() {
		return args
	}
	switch c := callee.(type) {
	case *Callable, *model.Array, *model.Object:
		return args
	case string:
		name := strings.TrimPrefix(c, "\\")
		if _, user := rt.userFns[name]; user || strings.Contains(name, "::") {
			return args
		}
	}
	return phpxBareArgs(args)
}
