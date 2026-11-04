//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxRecordFunc keeps no declaration untagged.
func phpxRecordFunc(_ *model.FuncDecl) {}

// phpxRememberTrait keeps nothing untagged.
func phpxRememberTrait(_ *model.TraitDecl) {}
