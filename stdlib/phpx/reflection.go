//go:build phpx

package phpx

import (
	"fmt"
	"strings"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxReflection)
}

// registerPhpxReflection installs the primitives the Reflection prelude is
// written over: what a class declares, an instance without its constructor,
// and a property read or write whatever its visibility.
func registerPhpxReflection(rt *runner.Runtime) {
	rt.RegisterFunc("__phpx_class_info", func(v any) any { return rt.PhpxReflectClass(v, true) })
	rt.RegisterFunc("__phpx_new_noctor", func(class string) (any, error) { return rt.PhpxNewWithoutConstructor(class) })
	rt.RegisterFunc("__phpx_param_default", func(class, method string, i int64) (any, error) {
		return rt.PhpxParamDefault(class, method, int(i))
	})
	rt.RegisterFunc("__phpx_prop_get", func(obj any, name string) (any, error) {
		o, ok := obj.(*model.Object)
		if !ok {
			return nil, fmt.Errorf("ReflectionProperty::getValue(): Argument #1 ($object) must be provided for instance properties")
		}
		return runner.PhpxGetProp(o, name), nil
	})
	rt.RegisterFunc("__phpx_prop_set", func(obj any, name string, v any) error {
		o, ok := obj.(*model.Object)
		if !ok {
			return fmt.Errorf("ReflectionProperty::setValue(): Argument #1 ($objectOrValue) must be of type object")
		}
		runner.PhpxSetProp(o, name, v)
		return nil
	})
	rt.RegisterFunc("__phpx_prop_isset", func(obj any, name string) bool {
		o, ok := obj.(*model.Object)
		if !ok {
			return false
		}
		_, has := o.Props[name]
		return has
	})
	rt.RegisterFunc("__phpx_static_prop", func(class, name string, rest ...any) (any, error) {
		return rt.PhpxStaticProp(strings.TrimPrefix(class, "\\"), name, rest...)
	})
	// constant("Class::NAME") reads a class constant, inherited ones included.
	if upstream, ok := rt.LookupFunc("constant"); ok {
		if prev, ok := upstream.(func(string) (any, error)); ok {
			rt.RegisterFunc("constant", func(name string) (any, error) {
				if class, konst, found := strings.Cut(name, "::"); found {
					return rt.PhpxClassConst(class, konst)
				}
				return prev(name)
			})
		}
	}
	// __phpx_classref names the class a value stands for: an object's class, or
	// the string itself. It is what `$cls::NAME` reads its class from.
	rt.RegisterFunc("__phpx_classref", func(v any) (string, error) {
		switch t := v.(type) {
		case string:
			return strings.TrimPrefix(t, "\\"), nil
		case *model.Object:
			if t.Class != nil {
				return t.Class.Name, nil
			}
		}
		return "", fmt.Errorf("Cannot use a value of type %T as a class name", v)
	})
	// __phpx_attributes(file, key) lists the attributes written above a class,
	// method or property, and __phpx_doc(file, key) its doc comment.
	rt.RegisterFunc("__phpx_attributes", func(file, key string) any {
		attrs, _ := rt.PhpxAttributes(file, key)
		return attrs
	})
	rt.RegisterFunc("__phpx_doc", func(file, key string) any {
		_, doc := rt.PhpxAttributes(file, key)
		return doc
	})
	rt.RegisterFunc("__phpx_attribute_args", func(file, class, args string) (any, error) {
		return rt.PhpxAttributeArguments(file, class, args)
	})
	rt.RegisterFunc("__phpx_func_info", func(f any) any { return rt.PhpxFuncInfo(f) })
}
