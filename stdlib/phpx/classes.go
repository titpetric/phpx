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
	rt.RegisterFunc("class_parents", func(class any, autoload ...bool) any {
		return rt.PhpxClassParents(class, len(autoload) == 0 || autoload[0])
	})
	rt.RegisterFunc("interface_exists", func(name string, autoload ...bool) (bool, error) {
		return rt.InterfaceExists(name, len(autoload) == 0 || autoload[0])
	})
}
