//go:build phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxPromote(t *testing.T) {
	ctor := &model.FuncDecl{Name: "__construct", Params: []model.Param{
		{Name: "a", Modifiers: "private readonly"},
		{Name: "b"},
		{Name: "c", Modifiers: "readonly", Type: "int"},
	}}
	cd := &model.ClassDecl{Name: "A", Methods: []*model.FuncDecl{ctor}}
	got := phpxPromote(cd)
	if len(cd.Fields) != 0 || len(ctor.Body) != 0 {
		t.Fatal("the declaration the formatter prints back was changed")
	}
	if len(got.Fields) != 2 || got.Fields[0].Visibility != "private" || got.Fields[1].Visibility != "public" || got.Fields[1].Type != "int" {
		t.Fatalf("fields = %+v", got.Fields)
	}
	if len(got.Methods[0].Body) != 2 {
		t.Fatalf("body = %d statements, want 2 assignments", len(got.Methods[0].Body))
	}
	plain := &model.ClassDecl{Name: "B", Methods: []*model.FuncDecl{{Name: "__construct", Params: []model.Param{{Name: "x"}}}}}
	if phpxPromote(plain) != plain {
		t.Fatal("a constructor with no modifier must come back untouched")
	}
}
