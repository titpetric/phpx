//go:build phpx

package runner

import (
	"fmt"
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxHostThrowableIs answers `$e instanceof Name` for a throwable a Go binding built.
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
	if _, ok := value.(PhpxHostIterator); ok {
		switch strings.ToLower(want) {
		case "traversable", "iterator", "generator":
			return true
		}
	}
	if name := fmt.Sprintf("%T", value); strings.HasPrefix(name, "*phpx.DateTime") && strings.EqualFold(want, "DateTimeInterface") {
		return strings.HasSuffix(name, "DateTime") || strings.HasSuffix(name, "DateTimeImmutable")
	}
	err, ok := value.(error)
	if !ok {
		return false
	}
	// A go error no class names is still catchable, and matchCatchType takes
	// it for Throwable, so instanceof must agree with the catch.
	if strings.EqualFold(want, "Throwable") {
		return true
	}
	class, ok := throwableClassOf(err)
	if !ok {
		return false
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
