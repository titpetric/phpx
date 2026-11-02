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
