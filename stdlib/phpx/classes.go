//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxClasses)
}

// registerPhpxClasses installs the class introspection that reads the chain
// `extends` recorded.
func registerPhpxClasses(rt *runner.Runtime) {
	// get_class() with no argument names the class of the running method;
	// upstream answers false. Any argument goes to upstream's binding.
	if prev, ok := rt.LookupFunc("get_class"); ok {
		if upstream, ok := prev.(func(...any) any); ok {
			rt.RegisterFunc("get_class", func(object ...any) any {
				if len(object) == 0 {
					return rt.PhpxDeclaringClass()
				}
				return upstream(object...)
			})
		}
	}
	rt.RegisterFunc("class_parents", func(class any, autoload ...bool) any {
		return rt.PhpxClassParents(class, len(autoload) == 0 || autoload[0])
	})
	rt.RegisterFunc("interface_exists", func(name string, autoload ...bool) (bool, error) {
		return rt.InterfaceExists(name, len(autoload) == 0 || autoload[0])
	})
}
