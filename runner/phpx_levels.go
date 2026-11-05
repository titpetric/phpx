//go:build phpx

package runner

import "github.com/titpetric/phpx/model"

// phpxOuterLevels records how many loops beyond the innermost a `break N` or
// `continue N` leaves. Zero and one both name the innermost alone.
func (rt *Runtime) phpxOuterLevels(levels int) {
	rt.phpxOuter = 0
	if levels > 1 {
		rt.phpxOuter = levels - 1
	}
}

// phpxLeave reports whether the loop that got fl from its body has to pass it
// on to the loop around it: a break or continue with levels left to go. It
// uses one level up.
func (rt *Runtime) phpxLeave(fl flow) bool {
	if fl == flowGoto {
		return true
	}
	if rt.phpxOuter > 0 && (fl == flowBreak || fl == flowContinue) {
		rt.phpxOuter--
		return true
	}
	return false
}

// phpxLabelIndex answers where in stmts the goto in flight lands: the index of
// the label it names, or -1 when this list holds none and the jump goes on to
// the list around it.
func (rt *Runtime) phpxLabelIndex(stmts []model.Stmt) int {
	for i, s := range stmts {
		if l, ok := s.(*model.Label); ok && l.Name == rt.phpxGotoLabel {
			return i
		}
	}
	return -1
}

// phpxGotoInto lands a goto in flight on a label inside an `if` of stmts, which
// php allows (a label may sit in a nested block, never in a loop or a switch).
// It runs the branch from the label to its end, then the statements after the
// `if`, and answers the result of the whole list; handled is false when no
// branch of this list holds the label.
func (rt *Runtime) phpxGotoInto(stmts []model.Stmt, scope *Scope) (handled bool, val any, fl flow, err error) {
	for j, s := range stmts {
		n, ok := s.(*model.If)
		if !ok {
			continue
		}
		for _, branch := range [][]model.Stmt{n.Then, n.Else} {
			at := rt.phpxLabelIndex(branch)
			if at >= 0 {
				val, fl, err = rt.exec(branch[at:], scope)
			} else if h, v, f, e := rt.phpxGotoInto(branch, scope); h {
				val, fl, err = v, f, e
			} else {
				continue
			}
			if err != nil || fl != flowNormal {
				return true, val, fl, err
			}
			val, fl, err = rt.exec(stmts[j+1:], scope)
			return true, val, fl, err
		}
	}
	return false, nil, flowNormal, nil
}
