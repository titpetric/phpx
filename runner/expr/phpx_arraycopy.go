package expr

import "github.com/titpetric/phpx/model"

// phpxElementReadsStorage reports whether an array literal's element reads an
// array out of storage that stays reachable: a variable, a property, a static
// property or an element. php copies it into the literal, and a handle shared
// there is written through later (a parser that snapshots `$this->stack` into
// an array and restores it).
func phpxElementReadsStorage(e model.Expr) bool {
	switch n := model.UnwrapParenthesized(e).(type) {
	case *model.Var:
		return !n.Const
	case *model.PropAccess, *model.StaticProp, *model.Index:
		return true
	case *model.Ternary:
		return phpxElementReadsStorage(n.Then) || phpxElementReadsStorage(n.Else)
	}
	return false
}

// phpxCopyElement wraps an element closure so an array read from storage is
// copied into the literal.
func phpxCopyElement(e model.Expr, c closure) closure {
	if !phpxElementReadsStorage(e) {
		return c
	}
	return func(env *Env) (any, error) {
		v, err := c(env)
		if arr, ok := v.(*model.Array); ok && err == nil {
			return model.CopyArray(arr), nil
		}
		return v, err
	}
}
