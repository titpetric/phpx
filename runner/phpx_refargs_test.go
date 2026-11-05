//go:build phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxBindRefArgsPlainWithoutRefs(t *testing.T) {
	in := []any{1, 2}
	out := phpxBindRefArgs(&model.FuncDecl{}, in, &Scope{})
	if len(out) != 2 || out[0] != 1 {
		t.Fatalf("got %v", out)
	}
}
