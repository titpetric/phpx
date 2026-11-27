//go:build phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxSnapshotSource(t *testing.T) {
	named := &model.Foreach{Source: &model.Var{Name: "a"}}
	if !phpxSnapshotSource(named) {
		t.Fatal("a variable source walks a copy")
	}
	byRef := &model.Foreach{Source: &model.Var{Name: "a"}, ByRef: true}
	if phpxSnapshotSource(byRef) {
		t.Fatal("a by-reference loop walks the live array")
	}
	call := &model.Foreach{Source: &model.Call{Name: "f"}}
	if phpxSnapshotSource(call) {
		t.Fatal("a call's array belongs to the loop")
	}
}

func TestPhpxSnapshotIsIndependent(t *testing.T) {
	a := model.NewArray()
	a.Append(int64(1))
	s := phpxSnapshot(a)
	a.Append(2)
	if s.Len() != 1 {
		t.Fatal("the snapshot followed the source")
	}
	a.Set(int64(0), "changed")
	if v, _ := s.Get(int64(0)); v != int64(1) {
		t.Fatalf("snapshot element = %v, want 1", v)
	}
}
