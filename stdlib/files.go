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
	if selected(rt, `Runtime\PS\Files`) {
		files.RegisterRoot(rt, dir)
	}
	if selected(rt, `Runtime\PS\GD`) {
		gd.RegisterRoot(rt, dir)
	}
	if selected(rt, `Runtime\PS\Auth`) {
		auth.RegisterRoot(rt, dir)
	}
	if selected(rt, `Runtime\PS\Session`) {
		session.RegisterRoot(rt, dir)
	}
	if rt.Profile()&Exec != 0 && selected(rt, `Runtime\PS\Exec`) {
		pexec.RegisterRoot(rt, dir)
	}
	if selected(rt, `Runtime\PX\Late`) {
		runner.PhpxRunLate(rt)
	}
}

// selected reports whether the runtime's areas include the named one: a runtime
// built with no selection carries every area, and one built with a selection
// reroots only the bindings of the areas it installed.
func selected(rt *runner.Runtime, area string) bool {
	return len(runner.SelectAreas([]runner.Area{{Name: area}}, rt.Areas())) == 1
}
