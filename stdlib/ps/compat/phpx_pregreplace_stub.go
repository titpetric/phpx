//go:build !phpx

package compat

import (
	"github.com/titpetric/phpx/runner"
)

// phpxRegisterReplace adds nothing untagged: upstream's preg_replace takes a
// pattern and a subject that are strings.
func (c *regexpCache) phpxRegisterReplace(_ *runner.Runtime) {}
