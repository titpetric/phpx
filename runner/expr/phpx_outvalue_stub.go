//go:build !phpx

package expr

import (
	"github.com/titpetric/phpx/model"
)

// phpxOutValueArgs answers the untagged build, whose calls pass a by-reference
// parameter alone.
func (dc *directCompiler) phpxOutValueArgs(_ *model.Call, args []closure) ([]closure, error) {
	return args, nil
}
