//go:build phpx

package phpx

import (
	"fmt"
	"strings"

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
	// Closure::bind rebinds `$this` and the class scope, which private
	// property access reads. Upstream answers the closure unchanged.
	rt.RegisterFunc("Closure::bind", func(closure any, args ...any) (any, error) {
		var newThis any
		if len(args) > 0 {
			newThis = args[0]
		}
		scope := ""
		if len(args) > 1 {
			switch s := args[1].(type) {
			case string:
				if s != "static" {
					scope = strings.TrimPrefix(s, "\\")
				}
			case *model.Object:
				if s.Class != nil {
					scope = s.Class.Name
				}
			}
		}
		out, ok := rt.PhpxRebind(closure, newThis, scope)
		if !ok {
			if _, callable := rt.Callable(closure); callable {
				return closure, nil
			}
			return nil, fmt.Errorf("Closure::bind(): argument #1 ($closure) must be a closure")
		}
		return out, nil
	})
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
	rt.RegisterFunc("class_implements", func(class any, autoload ...bool) any {
		return rt.PhpxClassImplements(class, len(autoload) == 0 || autoload[0])
	})
	rt.RegisterFunc("class_alias", func(original, alias string, autoload ...bool) bool {
		return rt.PhpxClassAlias(original, alias, len(autoload) == 0 || autoload[0])
	})
	rt.RegisterFunc("interface_exists", func(name string, autoload ...bool) (bool, error) {
		return rt.InterfaceExists(name, len(autoload) == 0 || autoload[0])
	})
}

func init() {
	runner.RegisterBinding(runner.PhpxRegisterPreludes)
}
