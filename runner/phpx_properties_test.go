//go:build phpx

package runner

import (
	"io"
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxPropVerdict walks the access table: public always, protected for the
// family, private for the declaring class alone, and a private of an ancestor
// reads as no property.
func TestPhpxPropVerdict(t *testing.T) {
	rt := New(io.Discard, Options{})
	a := &model.Class{Name: "A", Fields: []model.Field{{Name: "p", Visibility: "protected"}, {Name: "q", Visibility: "private"}, {Name: "r"}}}
	phpxBindProps(a, a.Fields, nil)
	b := &model.Class{Name: "B", Ancestors: []string{"A"}, Fields: a.Fields}
	phpxBindProps(b, nil, a)
	rt.classes = map[string]*model.Class{"a": a, "b": b}
	scopeOf := func(class string) *Scope {
		s := NewScope()
		s.Set("__class__", class)
		return s
	}
	obj := &model.Object{Class: b}
	for _, tc := range []struct {
		name, from string
		hidden     bool
		refused    bool
	}{
		{"r", "", false, false},
		{"p", "", false, true},
		{"p", "B", false, false},
		{"p", "A", false, false},
		{"q", "A", false, false},
		{"q", "B", true, false},
		{"q", "", true, false},
	} {
		hidden, err := rt.phpxPropVerdict(obj, tc.name, scopeOf(tc.from))
		if hidden != tc.hidden || (err != nil) != tc.refused {
			t.Errorf("%s from %q: hidden %v, err %v", tc.name, tc.from, hidden, err)
		}
	}
}

// TestPhpxReadonlyWrite walks the readonly table: a value already held is
// never replaced, and an empty slot takes its first write from the declaring
// class alone.
func TestPhpxReadonlyWrite(t *testing.T) {
	c := &model.Class{Name: "R", Fields: []model.Field{{Name: "v", Visibility: "public", Readonly: true}}}
	phpxBindProps(c, c.Fields, nil)
	scopeOf := func(class string) *Scope {
		s := NewScope()
		s.Set("__class__", class)
		return s
	}
	empty := model.NewObject(c)
	if err := phpxReadonlyWrite(empty, "v", scopeOf("R")); err != nil {
		t.Errorf("first write from the class: %v", err)
	}
	if err := phpxReadonlyWrite(empty, "v", scopeOf("")); err == nil {
		t.Error("first write from global scope was allowed")
	}
	held := model.NewObject(c)
	held.SetProp("v", 1)
	if err := phpxReadonlyWrite(held, "v", scopeOf("R")); err == nil {
		t.Error("second write from the class was allowed")
	}
}

// TestPhpxSlotKey pins the two slots: a private property of an ancestor and
// the child's declaration of the same name live apart, and each scope reads
// its own.
func TestPhpxSlotKey(t *testing.T) {
	p := &model.Class{Name: "P", Fields: []model.Field{{Name: "a", Visibility: "private"}}}
	phpxBindProps(p, p.Fields, nil)
	c := &model.Class{Name: "C", Ancestors: []string{"P"}, Fields: phpxMergeFields(p.Fields, []model.Field{{Name: "a", Visibility: "private"}})}
	phpxBindProps(c, []model.Field{{Name: "a", Visibility: "private"}}, p)
	scopeOf := func(class string) *Scope {
		s := NewScope()
		s.Set("__class__", class)
		return s
	}
	obj := model.NewObject(c)
	if err := (&Runtime{}).phpxSeedShadow(obj, c, NewScope()); err != nil {
		t.Fatalf("seed: %v", err)
	}
	obj.SetProp("a", "child")
	obj.SetProp(phpxShadowKey("P", "a"), "parent")
	if key := (&Runtime{}).phpxSlotKey(obj, "a", scopeOf("P")); key != phpxShadowKey("P", "a") {
		t.Errorf("P reads %q", key)
	}
	if key := (&Runtime{}).phpxSlotKey(obj, "a", scopeOf("C")); key != "a" {
		t.Errorf("C reads %q", key)
	}
	var seen string
	(&Runtime{}).phpxReadable(obj, scopeOf("P"), func(name string, v any) bool {
		if name == "a" {
			seen = v.(string)
		}
		return true
	})
	if seen != "parent" {
		t.Errorf("P lists %q", seen)
	}
	seen = ""
	(&Runtime{}).phpxReadable(obj, scopeOf("C"), func(name string, v any) bool {
		if name == "a" {
			seen = v.(string)
		}
		return true
	})
	if seen != "child" {
		t.Errorf("C lists %q", seen)
	}
}
