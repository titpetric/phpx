//go:build phpx

package runner

import (
	"fmt"
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
	switch tgt := model.UnwrapParenthesized(target).(type) {
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
		obj.BindProp(name, cell)
	default:
		if err := rt.assignTo(target, cell.V, scope); err != nil {
			return nil, err
		}
	}
	return cell.V, nil
}

// phpxCellOf answers the cell behind an lvalue, making one the first time it is
// shared, and a fresh cell around the value of any other expression.
func (rt *Runtime) phpxCellOf(e model.Expr, scope *Scope) (*model.Cell, error) {
	switch x := model.UnwrapParenthesized(e).(type) {
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
			return obj.PropCell(name), nil
		}
	}
	if idx, ok := model.UnwrapParenthesized(e).(*model.Index); ok {
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
	cur, _ := arr.Get(key)
	return &model.Cell{V: cur, Mirror: func(v any) { arr.Set(key, v) }}, nil
}

// phpxInstallRefAssign adds __phpx_refassign(node), the helper a compiled
// `$a->b = &$c` calls.
func (rt *Runtime) phpxInstallRefAssign(env map[string]any, ref *scopeRef) {
	env["__phpx_refassign"] = adapt(func(node *model.AssignExpr) (any, error) {
		src, _ := node.Value.(*model.Ref)
		return rt.phpxRefAssign(node.Target, src.X, ref.scope)
	})
}

// phpxInstallRefArg adds __phpx_refarg(expr), the helper that passes an
// lvalue argument as a *model.RefArg, so a by-reference parameter binds the
// caller's slot and every other callee reads the value.
func (rt *Runtime) phpxInstallRefArg(env map[string]any, ref *scopeRef) {
	env["__phpx_refarg"] = adapt(func(e model.Expr, name, fallback string) (any, error) {
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
	})
}

// phpxIsUserFunc reports whether a call by name or by its global fallback
// reaches a function written in php; a builtin takes plain values.
func (rt *Runtime) phpxIsUserFunc(name, fallback string) bool {
	for _, n := range []string{name, fallback} {
		if _, ok := rt.userFns[strings.TrimPrefix(n, "\\")]; ok && n != "" {
			return true
		}
	}
	return false
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
