//go:build phpx

package runner

import (
	"strings"

	"github.com/titpetric/phpx/model"
)

// PhpxClassParents answers php's class_parents(): the chain `extends` names,
// nearest first, keyed by class name. A class no declaration answers is false.
func (rt *Runtime) PhpxClassParents(class any, autoload bool) any {
	var chain []string
	switch v := class.(type) {
	case *model.Object:
		if v.Class == nil {
			return false
		}
		chain = v.Class.Ancestors
	case string:
		name := strings.TrimPrefix(v, "\\")
		if ok, _ := rt.ClassExists(name, autoload); !ok {
			return false
		}
		c, found := rt.lookupClass(name)
		if !found {
			return false
		}
		chain = c.Ancestors
	default:
		return false
	}
	out := model.NewArray()
	for _, name := range chain {
		out.Set(name, name)
	}
	return out
}
