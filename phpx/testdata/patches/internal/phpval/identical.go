package phpval

import (
	"reflect"

	"github.com/titpetric/phpx/model"
)

// This file is PHP's === over the values a script holds. Compare answers ==,
// which orders and coerces; identity answers the type and the value alone, and
// an array identity is a walk of the pair, never a walk of the Go value: a
// *model.Array carries fields a copy does not share, and comparing those
// answers false for two arrays PHP calls identical.

// Identical answers PHP's ===.
func Identical(a, b any) bool {
	switch x := a.(type) {
	case nil:
		return b == nil
	case bool:
		y, ok := b.(bool)
		return ok && x == y
	case string:
		if y, ok := Bytes(b); ok {
			return x == y
		}
		y, ok := b.(string)
		return ok && x == y
	case []byte:
		// A []byte is the string it holds, and === against one compares text
		// and never reports two different types.
		if y, ok := Bytes(b); ok {
			return string(x) == y
		}
		y, ok := b.(string)
		return ok && string(x) == y
	case int, int64:
		switch b.(type) {
		case int, int64:
			return Int(a) == Int(b)
		}
		return false
	case float64:
		y, ok := b.(float64)
		return ok && x == y
	case *model.Array:
		y, ok := b.(*model.Array)
		return ok && IdenticalArrays(x, y)
	case *model.Object:
		y, ok := b.(*model.Object)
		return ok && x == y
	}
	if b == nil {
		return false
	}
	// A binding's []string is not comparable and would panic under ==.
	return reflect.TypeOf(a) == reflect.TypeOf(b) && reflect.DeepEqual(a, b)
}

// IdenticalArrays reports the same key/value pairs in the same order, compared
// identically.
func IdenticalArrays(x, y *model.Array) bool {
	if x == y {
		return true
	}
	if x == nil || y == nil || x.Len() != y.Len() {
		return false
	}
	xk, yk := x.Keys(), y.Keys()
	for i := range xk {
		if !Identical(xk[i], yk[i]) {
			return false
		}
		xv, _ := x.Get(xk[i])
		yv, _ := y.Get(yk[i])
		if !Identical(xv, yv) {
			return false
		}
	}
	return true
}

// RecursiveDependencyError is the Error php raises when a comparison reaches a
// container it is already inside: "Nesting level too deep - recursive
// dependency?". It is thrown, not returned, because Compare answers an int.
type RecursiveDependencyError struct{}

// Error answers php's wording.
func (e *RecursiveDependencyError) Error() string {
	return "Nesting level too deep - recursive dependency?"
}

// ThrowableClass names the PHP class the comparison raises.
func (e *RecursiveDependencyError) ThrowableClass() string { return "Error" }
