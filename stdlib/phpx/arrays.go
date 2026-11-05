//go:build phpx

package phpx

import (
	"fmt"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Arrays`, registerPhpxArrays)
	runner.RegisterArea(`Runtime\PX\Arrays\ArrayMap`, registerPhpxArrayMap)
}

// registerPhpxArrays installs the array reducers and mergers that take arrays
// and nothing else.
func registerPhpxArrays(rt *runner.Runtime) {
	rt.RegisterFunc("array_product", func(values any) any {
		product := any(int64(1))
		model.ToArray(values).Range(func(_, v any) bool {
			product = phpval.Number(phpMul(product, phpval.Number(v)))
			return true
		})
		return product
	})
	rt.RegisterFunc("array_replace", func(base any, replacements ...any) *model.Array {
		out := model.NewArray()
		model.ToArray(base).Range(func(k, v any) bool { out.Set(k, v); return true })
		for _, r := range replacements {
			model.ToArray(r).Range(func(k, v any) bool { out.Set(k, v); return true })
		}
		return out
	})
}

// phpMul multiplies two numbers, an int product staying an int.
func phpMul(a, b any) any {
	x, xi := a.(int64)
	y, yi := b.(int64)
	if xi && yi {
		return phpval.MulInt(x, y)
	}
	return phpval.Float(a) * phpval.Float(b)
}

// registerPhpxArrayMap is array_map() as php has it: with one array the result
// keeps the array's keys, string keys included, and with several arrays the
// callback takes one element of each, the short ones padded with null. Upstream
// takes one array and renumbers.
func registerPhpxArrayMap(rt *runner.Runtime) {
	rt.RegisterFunc("array_map", func(callback any, arrays ...any) (any, error) {
		lists := make([]*model.Array, 0, len(arrays))
		for _, a := range arrays {
			arr, ok := a.(*model.Array)
			if !ok && a != nil && model.IsCollection(a) {
				arr, ok = model.ToArray(a), true
			}
			if !ok {
				return nil, fmt.Errorf("array_map(): Argument #2 ($array) must be of type array")
			}
			lists = append(lists, arr)
		}
		var call func(...any) (any, error)
		if callback != nil {
			c, ok := rt.Callable(callback)
			if !ok {
				return nil, fmt.Errorf("array_map(): Argument #1 ($callback) must be a valid callback or null")
			}
			call = c
		}
		if len(lists) == 1 {
			out := model.NewArraySize(lists[0].Len())
			var err error
			lists[0].Range(func(k, v any) bool {
				r := v
				if call != nil {
					if r, err = call(v); err != nil {
						return false
					}
				}
				out.Set(k, r)
				return true
			})
			return out, err
		}
		cols := make([][]any, len(lists))
		longest := 0
		for i, l := range lists {
			l.Range(func(_, v any) bool { cols[i] = append(cols[i], v); return true })
			longest = max(longest, len(cols[i]))
		}
		out := model.NewArraySize(longest)
		for row := 0; row < longest; row++ {
			args := make([]any, len(cols))
			for i := range cols {
				if row < len(cols[i]) {
					args[i] = cols[i][row]
				}
			}
			if call == nil {
				tuple := model.NewArraySize(len(args))
				for _, a := range args {
					tuple.Append(a)
				}
				out.Append(tuple)
				continue
			}
			r, err := call(args...)
			if err != nil {
				return nil, err
			}
			out.Append(r)
		}
		return out, nil
	})
}
