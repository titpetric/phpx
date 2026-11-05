package expr

import (
	"errors"
	"github.com/titpetric/phpx/model"
	"strings"
)

// isQuietCall names the language constructs that answer for a value out of
// reach and never raise: isset and empty.
func isQuietCall(name string) bool {
	if i := strings.LastIndexByte(name, '\\'); i >= 0 {
		name = name[i+1:]
	}
	return strings.EqualFold(name, "isset") || strings.EqualFold(name, "empty")
}

// quietOnRefusal turns an error that marks itself Quiet, the refusal of a
// protected or private property, into null: isset() is false and empty() is
// true, as in php. Any other error passes through.
func quietOnRefusal(c closure) closure {
	return func(env *Env) (any, error) {
		v, err := c(env)
		var quiet interface{ Quiet() bool }
		if err != nil && errors.As(err, &quiet) && quiet.Quiet() {
			return nil, nil
		}
		return v, err
	}
}

// phpxQuietMagic brackets the evaluation of an isset() or empty() argument so a
// property read asks __isset before __get.
func phpxQuietMagic(c closure, name string, arg model.Expr) closure {
	// isset($o->p) is answered by __isset alone; empty() and a nested
	// $a->b->c read the value too.
	prop, isProp := model.UnwrapParenthesized(arg).(*model.PropAccess)
	final := isProp && strings.EqualFold(name[strings.LastIndexByte(name, '\\')+1:], "isset")
	if final {
		_, nested := model.UnwrapParenthesized(prop.Base).(*model.PropAccess)
		final = !nested
	}
	return func(env *Env) (any, error) {
		enter, ok := env.Base["__phpx_quiet"].(func(...any) (any, error))
		if !ok {
			return c(env)
		}
		_, _ = enter(true, final)
		defer func() { _, _ = enter(false, false) }()
		return c(env)
	}
}
