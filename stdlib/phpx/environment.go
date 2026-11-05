//go:build phpx

package phpx

import (
	"os"

	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxEnvironment)
}

// registerPhpxEnvironment installs the functions that read the host and the
// last-error state of the json and pcre bindings. Those bindings raise on
// failure, so the state they leave is always "no error".
func registerPhpxEnvironment(rt *runner.Runtime) {
	// PHP_BINARY is the interpreter a script re-launches: the running binary,
	// which answers `run` the way php answers a script.
	if exe, err := os.Executable(); err == nil {
		rt.SetConst("PHP_BINARY", exe)
	}
	rt.RegisterFunc("gethostname", func() any {
		name, err := os.Hostname()
		if err != nil {
			return false
		}
		return name
	})
	rt.RegisterFunc("sys_get_temp_dir", func() string { return os.TempDir() })
}
