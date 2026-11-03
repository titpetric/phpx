//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
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
	// get_object_vars() lists the properties the calling body may read, as
	// php does: a protected one is out of reach from outside the family.
	rt.RegisterFunc("get_object_vars", func(object any) *model.Array {
		obj, ok := object.(*model.Object)
		if !ok {
			return model.NewArray()
		}
		out := model.NewArraySize(obj.Len())
		rt.PhpxVisibleProps(obj, func(name string, v any) bool {
			out.Set(phpval.Key(name), v)
			return true
		})
		return out
	})
	rt.RegisterFunc("class_parents", func(class any, autoload ...bool) any {
		return rt.PhpxClassParents(class, len(autoload) == 0 || autoload[0])
	})
	rt.RegisterFunc("interface_exists", func(name string, autoload ...bool) (bool, error) {
		return rt.InterfaceExists(name, len(autoload) == 0 || autoload[0])
	})
}
