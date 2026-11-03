//go:build phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxShareStatics pins which inherited statics point at the declarer: one
// the child redeclares keeps its own storage, a grandchild names the root.
func TestPhpxShareStatics(t *testing.T) {
	t.Parallel()
	parent := &model.Class{
		Name:        "A",
		Statics:     []model.Field{{Name: "n"}, {Name: "r"}},
		MethodOwner: map[string]string{},
	}
	child := &model.Class{Name: "B", Statics: []model.Field{{Name: "r"}}, MethodOwner: map[string]string{}}
	phpxShareStatics(child, parent)
	if got := child.MethodOwner["$n"]; got != "A" {
		t.Errorf("inherited owner = %q, want A", got)
	}
	if _, ok := child.MethodOwner["$r"]; ok {
		t.Errorf("a redeclared static must keep its own storage")
	}
	grand := &model.Class{Name: "C", MethodOwner: map[string]string{}}
	grand.MethodOwner = map[string]string{}
	phpxShareStatics(grand, &model.Class{Name: "B", Statics: parent.Statics, MethodOwner: child.MethodOwner})
	if got := grand.MethodOwner["$n"]; got != "A" {
		t.Errorf("grandchild owner = %q, want A", got)
	}
}
