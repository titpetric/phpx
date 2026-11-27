//go:build phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxMethodKey names one memoized lookup: a class and the method spelling.
type phpxMethodKey struct {
	class  *model.Class
	method string
}

// phpxMethodHit is what one lookup answered, nil for a name no class declares.
type phpxMethodHit struct {
	decl *model.FuncDecl
	ok   bool
}

// phpxMethod answers lookupPHPMethod once per class and spelling, never scanning twice.
func (rt *Runtime) phpxMethod(class *model.Class, method string) (*model.FuncDecl, bool) {
	if class == nil {
		return nil, false
	}
	if rt.phpxMethods == nil {
		rt.phpxMethods = make(map[phpxMethodKey]phpxMethodHit)
	}
	key := phpxMethodKey{class: class, method: method}
	if hit, ok := rt.phpxMethods[key]; ok {
		return hit.decl, hit.ok
	}
	decl, ok := lookupPHPMethod(class, method)
	rt.phpxMethods[key] = phpxMethodHit{decl: decl, ok: ok}
	return decl, ok
}
