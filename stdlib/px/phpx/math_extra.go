//go:build phpx

package phpx

import (
	"math"
	"math/rand/v2"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\MathExtra`, registerMathExtra)
}

// registerMathExtra installs the trigonometric and logarithmic functions the
// core math binding leaves out, and the random aliases.
func registerMathExtra(rt *runner.Runtime) {
	unary := map[string]func(float64) float64{
		"sin": math.Sin, "cos": math.Cos, "tan": math.Tan,
		"asin": math.Asin, "acos": math.Acos, "atan": math.Atan,
		"sinh": math.Sinh, "cosh": math.Cosh, "tanh": math.Tanh,
		"asinh": math.Asinh, "acosh": math.Acosh, "atanh": math.Atanh,
		"expm1": math.Expm1, "log1p": math.Log1p, "exp": math.Exp, "log10": math.Log10,
		"deg2rad": func(x float64) float64 { return x / 180 * math.Pi },
		"rad2deg": func(x float64) float64 { return x / math.Pi * 180 },
	}
	for name, f := range unary {
		rt.RegisterFunc(name, func(x any) float64 { return f(phpval.Float(x)) })
	}
	rt.RegisterFunc("atan2", func(y, x any) float64 { return math.Atan2(phpval.Float(y), phpval.Float(x)) })
	rt.RegisterFunc("hypot", func(x, y any) float64 { return math.Hypot(phpval.Float(x), phpval.Float(y)) })
	rt.RegisterFunc("pi", func() float64 { return math.Pi })
	rt.RegisterFunc("lcg_value", func() float64 { return rand.Float64() })
	rt.RegisterFunc("mt_getrandmax", func() int64 { return 2147483647 })
	rt.RegisterFunc("getrandmax", func() int64 { return 2147483647 })
	rt.RegisterFunc("srand", func(rest ...any) (any, error) {
		seed, ok := rt.Callable("mt_srand")
		if !ok {
			return nil, nil
		}
		return seed(rest...)
	})
}
