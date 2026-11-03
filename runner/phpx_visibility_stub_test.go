//go:build !phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxVisibilityConfersNothingUntagged pins upstream's answer: a private
// method runs from anywhere. The fork's build is phpx_visibility_test.go.
func TestPhpxVisibilityConfersNothingUntagged(t *testing.T) {
	t.Parallel()
	rt := &Runtime{}
	decl := &model.FuncDecl{Name: "f", Visibility: "private"}
	if err := rt.phpxCheckVisible(nil, &model.Class{Name: "A"}, decl); err != nil {
		t.Errorf("phpxCheckVisible = %v, want nil in an untagged build", err)
	}
}
