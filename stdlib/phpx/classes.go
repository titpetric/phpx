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
	runner.RegisterBinding(runner.PhpxRegisterPreludes)
	runner.RegisterBinding(func(rt *runner.Runtime) {
		rt.RegisterFunc("eval", func(code string) (any, error) { return rt.PhpxEval(code) })
	})
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
	// A value that is no script object goes to upstream's binding, which reads
	// the exported fields of what a binding returned.
	upstreamVars := func(any) *model.Array { return model.NewArray() }
	if prev, ok := rt.LookupFunc("get_object_vars"); ok {
		if fn, ok := prev.(func(any) *model.Array); ok {
			upstreamVars = fn
		}
	}
	rt.RegisterFunc("get_object_vars", func(object any) *model.Array {
		obj, ok := object.(*model.Object)
		if !ok {
			return upstreamVars(object)
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
	// This package installs after core, so the class-name form of
	// is_subclass_of replaces the object-only one core's reflect.go registers.
	rt.RegisterFunc("is_subclass_of", func(value any, class string, allowString ...bool) bool {
		return rt.PhpxIsSubclassOf(value, class, len(allowString) == 0 || allowString[0])
	})
	rt.RegisterFunc("class_implements", func(class any, autoload ...bool) any {
		return rt.PhpxClassImplements(class, len(autoload) == 0 || autoload[0])
	})
	rt.RegisterFunc("class_alias", func(original, alias string, autoload ...bool) bool {
		return rt.PhpxClassAlias(original, alias, len(autoload) == 0 || autoload[0])
	})
	if prior, ok := rt.LookupFunc("class_exists"); ok {
		if core, typed := prior.(func(string, ...bool) (bool, error)); typed {
			rt.RegisterFunc("class_exists", func(name string, autoload ...bool) (bool, error) {
				if runner.PhpxBuiltinClass(name) {
					return true, nil
				}
				return core(name, autoload...)
			})
		}
	}
	rt.RegisterFunc("interface_exists", func(name string, autoload ...bool) (bool, error) {
		if phpxBuiltinInterfaces[strings.ToLower(strings.TrimPrefix(name, "\\"))] {
			return true, nil
		}
		return rt.InterfaceExists(name, len(autoload) == 0 || autoload[0])
	})
}

// phpxBuiltinInterfaces are the interfaces php declares itself, which the
// interface table of a script does not hold.
var phpxBuiltinInterfaces = map[string]bool{
	"throwable": true, "stringable": true, "countable": true, "arrayaccess": true,
	"iterator": true, "iteratoraggregate": true, "traversable": true,
	"jsonserializable": true, "unitenum": true, "backedenum": true, "serializable": true,
}
