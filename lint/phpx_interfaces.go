//go:build phpx

package lint

import "github.com/titpetric/phpx/model"

// phpxInterfaceViolations reports a class that does not meet its `implements`
// contract using the methods it has once its traits and ancestors contribute,
// which is the runtime's answer and php's. It is the linter's half of the
// runner's check; the untagged build keeps upstream's AST-only rule.
//
// An anonymous class is left out: the linter reports a declaration a reader
// can point at and an anonymous class has no statement span. The runtime
// checks it. A class whose ancestor this file does not declare is left out
// too: the method may be the parent's, in another file, and the runtime
// autoloads the parent and checks the finished class.
func phpxInterfaceViolations(prog *model.Program) []model.InterfaceViolation {
	all := model.PhpxCheckInterfaces(prog, func(cd *model.ClassDecl) map[string]bool {
		return model.PhpxEffectiveMethodNames(cd, prog.Stmts)
	})
	out := make([]model.InterfaceViolation, 0, len(all))
	for _, v := range all {
		if _, ok := prog.SourceSpans[v.Decl]; !ok {
			continue
		}
		if !model.PhpxContractResolved(v.Decl, prog.Stmts) {
			continue
		}
		out = append(out, v)
	}
	return out
}
