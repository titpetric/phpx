//go:build phpx

package runner

import (
	"fmt"
	"github.com/titpetric/phpx/modelutil"
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxRefAssign runs `target = &source`: the target becomes another name for
// the source's slot. A variable and an object property are slots; anything else
// (an array element, a static property, a call) gives the target the value
// alone, which is the part of the reference this runtime keeps.
func (rt *Runtime) phpxRefAssign(target, source model.Expr, scope *Scope) (any, error) {
	cell, err := rt.phpxCellOf(source, scope)
	if err != nil {
		return nil, err
	}
	switch tgt := modelutil.UnwrapParenthesized(target).(type) {
	case *model.Var:
		scope.bindRef(tgt.Name, cell)
	case *model.PropAccess:
		base, err := rt.Eval(tgt.Base, scope)
		if err != nil {
			return nil, err
		}
		name, err := rt.propName(tgt, scope)
		if err != nil {
			return nil, err
		}
		obj, ok := base.(*model.Object)
		if !ok {
			return nil, fmt.Errorf("assign: %q is not an object property", name)
		}
		obj.BindProp(rt.phpxSlotKey(obj, name, scope), cell)
	case *model.Index:
		if err := rt.phpxBindElementRef(tgt, cell, scope); err != nil {
			return nil, err
		}
	default:
		if err := rt.assignTo(target, cell.V, scope); err != nil {
			return nil, err
		}
	}
	return cell.V, nil
}

// phpxBindElementRef stores cell as an array element, which is what
// `$a[key] = &$x` and `$a[] = &$x` do: the element reads and writes the slot
// the cell names. A key of nil appends, as php does.
func (rt *Runtime) phpxBindElementRef(tgt *model.Index, cell *model.Cell, scope *Scope) error {
	base, err := rt.container(tgt.Base, scope)
	if err != nil {
		return err
	}
	arr, ok := base.(*model.Array)
	if !ok {
		// Anything that is not an array takes the value alone, the part of a
		// reference this runtime keeps for an object or a native collection.
		return rt.assignTo(tgt, cell.V, scope)
	}
	arr.Split()
	if tgt.Index == nil {
		arr.Append(cell)
		return nil
	}
	key, err := rt.Eval(tgt.Index, scope)
	if err != nil {
		return err
	}
	arr.Set(normalizeKey(key), cell)
	return nil
}

// phpxCellOf answers the cell behind an lvalue, making one the first time it is
// shared, and a fresh cell around the value of any other expression.
func (rt *Runtime) phpxCellOf(e model.Expr, scope *Scope) (*model.Cell, error) {
	switch x := modelutil.UnwrapParenthesized(e).(type) {
	case *model.Var:
		return scope.refCell(x.Name), nil
	case *model.PropAccess:
		base, err := rt.Eval(x.Base, scope)
		if err != nil {
			return nil, err
		}
		name, err := rt.propName(x, scope)
		if err != nil {
			return nil, err
		}
		if obj, ok := base.(*model.Object); ok {
			return obj.PropCell(rt.phpxSlotKey(obj, name, scope)), nil
		}
	}
	if idx, ok := modelutil.UnwrapParenthesized(e).(*model.Index); ok {
		if cell, err := rt.phpxElementCell(idx, scope); cell != nil || err != nil {
			return cell, err
		}
	}
	v, err := rt.Eval(e, scope)
	if err != nil {
		return nil, err
	}
	return &model.Cell{V: v}, nil
}

// phpxElementCell answers a cell for `&$array[key]` or `&$array[]`: it holds
// the element's value and writes every value assigned through it back to the
// element. A write to the element does not reach the cell, which is the half of
// php's reference this runtime keeps.
func (rt *Runtime) phpxElementCell(x *model.Index, scope *Scope) (*model.Cell, error) {
	base, err := rt.container(x.Base, scope)
	if err != nil {
		return nil, err
	}
	arr, ok := base.(*model.Array)
	if !ok {
		return nil, nil
	}
	// The cell writes the element back, so an array shared with a copy splits
	// here: the element it hands out is this container's own handle.
	arr.Split()
	var key any
	if x.Index == nil {
		arr.Append(nil)
		keys := arr.Keys()
		key = keys[len(keys)-1]
	} else {
		k, err := rt.Eval(x.Index, scope)
		if err != nil {
			return nil, err
		}
		key = normalizeKey(k)
	}
	// The element becomes a reference, as in php: a cell already there is the
	// slot, and a plain value gets one installed under it. Every reader and
	// writer of the element goes through the cell from here.
	if cell, ok := arr.CellAt(key); ok {
		return cell, nil
	}
	cur, _ := arr.Get(key)
	cell := &model.Cell{V: cur}
	arr.Set(key, cell)
	return cell, nil
}

// phpxInstallRefAssign adds __phpx_refassign(node), the helper a compiled
// `$a->b = &$c` calls.
func (rt *Runtime) phpxInstallRefAssign(env map[string]any, ref *scopeRef) {
	env["__phpx_refassign"] = adapt(func(node *model.AssignExpr) (any, error) {
		src, _ := node.Value.(*model.Ref)
		return rt.phpxRefAssign(node.Target, src.X, ref.scope)
	})
}

// phpxInstallRefArg adds __phpx_refarg(expr, name, fallback), the helper that
// passes an lvalue argument as a *model.RefArg, so a by-reference parameter
// binds the caller's slot and every other callee reads the value. The helper is
// a typed function and not the uniform variadic shape: once a program declares
// one by-reference parameter every argument of every call comes through here,
// and the variadic call put each argument into an []any on the heap.
func (rt *Runtime) phpxInstallRefArg(env map[string]any, ref *scopeRef) {
	env["__phpx_refarg"] = func(e model.Expr, name, fallback string) (out any, err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				out, err = nil, recoveredError(recovered, "__phpx_refarg")
			}
		}()
		v, err := rt.Eval(e, ref.scope)
		if err != nil {
			return nil, err
		}
		if name != "" && !rt.phpxIsUserFunc(name, fallback) {
			return v, nil
		}
		scope := ref.scope
		return &model.RefArg{Value: v, Make: func() *model.Cell {
			c, err := rt.phpxCellOf(e, scope)
			if err != nil {
				return &model.Cell{V: v}
			}
			return c
		}}, nil
	}
}

// phpxInstallSplitBase adds __phpx_splitbase(expr), the helper an argument a
// builtin writes in place resolves through. The container it names splits
// along the path, so the write lands in the caller's array and not in a copy
// that shares its storage.
func (rt *Runtime) phpxInstallSplitBase(env map[string]any, ref *scopeRef) {
	env["__phpx_splitbase"] = adapt(func(e model.Expr) (any, error) {
		return rt.phpxSplitBase(e, ref.scope)
	})
}

// phpxIsUserFunc reports whether a call by name or by its global fallback
// reaches a function written in php; a builtin takes plain values.
func (rt *Runtime) phpxIsUserFunc(name, fallback string) bool {
	if name != "" {
		if _, ok := rt.userFns[strings.TrimPrefix(name, "\\")]; ok {
			return true
		}
	}
	if fallback != "" {
		if _, ok := rt.userFns[strings.TrimPrefix(fallback, "\\")]; ok {
			return true
		}
	}
	return false
}

// phpxInstallTraverse adds __phpx_traverse(object), the array of keys and
// values a Traversable object yields, for a spread to unpack.
func (rt *Runtime) phpxInstallTraverse(env map[string]any) {
	env["__phpx_traverse"] = adapt(func(obj *model.Object) (any, error) {
		out := model.NewArray()
		_, err := rt.PhpxIterate(obj, func(k, v any) bool {
			if _, isStr := k.(string); isStr {
				out.Set(k, v)
			} else {
				out.Append(v)
			}
			return true
		})
		return out, err
	})
}

// phpxIsRefAssign reports whether an assignment binds by reference and answers the source.
func phpxIsRefAssign(op string, value model.Expr) (model.Expr, bool) {
	if op != "=" && op != "" {
		return nil, false
	}
	if ref, ok := value.(*model.Ref); ok {
		return ref.X, true
	}
	return nil, false
}
