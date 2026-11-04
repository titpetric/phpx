//go:build !phpx

package runner

import "github.com/titpetric/phpx/model"

// phpxPrelude declares nothing untagged.
func (rt *Runtime) phpxPrelude(_ string) bool { return false }

// phpxRecordFunc keeps no declaration untagged.
func phpxRecordFunc(_ *model.FuncDecl) {}
