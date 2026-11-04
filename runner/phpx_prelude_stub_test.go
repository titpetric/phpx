//go:build !phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

func TestPhpxRecordFuncStubKeepsNothing(t *testing.T) {
	t.Parallel()
	phpxRecordFunc(&model.FuncDecl{Name: "f"})
}
