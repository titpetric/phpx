//go:build phpx

package runner

import (
	"strings"
	"sync"

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

// phpxImplementsAs keeps the `implements` names of a class as the declaration
// spelled them. model.Class holds them lower-cased, which is enough to answer
// instanceof and not enough for class_implements() to name them.
var phpxImplementsAs sync.Map // *model.Class -> []string

// phpxRecordImplements notes the names cd listed, as written.
func phpxRecordImplements(c *model.Class, cd *model.ClassDecl) {
	if len(cd.Implements) > 0 {
		phpxImplementsAs.Store(c, cd.Implements)
	}
}

// PhpxClassImplements answers php's class_implements(): every interface the
// class lists, its parents' and the ones they extend, keyed by name. A class no
// declaration answers is false.
func (rt *Runtime) PhpxClassImplements(class any, autoload bool) any {
	var names []string
	switch v := class.(type) {
	case *model.Object:
		if v.Class == nil {
			return false
		}
		names = rt.phpxClassInterfaces(v.Class)
	case string:
		name := strings.TrimPrefix(v, "\\")
		if ok, _ := rt.ClassExists(name, autoload); ok {
			c, found := rt.lookupClass(name)
			if !found {
				return false
			}
			names = rt.phpxClassInterfaces(c)
			break
		}
		if ok, _ := rt.InterfaceExists(name, autoload); !ok {
			return false
		}
		names = rt.phpxInterfaceChain(name, nil)
	default:
		return false
	}
	out := model.NewArray()
	for _, name := range names {
		name = rt.phpxInterfaceName(name)
		out.Set(name, name)
	}
	return out
}

// phpxClassInterfaces lists the interfaces of c and its ancestors with the
// interfaces those extend, nearest first.
func (rt *Runtime) phpxClassInterfaces(c *model.Class) []string {
	var out []string
	seen := map[string]bool{}
	add := func(name string) {
		if !seen[strings.ToLower(name)] {
			seen[strings.ToLower(name)] = true
			out = append(out, name)
		}
	}
	walk := func(cl *model.Class) {
		listed, ok := phpxImplementsAs.Load(cl)
		if !ok {
			for _, n := range cl.Implements {
				add(n)
			}
			return
		}
		for _, n := range listed.([]string) {
			n = strings.TrimPrefix(n, "\\")
			add(n)
			rt.InterfaceExists(n, true)
			for _, up := range rt.phpxInterfaceChain(n, nil) {
				add(up)
			}
		}
	}
	walk(c)
	for _, anc := range c.Ancestors {
		if parent, ok := rt.lookupClass(anc); ok {
			walk(parent)
		}
	}
	return out
}

// phpxBuiltinInterfaces spells the interfaces the engine knows without a
// declaration. The class table keeps `implements` names lower-cased.
var phpxBuiltinInterfaces = map[string]string{
	"countable": "Countable", "arrayaccess": "ArrayAccess", "iterator": "Iterator",
	"iteratoraggregate": "IteratorAggregate", "traversable": "Traversable",
	"stringable": "Stringable", "jsonserializable": "JsonSerializable",
	"throwable": "Throwable", "unitenum": "UnitEnum", "backedenum": "BackedEnum",
	"serializable": "Serializable",
}

// phpxInterfaceName restores the spelling an interface was declared with.
func (rt *Runtime) phpxInterfaceName(name string) string {
	if id, ok := rt.lookupInterface(name); ok {
		return id.Name
	}
	if spelled, ok := phpxBuiltinInterfaces[strings.ToLower(name)]; ok {
		return spelled
	}
	return name
}

// phpxInterfaceChain lists the interfaces an interface extends, transitively.
func (rt *Runtime) phpxInterfaceChain(name string, seen map[string]bool) []string {
	id, ok := rt.lookupInterface(name)
	if !ok {
		return nil
	}
	if seen == nil {
		seen = map[string]bool{}
	}
	var out []string
	for _, parent := range id.Extends {
		parent = strings.TrimPrefix(parent, "\\")
		if seen[parent] {
			continue
		}
		seen[parent] = true
		out = append(out, parent)
		out = append(out, rt.phpxInterfaceChain(parent, seen)...)
	}
	return out
}

// PhpxClassAlias makes alias another name for the declared class original, which
// is php's class_alias(). It reports false when original is not declared or
// alias is taken.
func (rt *Runtime) PhpxClassAlias(original, alias string, autoload bool) bool {
	original, alias = strings.TrimPrefix(original, "\\"), strings.TrimPrefix(alias, "\\")
	if ok, _ := rt.ClassExists(original, autoload); !ok {
		return false
	}
	c, found := rt.lookupClass(original)
	if !found {
		return false
	}
	if _, taken := rt.classes[alias]; taken {
		return false
	}
	rt.classes[alias] = c
	return true
}
