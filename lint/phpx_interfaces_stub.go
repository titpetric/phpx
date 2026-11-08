//go:build !phpx

package lint

import (
	"github.com/titpetric/phpx/modelutil"
	"github.com/titpetric/phpx/model"
)

// phpxInterfaceViolations answers the untagged build with upstream's rule: a
// class has the methods it declares in its own body, and neither `extends`
// nor a trait confers one.
func phpxInterfaceViolations(prog *model.Program) []model.InterfaceViolation {
	return modelutil.CheckInterfaces(prog.Stmts)
}
