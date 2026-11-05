//go:build phpx

package phpx

import (
	"fmt"

	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerNamedClosure)
}

// registerNamedClosure makes Closure::fromCallable('name') keep the name, which
// ReflectionFunction reads back; every other callable is what core answers.
func registerNamedClosure(rt *runner.Runtime) {
	rt.RegisterFunc("Closure::fromCallable", func(callback any) (any, error) {
		if runner.PhpxIsClosure(callback) {
			return callback, nil
		}
		if name, ok := callback.(string); ok {
			if c, found := rt.PhpxNamedCallable(name); found {
				return c, nil
			}
		}
		fn, ok := rt.Callable(callback)
		if !ok {
			return nil, fmt.Errorf("Closure::fromCallable(): argument #1 ($callback) is not callable")
		}
		return fn, nil
	})
}
