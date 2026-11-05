//go:build phpx

package phpx

import (
	"os"
	"os/user"
	"strconv"

	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\SAPIStubs`, registerSapiStubs)
}

// registerSapiStubs installs the process and SAPI functions a CLI run answers
// with a constant or a syscall: header handling that has no response to touch,
// flush, and the owner of the process.
func registerSapiStubs(rt *runner.Runtime) {
	rt.RegisterFunc("header_remove", func(_ ...any) {})
	rt.RegisterFunc("headers_list", func() []string { return []string{} })
	rt.RegisterFunc("flush", func() {})
	rt.RegisterFunc("ob_implicit_flush", func(_ ...any) {})
	rt.RegisterFunc("connection_status", func() int64 { return 0 })
	rt.RegisterFunc("getmyuid", func() int64 { return int64(os.Getuid()) })
	rt.RegisterFunc("get_current_user", func() string {
		if u, err := user.Current(); err == nil {
			return u.Username
		}
		return strconv.Itoa(os.Getuid())
	})
	rt.RegisterFunc("get_cfg_var", func(_ string) any { return false })
	rt.RegisterFunc("ini_restore", func(_ string) {})
	rt.RegisterFunc("fastcgi_finish_request", func() bool { return false })
	rt.RegisterFunc("disk_free_space", func(_ string) any { return false })
}
