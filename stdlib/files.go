package stdlib

import (
	"github.com/titpetric/phpx/runner"
	"github.com/titpetric/phpx/stdlib/auth"
	"github.com/titpetric/phpx/stdlib/files"
	"github.com/titpetric/phpx/stdlib/gd"
	"github.com/titpetric/phpx/stdlib/pexec"
	"github.com/titpetric/phpx/stdlib/session"
)

// RegisterFS reroots the filesystem bindings Register installed at dir, which
// a path from PHP resolves under and cannot climb out of. gd, auth and session
// go with them, and pexec where the mount included it: the root is where a
// command starts, so that getcwd() agrees with it.
func RegisterFS(rt *runner.Runtime, dir string) {
	files.RegisterRoot(rt, dir)
	gd.RegisterRoot(rt, dir)
	auth.RegisterRoot(rt, dir)
	session.RegisterRoot(rt, dir)
	if rt.Profile()&Exec != 0 {
		pexec.RegisterRoot(rt, dir)
	}
}
