//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxHostConstruct builds no host parent untagged.
func (rt *Runtime) phpxHostConstruct(_, _ string, _ []any, _ *Scope) (bool, error) { return false, nil }

// phpxHostValue finds no host parent untagged.
func phpxHostValue(_ *model.Object) (any, bool) { return nil, false }

// phpxHostInherit builds nothing untagged: a class extends no host class.
func (rt *Runtime) phpxHostInherit(_ *model.Object, _ *model.Class, _ []any, _ *Scope) error {
	return nil
}

// phpxWrapHostStatic answers the value as it came untagged.
func (rt *Runtime) phpxWrapHostStatic(_, _ string, out any, _ *Scope) any { return out }

// phpxHostStaticOf finds nothing untagged.
func (rt *Runtime) phpxHostStaticOf(_ *model.Class, _, _ string, _ []any, _ *Scope) (any, bool, error) {
	return nil, false, nil
}

// phpxHostCall calls the method on the value untagged.
func (rt *Runtime) phpxHostCall(_ *model.Object, host any, method string, args []any, scope *Scope) (any, error) {
	return rt.callGoMethod(host, method, args, func() *Scope { return scope })
}

// phpxHostParentCall handles nothing untagged.
func (rt *Runtime) phpxHostParentCall(_, _ string, _ []any, _ *Scope) (any, bool, error) {
	return nil, false, nil
}

// phpxHostHasMethod is true untagged, where a host value is called as it was.
func phpxHostHasMethod(_ any, _ string) bool { return true }

// phpxHostArg finds no host value untagged.
func phpxHostArg(_ any) (any, bool) { return nil, false }

// phpxHostClassConst finds nothing untagged.
func (rt *Runtime) phpxHostClassConst(_, _ string, _ *scopeRef) (any, bool) { return nil, false }

// PhpxUnwrapHost answers the value as it is untagged.
func PhpxUnwrapHost(v any) any { return v }
