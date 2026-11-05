//go:build !phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxBindRefArgsStubKeepsArgs(t *testing.T) {
	in := []any{1}
	if out := phpxBindRefArgs(&model.FuncDecl{}, in, nil); len(out) != 1 {
		t.Fatalf("got %v", out)
	}
}
