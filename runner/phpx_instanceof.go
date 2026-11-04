//go:build phpx

package runner

import (
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxHostThrowableIs answers `$e instanceof Name` for a throwable a Go binding
// built: the class it carries, every ancestor php gives it, and Throwable.
func phpxHostThrowableIs(value any, want string) bool {
	if obj, ok := value.(*model.Object); ok && obj.Class != nil && strings.EqualFold(want, "Traversable") {
		// Iterator and IteratorAggregate both extend Traversable.
		for _, name := range obj.Class.Implements {
			if strings.EqualFold(name, "iterator") || strings.EqualFold(name, "iteratoraggregate") ||
				strings.EqualFold(name, "recursiveiterator") || strings.EqualFold(name, "outeriterator") {
				return true
			}
		}
		return false
	}
	err, ok := value.(error)
	if !ok {
		return false
	}
	class, ok := throwableClassOf(err)
	if !ok {
		return false
	}
	if strings.EqualFold(want, "Throwable") {
		return true
	}
	for class != "" {
		if strings.EqualFold(want, class) {
			return true
		}
		next := ""
		for _, pair := range phpxThrowableParents {
			if strings.EqualFold(pair[0], class) {
				next = pair[1]
				break
			}
		}
		class = next
	}
	return false
}
