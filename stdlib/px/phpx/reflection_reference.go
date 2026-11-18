//go:build phpx

package phpx

import (
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

// ReflectionReference::fromArrayElement is what a library reads to tell a
// reference in an array from a value: sentry's data filter and symfony's var
// dumper both call it. The runtime's arrays keep no reference element, so the
// answer is null for an element that exists, which is the value-only case php
// answers too.
func init() {
	runner.PhpxRegisterLate(registerReflectionReference)
}

func registerReflectionReference(rt *runner.Runtime) {
	rt.RegisterFunc("ReflectionReference::fromArrayElement", func(array any, key any) any {
		values, ok := array.(*model.Array)
		if !ok {
			return nil
		}
		normalized := normalizeArrayKey(key)
		if _, found := values.Get(normalized); !found {
			return &reflectionError{"ReflectionReference::fromArrayElement(): Array key not found"}
		}
		return nil
	})
}

// normalizeArrayKey reads a key the way an array lookup does: a whole float
// truncates and a bool becomes 0 or 1.
func normalizeArrayKey(key any) any {
	switch k := key.(type) {
	case int:
		return int64(k)
	case float64:
		return int64(k)
	case bool:
		if k {
			return int64(1)
		}
		return int64(0)
	}
	return key
}

// reflectionError is a ReflectionException with the message php prints.
type reflectionError struct{ message string }

func (e *reflectionError) Error() string          { return e.message }
func (e *reflectionError) ThrowableClass() string { return "ReflectionException" }
