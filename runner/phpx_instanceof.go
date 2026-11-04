//go:build phpx

package runner

import "strings"

// phpxHostThrowableIs answers `$e instanceof Name` for a throwable a Go binding
// built: the class it carries, every ancestor php gives it, and Throwable.
func phpxHostThrowableIs(value any, want string) bool {
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
