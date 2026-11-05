//go:build phpx

package phpx

import (
	"os"
	"syscall"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerTmpfile)
}

// registerTmpfile installs tmpfile() and umask(). The temporary file is
// unlinked as soon as it is open, so it is gone when the handle closes.
func registerTmpfile(rt *runner.Runtime) {
	rt.RegisterFunc("tmpfile", func() any {
		f, err := os.CreateTemp("", "php")
		if err != nil {
			return false
		}
		_ = os.Remove(f.Name())
		return f
	})
	rt.RegisterFunc("umask", func(mask ...any) int64 {
		old := syscall.Umask(0)
		if len(mask) == 0 || mask[0] == nil {
			syscall.Umask(old)
		} else {
			syscall.Umask(int(phpval.Int(mask[0])) & 0o777)
		}
		return int64(old)
	})
}
