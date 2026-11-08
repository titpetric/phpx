//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/modelutil"
	"sort"
	"strings"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

// registerArrayPredicates installs php 8.4's array_find family and the
// forward_static_call pair.
func registerArrayPredicates(rt *runner.Runtime) {
	scan := func(arr, cb any) (entry, bool, bool, error) {
		fn, ok := rt.Callable(cb)
		if !ok {
			return entry{}, false, false, &valueError{message: "Argument #2 ($callback) must be a valid callback"}
		}
		for _, e := range entriesOf(arr) {
			r, err := fn(e.val, e.key)
			if err != nil {
				return entry{}, false, false, err
			}
			if phpval.Truthy(r) {
				return e, true, true, nil
			}
		}
		return entry{}, false, true, nil
	}
	rt.RegisterFunc("array_find", func(arr, cb any) (any, error) {
		e, found, _, err := scan(arr, cb)
		if !found {
			return nil, err
		}
		return e.val, nil
	})
	rt.RegisterFunc("array_find_key", func(arr, cb any) (any, error) {
		e, found, _, err := scan(arr, cb)
		if !found {
			return nil, err
		}
		return e.key, nil
	})
	rt.RegisterFunc("array_any", func(arr, cb any) (bool, error) {
		_, found, _, err := scan(arr, cb)
		return found, err
	})
	rt.RegisterFunc("array_all", func(arr, cb any) (bool, error) {
		fn, ok := rt.Callable(cb)
		if !ok {
			return false, &valueError{message: "Argument #2 ($callback) must be a valid callback"}
		}
		for _, e := range entriesOf(arr) {
			r, err := fn(e.val, e.key)
			if err != nil || !phpval.Truthy(r) {
				return false, err
			}
		}
		return true, nil
	})
	rt.RegisterFunc("forward_static_call", func(cb any, args ...any) (any, error) {
		fn, ok := rt.Callable(cb)
		if !ok {
			return nil, &valueError{message: "forward_static_call(): Argument #1 ($callback) must be a valid callback"}
		}
		return fn(args...)
	})
	rt.RegisterFunc("forward_static_call_array", func(cb any, args any) (any, error) {
		fn, ok := rt.Callable(cb)
		if !ok {
			return nil, &valueError{message: "forward_static_call_array(): Argument #1 ($callback) must be a valid callback"}
		}
		var list []any
		for _, e := range entriesOf(args) {
			list = append(list, e.val)
		}
		return fn(list...)
	})
}

// registerArrayWalk installs array_walk and array_walk_recursive. The callback
// takes the value by reference: a closure gets a cell, and the value it leaves
// there is written back to the element.
func registerArrayWalk(rt *runner.Runtime) {
	var walk func(arr any, cb any, extra []any, recursive bool) error
	walk = func(arr any, cb any, extra []any, recursive bool) error {
		fn, ok := rt.Callable(cb)
		if !ok {
			return &valueError{message: "array_walk(): Argument #2 ($callback) must be a valid callback"}
		}
		target, writable := arr.(*model.Array)
		for _, e := range entriesOf(arr) {
			if recursive && modelutil.IsCollection(e.val) {
				if err := walk(e.val, cb, extra, true); err != nil {
					return err
				}
				continue
			}
			var cell *model.Cell
			first := e.val
			if runner.PhpxIsClosure(cb) {
				first = &model.RefArg{Value: e.val, Make: func() *model.Cell {
					cell = &model.Cell{V: e.val}
					return cell
				}}
			}
			args := append([]any{first, e.key}, extra...)
			if _, err := fn(args...); err != nil {
				return err
			}
			if cell != nil && writable {
				target.Set(e.key, cell.V)
			}
		}
		return nil
	}
	rt.RegisterFunc("array_walk", func(arr, cb any, extra ...any) (bool, error) {
		return true, walk(arr, cb, extra, false)
	})
	rt.RegisterFunc("array_walk_recursive", func(arr, cb any, extra ...any) (bool, error) {
		return true, walk(arr, cb, extra, true)
	})
}

// registerNatSort installs natsort and natcasesort, which order an array by
// natural comparison and keep the keys.
func registerNatSort(rt *runner.Runtime) {
	natural := func(fold bool) func(arr any) bool {
		return func(arr any) bool {
			target, ok := arr.(*model.Array)
			if !ok {
				return false
			}
			cmp, found := rt.Callable("strnatcmp")
			if !found {
				return false
			}
			items := entriesOf(target)
			sort.SliceStable(items, func(i, j int) bool {
				a, b := phpval.String(items[i].val), phpval.String(items[j].val)
				if fold {
					a, b = strings.ToLower(a), strings.ToLower(b)
				}
				r, _ := cmp(a, b)
				return phpval.Int(r) < 0
			})
			target.Clear()
			for _, e := range items {
				target.Set(e.key, e.val)
			}
			return true
		}
	}
	rt.RegisterFunc("natsort", natural(false))
	rt.RegisterFunc("natcasesort", natural(true))
}
