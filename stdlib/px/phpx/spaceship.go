//go:build phpx

package phpx

import (
	"math"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Spaceship`, registerPhpxSpaceship)
}

// registerPhpxSpaceship installs the function the parser lowers `<=>` to.
func registerPhpxSpaceship(rt *runner.Runtime) {
	rt.RegisterFunc("__phpx_spaceship", func(a, b any) int64 {
		if isNaN(a) || isNaN(b) {
			return 1
		}
		c := phpval.Compare(a, b)
		switch {
		case c < 0:
			return -1
		case c > 0:
			return 1
		}
		return 0
	})
}

func isNaN(v any) bool {
	f, ok := v.(float64)
	return ok && math.IsNaN(f)
}
