package expr

import (
	"fmt"

	"github.com/titpetric/phpx/model"
)

// phpxSpreadArray compiles an array literal holding `...$other`. Each spread
// splices its elements in: an integer key appends, a string key keeps its name,
// as php 8.1 does. It reports false for a literal with no spread.
func (dc *directCompiler) phpxSpreadArray(n *model.ArrayLit) (closure, bool, error) {
	spread := false
	for _, it := range n.Items {
		if it.Val != nil && isSpread(it.Val) {
			spread = true
		}
	}
	if !spread {
		return nil, false, nil
	}
	type entry struct {
		key, val closure
		spread   bool
	}
	entries := make([]entry, len(n.Items))
	for i, it := range n.Items {
		var e entry
		var err error
		if it.Key != nil {
			if e.key, err = dc.compile(it.Key); err != nil {
				return nil, true, err
			}
		}
		val := it.Val
		if isSpread(val) {
			e.spread, val = true, val.(*model.Unary).X
		}
		if e.val, err = dc.compile(val); err != nil {
			return nil, true, err
		}
		entries[i] = e
	}
	return func(env *Env) (any, error) {
		out := model.NewArray()
		for _, e := range entries {
			var key any
			if e.key != nil {
				k, err := e.key(env)
				if err != nil {
					return nil, err
				}
				key = k
			}
			v, err := e.val(env)
			if err != nil {
				return nil, err
			}
			if !e.spread {
				if key == nil {
					out.Append(v)
				} else {
					out.Set(key, v)
				}
				continue
			}
			src, ok := v.(*model.Array)
			if !ok {
				var err error
				if src, ok, err = traversePairs(env, v); err != nil {
					return nil, err
				}
				if !ok {
					if src, ok = goPairs(v); !ok {
						return nil, fmt.Errorf("Only arrays and Traversables can be unpacked")
					}
				}
			}
			src.Range(func(k, val any) bool {
				if _, isStr := k.(string); isStr {
					out.Set(k, val)
				} else {
					out.Append(val)
				}
				return true
			})
		}
		return out, nil
	}, true, nil
}
