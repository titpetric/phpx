//go:build phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxMethodMemoizesLookup covers the three answers the memo holds: an
// exact key, the EqualFold fallback that costs a scan, and a miss. A second
// ask must answer the same as the first for each.
func TestPhpxMethodMemoizesLookup(t *testing.T) {
	t.Parallel()
	class := &model.Class{Name: "A", Methods: map[string]*model.FuncDecl{
		"speak": {Name: "speak"},
		"__GET": {Name: "__GET"},
	}}
	other := &model.Class{Name: "B"}
	rt := New(nil, Options{})
	for _, tc := range []struct {
		name   string
		class  *model.Class
		method string
		ok     bool
	}{
		{"exact", class, "speak", true},
		{"folded", class, "__get", true},
		{"miss", class, "__set", false},
		{"other class", other, "speak", false},
	} {
		first, firstOK := rt.phpxMethod(tc.class, tc.method)
		second, secondOK := rt.phpxMethod(tc.class, tc.method)
		if firstOK != tc.ok || secondOK != tc.ok {
			t.Errorf("%s: ok = %v then %v, want %v", tc.name, firstOK, secondOK, tc.ok)
		}
		if first != second {
			t.Errorf("%s: the memo answered a different method", tc.name)
		}
	}
	if _, ok := rt.phpxMethod(nil, "speak"); ok {
		t.Error("a nil class has no method")
	}
}
