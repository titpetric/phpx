//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerMethodExists)
}

// registerMethodExists makes method_exists('Class', 'm') autoload Class, as php
// does, and leaves every other argument to the binding core registered.
func registerMethodExists(rt *runner.Runtime) {
	prior, ok := rt.LookupFunc("method_exists")
	core, typed := prior.(func(any, string) bool)
	if !ok || !typed {
		return
	}
	rt.RegisterFunc("method_exists", func(objectOrClass any, method string) bool {
		if name, isName := objectOrClass.(string); isName {
			if loaded, _ := rt.ClassExists(name, true); !loaded {
				return false
			}
		}
		return core(objectOrClass, method)
	})
}
