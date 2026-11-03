//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxArrays)
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
