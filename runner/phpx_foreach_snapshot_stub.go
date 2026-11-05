//go:build !phpx

package runner

import "github.com/titpetric/phpx/model"

// phpxSnapshotSource answers the untagged build, which walks the live array.
func phpxSnapshotSource(_ *model.Foreach) bool { return false }

// phpxSnapshot answers the untagged build; it is never reached.
func phpxSnapshot(src *model.Array) *model.Array { return src }
