//go:build phpx

package runner

// PhpxRebind answers a copy of a closure bound to newThis and, when scope is
// not empty, to the class named scope, which is how composer's loader reads a
// private property from `Closure::bind(fn, null, ClassLoader::class)`. A value
// that is not a closure literal answers false.
func (rt *Runtime) PhpxRebind(closure any, newThis any, scope string) (any, bool) {
	c, ok := closure.(*Callable)
	if !ok || c.closure == nil {
		return nil, false
	}
	env := c.env
	env.this = newThis
	if scope != "" {
		env.class = scope
		env.called = scope
	}
	return rt.newClosure(c.closure, env), true
}
