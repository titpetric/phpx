//go:build phpx

package runner

import (
	"github.com/titpetric/phpx/modelutil"

	"github.com/titpetric/phpx/model"
)

// phpxSnapshotSource reports whether a by-value foreach walks a copy of its
// array. php iterates the array as it was when the loop began: a body that
// writes the source, unsets from it or calls something that appends to it
// leaves the walk alone. A source that is an expression (a call, a literal)
// belongs to the loop and no other handle can reach it.
func phpxSnapshotSource(n *model.Foreach) bool {
	if n.ByRef {
		return false
	}
	switch modelutil.UnwrapParenthesized(n.Source).(type) {
	case *model.Var, *model.PropAccess, *model.StaticProp, *model.Index:
		return true
	}
	return false
}

// phpxSnapshot returns the copy the loop walks: a second handle over the same
// storage, split by the first write to either side. Walking the handle costs
// one allocation, where copying the entries cost one per entry.
func phpxSnapshot(src *model.Array) *model.Array {
	return src.Clone()
}
