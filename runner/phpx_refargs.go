//go:build phpx

package runner

import "github.com/titpetric/phpx/model"

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
