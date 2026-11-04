//go:build phpx

package phpx

import (
	"io"
	"math"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxTypes)
}

// registerPhpxTypes installs the type predicates and conversions vendor code
// reaches for, each with one argument.
func registerPhpxTypes(rt *runner.Runtime) {
	rt.RegisterFunc("is_scalar", func(value any) bool {
		switch value.(type) {
		case int, int64, float64, string, bool:
			return true
		}
		return false
	})
	rt.RegisterFunc("is_resource", func(value any) bool {
		switch value.(type) {
		case nil, *model.Object, *model.Array, string, bool, int, int64, float64:
			return false
		}
		_, reads := value.(io.Reader)
		_, writes := value.(io.Writer)
		return reads || writes
	})
	rt.RegisterFunc("is_iterable", func(value any) bool {
		if model.IsCollection(value) {
			return true
		}
		switch value.(type) {
		case nil, string, bool, int, int64, float64:
			return false
		}
		return rt.PhpxIsA(value, "Traversable", false)
	})
	rt.RegisterFunc("is_countable", func(value any) bool {
		// A script object that implements Countable counts too.
		if obj, ok := value.(*model.Object); ok && obj.Class != nil {
			for _, name := range obj.Class.Implements {
				if name == "countable" {
					return true
				}
			}
		}
		return model.IsCollection(value)
	})
	rt.RegisterFunc("boolval", func(value any) bool { return phpval.Truthy(value) })
	rt.RegisterFunc("strval", func(value any) string { return phpval.String(value) })
	// is_numeric: upstream answers false for a numeric string.
	rt.RegisterFunc("is_numeric", func(value any) bool {
		switch v := value.(type) {
		case int64, int, float64:
			return true
		case string:
			return phpval.IsNumericString(v)
		}
		return false
	})
	rt.RegisterFunc("is_nan", func(value float64) bool { return math.IsNaN(value) })
	rt.RegisterFunc("is_finite", func(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) })
	rt.RegisterFunc("get_debug_type", func(value any) string { return debugType(value) })
}

// debugType names a value as get_debug_type() does: the scalar names php 8
// settled on, and a class name for an object.
func debugType(value any) string {
	switch v := value.(type) {
	case nil:
		return "null"
	case bool:
		return "bool"
	case int, int64:
		return "int"
	case float64:
		return "float"
	case string:
		return "string"
	case *model.Array:
		return "array"
	case *model.Object:
		if v.Class != nil {
			return v.Class.Name
		}
		return "object"
	}
	if model.IsCollection(value) {
		return "array"
	}
	if is, _ := value.(io.Reader); is != nil {
		return "resource (stream)"
	}
	return "object"
}
