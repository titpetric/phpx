package runner

import (
	"strings"
)

// calledKey is the frame slot holding the class a call was made on, the class
// `static::` names. `__class__` is the class that wrote the running body, which
// is what `self::` names; the two differ once a child inherits a method.
const calledKey = "__static__"

// phpxSetCalled records the called class on a method frame.
func phpxSetCalled(scope *Scope, class string) {
	if class != "" {
		scope.Set(calledKey, class)
	}
}

// phpxCalled reads the called class of the running frame.
func phpxCalled(scope *Scope) (string, bool) {
	if scope == nil {
		return "", false
	}
	value, ok := scope.Get(calledKey)
	if !ok {
		return "", false
	}
	name, ok := value.(string)
	return name, ok && name != ""
}

// resolveCalledClass is resolveClassName with late static binding: `static`
// names the called class. Static properties keep resolveClassName, because
// php shares an inherited static between parent and child.
func resolveCalledClass(class string, scope *Scope) (string, error) {
	if strings.EqualFold(class, "static") {
		if name, ok := phpxCalled(scope); ok {
			return name, nil
		}
	}
	return resolveClassName(class, scope)
}

// phpxForwardedClass is the class a static call runs as. `self::`, `parent::`
// and `static::` forward the caller's called class; a named class does not.
func phpxForwardedClass(keyword, resolved string, scope *Scope) string {
	switch strings.ToLower(keyword) {
	case "self", "parent", "static":
		if name, ok := phpxCalled(scope); ok {
			return name
		}
	}
	return resolved
}

// PhpxCalledClass answers get_called_class() for the frame that called it.
func (rt *Runtime) PhpxCalledClass() any {
	if len(rt.frames) == 0 {
		return false
	}
	if name, ok := phpxCalled(rt.frames[len(rt.frames)-1]); ok {
		return name
	}
	return false
}
