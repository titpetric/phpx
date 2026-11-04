//go:build !phpx

package runner

import "github.com/titpetric/phpx/model"

// phpxStampThrowable answers the untagged build, which records no creation
// site on a throwable.
func (rt *Runtime) phpxStampThrowable(_ *model.Object, _ *Scope) {}
