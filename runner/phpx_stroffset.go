//go:build phpx

package runner

import (
	"fmt"

	"github.com/titpetric/phpx/model"
)

// phpxStringOffset writes one byte of a string, `$s[0] = 'X'`: the string is
// rebuilt and stored back into the place it was read from. The first byte of
// the value is the one written, and an offset past the end pads with spaces.
// It reports false when base is not a string.
func (rt *Runtime) phpxStringOffset(target *model.Index, base any, op string, rhs any, scope *Scope) (bool, error) {
	s, ok := base.(string)
	if !ok || target.Index == nil {
		return false, nil
	}
	keyValue, err := rt.Eval(target.Index, scope)
	if err != nil {
		return true, err
	}
	at, isInt := keyValue.(int64)
	if !isInt {
		if n, ok := keyValue.(int); ok {
			at, isInt = int64(n), true
		}
	}
	if !isInt {
		return true, fmt.Errorf("Cannot access offset of type %T on string", keyValue)
	}
	if at < 0 {
		at += int64(len(s))
	}
	if at < 0 {
		return true, fmt.Errorf("Illegal string offset %d", at)
	}
	if op != "=" && op != "" {
		cur := ""
		if at < int64(len(s)) {
			cur = s[at : at+1]
		}
		next, err := applyAssignOp(op, cur, rhs)
		if err != nil {
			return true, err
		}
		rhs = next
	}
	text, err := phpxStringOf(rhs)
	if err != nil {
		return true, err
	}
	if text == "" {
		return true, fmt.Errorf("Cannot assign an empty string to a string offset")
	}
	b := []byte(s)
	for int64(len(b)) <= at {
		b = append(b, ' ')
	}
	b[at] = text[0]
	return true, rt.assignTo(target.Base, string(b), scope)
}
