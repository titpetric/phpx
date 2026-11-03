//go:build phpx

package core

import "github.com/titpetric/phpx/runner"

func init() {
	runner.RegisterBinding(registerPhpxCalled)
}

// registerPhpxCalled installs get_called_class(), the runtime form of `static::class`.
func registerPhpxCalled(rt *runner.Runtime) {
	rt.RegisterFunc("get_called_class", func() any {
		return rt.PhpxCalledClass()
	})
}
