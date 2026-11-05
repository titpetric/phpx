//go:build !phpx

package compat

import (
	"github.com/titpetric/phpx/runner"
)

// phpxBadUTF8 never refuses a subject untagged.
func (c *regexpCache) phpxBadUTF8(_, _ string) bool { return false }

// phpxRegisterErrors adds nothing untagged.
func (c *regexpCache) phpxRegisterErrors(_ *runner.Runtime) {}

// phpxUCP leaves the pattern as it came untagged.
func phpxUCP(body string) (string, bool) { return body, false }
