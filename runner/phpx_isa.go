//go:build phpx

package runner

import (
	"strings"

	"github.com/titpetric/phpx/model"
)

// PhpxIsA answers php's is_a(): an object is of the class `instanceof` names,
// and a class name string is, when allowString holds, a class that derives
// from or implements it. A scalar, null and an array answer false.
func (rt *Runtime) PhpxIsA(value any, class string, allowString bool) bool {
	if name, ok := value.(string); ok {
		if !allowString {
			return false
		}
		c, found := rt.lookupClass(strings.TrimPrefix(name, "\\"))
		if !found {
			return false
		}
		return classIsA(c, strings.TrimPrefix(class, "\\"))
	}
	return phpInstanceOf(value, class)
}

// classIsA reports whether c is the named class or lists it as an ancestor or
// an interface.
func classIsA(c *model.Class, want string) bool {
	if strings.EqualFold(c.Name, want) {
		return true
	}
	for _, list := range [][]string{c.Ancestors, c.Implements} {
		for _, name := range list {
			if strings.EqualFold(name, want) {
				return true
			}
		}
	}
	return false
}

// PhpxIsSubclassOf answers php's is_subclass_of(): the class of an object, or a
// class name string when allowString holds, derives from or implements the
// class named and is not that class itself. An interface name answers for the
// interfaces it extends.
func (rt *Runtime) PhpxIsSubclassOf(value any, class string, allowString bool) bool {
	want := strings.TrimPrefix(class, "\\")
	var c *model.Class
	switch v := value.(type) {
	case *model.Object:
		c = v.Class
	case string:
		if !allowString {
			return false
		}
		name := strings.TrimPrefix(v, "\\")
		if ok, _ := rt.ClassExists(name, true); ok {
			c, _ = rt.lookupClass(name)
		} else if ok, _ := rt.InterfaceExists(name, true); ok {
			if strings.EqualFold(name, want) {
				return false
			}
			if id, found := rt.lookupInterface(name); found && strings.EqualFold(want, "Stringable") {
				for _, m := range id.Methods {
					if strings.EqualFold(m.Name, "__toString") {
						return true
					}
				}
			}
			for _, parent := range rt.phpxInterfaceChain(name, nil) {
				if strings.EqualFold(rt.phpxInterfaceName(parent), want) || strings.EqualFold(parent, want) {
					return true
				}
			}
			return false
		}
	}
	if c == nil || strings.EqualFold(c.Name, want) {
		return false
	}
	if classIsA(c, want) {
		return true
	}
	if strings.EqualFold(want, "Stringable") {
		if _, ok := lookupPHPMethod(c, "__toString"); ok {
			return true
		}
	}
	for _, name := range rt.phpxClassInterfaces(c) {
		if strings.EqualFold(name, want) || strings.EqualFold(rt.phpxInterfaceName(name), want) {
			return true
		}
	}
	return false
}
