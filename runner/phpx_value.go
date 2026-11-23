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

// phpxSplitBase resolves the base of an lvalue for a write, splitting every
// array the write reaches through so the value it answers is the container's
// own handle and not one a copy also reads. It creates nothing: a path that
// names no storage answers nil.
func (rt *Runtime) phpxSplitBase(e model.Expr, scope *Scope) (any, error) {
	switch node := modelutil.UnwrapParenthesized(e).(type) {
	case *model.Index:
		parent, err := rt.phpxSplitBase(node.Base, scope)
		if err != nil || parent == nil || node.Index == nil {
			return nil, err
		}
		if arr, ok := parent.(*model.Array); ok {
			arr.Split()
		}
		key, err := rt.Eval(node.Index, scope)
		if err != nil {
			return nil, err
		}
		return helperIndex(parent, key), nil
	case *model.PropAccess:
		base, err := rt.Eval(node.Base, scope)
		if err != nil {
			return nil, err
		}
		name, err := rt.propName(node, scope)
		if err != nil {
			return nil, err
		}
		v := rt.helperGet(&scopeRef{scope: scope})(base, name)
		if arr, ok := v.(*model.Array); ok {
			arr.Split()
		}
		return v, nil
	}
	v, err := rt.Eval(e, scope)
	if arr, ok := v.(*model.Array); ok {
		arr.Split()
	}
	return v, err
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
