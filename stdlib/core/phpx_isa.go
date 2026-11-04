//go:build phpx

package core

import (
	"io"

	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxIsA)
}

// registerPhpxIsA installs is_a(), which upstream leaves out while
// is_subclass_of() is there. Framework code guards an object with it.
func registerPhpxIsA(rt *runner.Runtime) {
	rt.RegisterFunc("is_a", func(value any, class string, allowString ...bool) bool {
		return rt.PhpxIsA(value, class, len(allowString) > 0 && allowString[0])
	})
	// gettype names a stream "resource", as php does, where upstream's answer
	// for a Go reader is "unknown type".
	rt.RegisterFunc("gettype", func(value any) string {
		_, reads := value.(io.Reader)
		_, writes := value.(io.Writer)
		if reads || writes {
			return "resource"
		}
		return phpGetType(value)
	})
}
