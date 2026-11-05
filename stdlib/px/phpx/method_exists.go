//go:build phpx

package phpx

import (
	"strings"

	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\MethodExists`, registerMethodExists)
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
		if runner.PhpxIsClosure(objectOrClass) {
			return closureMethod(method)
		}
		if name, isName := objectOrClass.(string); isName {
			if loaded, _ := rt.ClassExists(name, true); !loaded {
				return false
			}
		}
		return core(objectOrClass, method)
	})
}

// closureMethod reports whether Closure has the method: __invoke, bindTo,
// call, bind and fromCallable.
func closureMethod(name string) bool {
	switch strings.ToLower(name) {
	case "__invoke", "bindto", "call", "bind", "fromcallable":
		return true
	}
	return false
}
