package runner

import (
	"reflect"
	"strings"

	"github.com/titpetric/phpx/model"
)

// PropertyExists reports whether a value holds the property a script names:
// an interpreted object's own properties and the declarations that outlive an
// unset, a class name's declarations, and the exported Go fields of a value a
// binding returned, resolved through structField the way a property read
// resolves them.
func (rt *Runtime) PropertyExists(value any, name string) bool {
	if object, ok := value.(*model.Object); ok {
		if _, set := object.Prop(name); set {
			return true
		}
		return classDeclares(object.Class, name)
	}
	// php reads a class name where an object goes and autoloads the name:
	// `property_exists(static::class, 'key')` is how livewire finds a
	// synthesizer's key.
	if className, ok := value.(string); ok {
		className = strings.TrimPrefix(className, "\\")
		if loaded, _ := rt.ClassExists(className, true); !loaded {
			return false
		}
		class, found := rt.lookupClass(className)
		return found && classDeclares(class, name)
	}

	t := reflect.TypeOf(value)
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return false
	}
	_, found := rt.structField(t, name)
	return found
}

// classDeclares reports whether the class declares the property, an instance
// one or a static one, whatever its visibility.
func classDeclares(class *model.Class, name string) bool {
	if class == nil {
		return false
	}
	if _, declared := class.Field(name); declared {
		return true
	}
	for _, field := range class.Statics {
		if field.Name == name {
			return true
		}
	}
	return false
}
