//go:build phpx

package runner

import "github.com/titpetric/phpx/model"

// phpxDeclareFunc declares a function when its statement runs. hoist
// registers the declarations at the top of a file; one inside an `if`, a
// block or another function body is declared by php when control reaches it,
// which is how `if (!function_exists('f')) { function f() {} }` guards a
// polyfill. A name already registered is left alone, because hoist put the
// top-level declaration there and a guard that lost the race declares nothing.
func (rt *Runtime) phpxDeclareFunc(fd *model.FuncDecl) error {
	if fd.Class != "" {
		return nil
	}
	if _, ok := rt.lookupFunc(fd.Name); ok {
		return nil
	}
	if err := rt.declareFunc(fd.Name, FuncSite{File: fd.Filename}); err != nil {
		return err
	}
	rt.registerUserFunc(fd.Name, func(args ...any) (any, error) {
		return rt.invokeFunc(fd, args)
	})
	return nil
}
