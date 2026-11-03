//go:build phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxCheckVisible holds the rule down to its parts: private reaches the
// declaring class alone, protected its family, and the end-to-end behaviour is
// settled against php by phpx/testdata/classes/method_visibility.phpt.
func TestPhpxCheckVisible(t *testing.T) {
	t.Parallel()
	rt := &Runtime{classes: map[string]*model.Class{
		"A": {Name: "A"},
		"B": {Name: "B", Ancestors: []string{"A"}},
		"X": {Name: "X"},
	}}
	scopeOf := func(class string) *Scope {
		s := rt.newScope()
		if class != "" {
			s.Set("__class__", class)
		}
		return s
	}
	cases := []struct {
		visibility, from string
		ok               bool
	}{
		{"public", "", true},
		{"", "", true},
		{"private", "A", true},
		{"private", "B", false},
		{"private", "", false},
		{"protected", "B", true},
		{"protected", "A", true},
		{"protected", "X", false},
		{"protected", "", false},
	}
	class := rt.classes["A"]
	for _, c := range cases {
		decl := &model.FuncDecl{Name: "f", Visibility: c.visibility}
		err := rt.phpxCheckVisible(scopeOf(c.from), class, decl)
		if (err == nil) != c.ok {
			t.Errorf("%q from %q: err = %v, want ok=%v", c.visibility, c.from, err, c.ok)
		}
	}
}
