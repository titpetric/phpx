//go:build !phpx

package compat

import "github.com/titpetric/phpx/runner"

// phpxRegisterBuffers adds nothing to the untagged build.
func phpxRegisterBuffers(*runner.Runtime, *buffers) {}
