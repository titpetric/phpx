//go:build phpx

package runner

import (
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxStaticKey is the MethodOwner key a static property's owner sits under.
// A method name cannot start with a dollar sign, so no method collides.
func phpxStaticKey(name string) string {
	return "$" + name
}

// phpxShareStatics records, for each static a class inherits without
// redeclaring, the class that holds its storage. Called before the merge, so
// c.Statics still lists the class's own and trait-supplied declarations.
func phpxShareStatics(c *model.Class, parent *model.Class) {
	for _, field := range parent.Statics {
		if phpxHasField(c.Statics, field.Name) {
			continue
		}
		owner := parent.MethodOwner[phpxStaticKey(field.Name)]
		if owner == "" {
			owner = parent.Name
		}
		c.MethodOwner[phpxStaticKey(field.Name)] = owner
	}
}

// staticStorageFor resolves `Class::$name` to the bag holding that property.
// A property the class inherited lives in the bag of the class that declared
// it, which is php's shared storage; a redeclaration keeps its own.
func (rt *Runtime) staticStorageFor(class, name string, scope *Scope) (map[string]any, error) {
	// `static::$x` names the called class, which may redeclare the property;
	// the owner lookup below sends an inherited one to the class that holds it.
	resolved, err := resolveCalledClass(class, scope)
	if err != nil {
		return nil, err
	}
	bag, err := rt.staticStorage(resolved, scope)
	if err != nil {
		return nil, err
	}
	decl, ok := rt.lookupClass(resolved)
	if !ok {
		return bag, nil
	}
	owner := decl.MethodOwner[phpxStaticKey(name)]
	if owner == "" || strings.EqualFold(owner, decl.Name) {
		return bag, nil
	}
	return rt.staticStorage(owner, scope)
}

// phpxHasField reports whether a field list declares the name.
func phpxHasField(fields []model.Field, name string) bool {
	for _, field := range fields {
		if strings.EqualFold(field.Name, name) {
			return true
		}
	}
	return false
}
