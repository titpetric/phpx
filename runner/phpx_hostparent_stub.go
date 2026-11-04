//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxHostConstruct builds no host parent untagged.
func (rt *Runtime) phpxHostConstruct(_, _ string, _ []any, _ *Scope) (bool, error) { return false, nil }

// phpxHostValue finds no host parent untagged.
func phpxHostValue(_ *model.Object) (any, bool) { return nil, false }
