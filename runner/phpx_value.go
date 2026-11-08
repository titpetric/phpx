//go:build phpx

package runner

import (
	"github.com/titpetric/phpx/modelutil"
	"sync"

	"github.com/titpetric/phpx/model"
)

// A php array is a value: `$b = $a; sort($b);` leaves $a alone. The runtime holds
// an array as a handle, so the copy php makes at assignment is made here, at the
// places a handle would otherwise be shared and then written through.

// phpxReadsStorage reports whether e reads an array out of storage that stays
// reachable after the read: a variable, a property, a static property or an
// element. An array literal and a call build a value of their own.
func phpxReadsStorage(e model.Expr) bool {
	switch n := modelutil.UnwrapParenthesized(e).(type) {
	case *model.Var:
		return !n.Const
	case *model.PropAccess, *model.StaticProp, *model.Index:
		return true
	case *model.Ternary:
		return phpxReadsStorage(n.Then) || phpxReadsStorage(n.Else)
	}
	return false
}

// phpxCopyRead copies an array that a read of storage produced.
func phpxCopyRead(e model.Expr, v any) any {
	if arr, ok := v.(*model.Array); ok && phpxReadsStorage(e) {
		return modelutil.CopyArray(arr)
	}
	return v
}

var phpxMutatedParams sync.Map // *model.FuncDecl -> map[string]bool

// phpxParamMutated reports whether the body writes through the parameter: an
// index or property write, or a by-reference call such as sort($p). Only then
// does the callee need an array of its own.
func phpxParamMutated(decl *model.FuncDecl, name string) bool {
	if v, ok := phpxMutatedParams.Load(decl); ok {
		return v.(map[string]bool)[name]
	}
	table := map[string]bool{}
	for _, p := range decl.Params {
		// A parameter taken by reference is the caller's array, and the write
		// through it is the point.
		table[p.Name] = !p.ByRef && modelutil.AssignsTo(&model.Var{Name: p.Name}, decl.Body)
	}
	phpxMutatedParams.Store(decl, table)
	return table[name]
}

// phpxBindArray is the value a parameter takes when the argument is an array.
func phpxBindArray(decl *model.FuncDecl, name string, arg any) any {
	if arr, ok := arg.(*model.Array); ok && phpxParamMutated(decl, name) {
		return modelutil.CopyArray(arr)
	}
	return phpxOwnedArray(arg)
}
