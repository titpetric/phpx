//go:build !phpx

package runner

import "github.com/titpetric/phpx/model"

// phpxRefAssign answers the untagged build, where `=&` binds by value.
func (rt *Runtime) phpxRefAssign(_, _ model.Expr, _ *Scope) (any, error) { return nil, nil }

// phpxInstallRefAssign answers the untagged build, which has no such helper.
func (rt *Runtime) phpxInstallRefAssign(_ map[string]any, _ *scopeRef) {}

// phpxIsRefAssign answers the untagged build: no assignment is by reference.
func phpxIsRefAssign(_ string, _ model.Expr) (model.Expr, bool) { return nil, false }

// phpxInstallRefArg answers the untagged build, which has no such helper.
func (rt *Runtime) phpxInstallRefArg(_ map[string]any, _ *scopeRef) {}

// phpxInstallVarVar answers the untagged build, which has no variable variables.
func (rt *Runtime) phpxInstallVarVar(_ map[string]any, _ *scopeRef) {}
