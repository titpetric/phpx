//go:build phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxMayReach(t *testing.T) {
	pure := []model.Stmt{&model.Echo{Args: []model.Expr{&model.Var{Name: "x"}}}}
	if phpxMayReach(pure) {
		t.Fatal("an echo of a variable reaches nothing")
	}
	call := []model.Stmt{&model.ExprStmt{X: &model.MethodCall{Base: &model.Var{Name: "o"}, Method: "m"}}}
	if !phpxMayReach(call) {
		t.Fatal("a method call can reach the array")
	}
}

func TestPhpxSnapshotIsIndependent(t *testing.T) {
	a := model.NewArray()
	a.Append(1)
	s := phpxSnapshot(a)
	a.Append(2)
	if s.Len() != 1 {
		t.Fatal("the snapshot followed the source")
	}
}
