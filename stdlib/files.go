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
	if selected(rt, fsAreas[0]) {
		files.RegisterRoot(rt, dir)
	}
	if selected(rt, fsAreas[1]) {
		gd.RegisterRoot(rt, dir)
	}
	if selected(rt, fsAreas[2]) {
		auth.RegisterRoot(rt, dir)
	}
	if selected(rt, fsAreas[3]) {
		session.RegisterRoot(rt, dir)
	}
	if rt.Profile()&Exec != 0 && selected(rt, fsAreas[4]) {
		pexec.RegisterRoot(rt, dir)
	}
	if selected(rt, runner.NamespaceLate) || selectsOverlay(rt) {
		runner.PhpxRunLate(rt)
	}
}

// fsAreas are the areas RegisterFS reroots, each the name its package registers under.
var fsAreas = []string{`Runtime\PS\Files`, `Runtime\PS\GD`, `Runtime\PS\Auth`, `Runtime\PS\Session`, `Runtime\PS\Exec`}

// selectsOverlay reports whether the runtime's selection names an area of the
// overlay: the late installers belong to it and run once any of it is installed.
func selectsOverlay(rt *runner.Runtime) bool {
	for _, ns := range rt.Areas() {
		if runner.InNamespace(ns, runner.NamespacePX) {
			return true
		}
	}
	return false
}

// selected reports whether the runtime's areas include the named one: a runtime
// built with no selection carries every area, and one built with a selection
// reroots only the bindings of the areas it installed.
func selected(rt *runner.Runtime, area string) bool {
	return len(runner.SelectAreas([]runner.Area{{Name: area}}, rt.Areas())) == 1
}

// FSAreas lists the areas RegisterFS reroots.
func FSAreas() []string { return append([]string(nil), fsAreas...) }
