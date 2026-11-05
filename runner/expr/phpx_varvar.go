package expr

import "github.com/titpetric/phpx/model"

// phpxVarVar compiles the two calls the parser makes of a variable variable
// into the env helpers that read and write the scope by name. The scope is
// written mid-expression, so variables read live.
func (dc *directCompiler) phpxVarVar(n *model.Call) (closure, bool, error) {
	if n.Name != "__phpx_varvar" && n.Name != "__phpx_varvar_set" {
		return nil, false, nil
	}
	args, err := dc.compileArgs(n.Args)
	if err != nil {
		return nil, true, err
	}
	dc.outParams = true
	return baseCall(n.Name, args), true, nil
}
