//go:build phpx

package phpx

import (
	"sort"
	"strings"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
)

const (
	sortDesc    = 3
	sortAsc     = 4
	sortRegular = 0
	sortNumeric = 1
	sortString  = 2
	sortCase    = 8
)

// sortColumn is one array array_multisort orders by, with its order and flags.
type sortColumn struct {
	target *model.Array
	keys   []any
	vals   []any
	desc   bool
	flags  int64
}

func (c *sortColumn) compare(i, j int) int {
	a, b := c.vals[i], c.vals[j]
	var r int
	switch c.flags &^ sortCase {
	case sortNumeric:
		fa, fb := phpval.Float(a), phpval.Float(b)
		switch {
		case fa < fb:
			r = -1
		case fa > fb:
			r = 1
		}
	case sortString:
		sa, sb := phpval.String(a), phpval.String(b)
		if c.flags&sortCase != 0 {
			sa, sb = strings.ToLower(sa), strings.ToLower(sb)
		}
		r = strings.Compare(sa, sb)
	default:
		r = phpval.Compare(a, b)
	}
	if c.desc {
		return -r
	}
	return r
}

// multisort is array_multisort: the arrays are columns of one table, ordered by
// the first, ties broken by the next, and every array is rewritten in place.
func multisort(args []any) (bool, error) {
	var cols []*sortColumn
	for _, a := range args {
		switch t := a.(type) {
		case *model.Array:
			col := &sortColumn{target: t, flags: sortRegular}
			t.Range(func(k, v any) bool {
				col.keys = append(col.keys, k)
				col.vals = append(col.vals, v)
				return true
			})
			cols = append(cols, col)
		case int64:
			if len(cols) == 0 {
				return false, &valueError{message: "array_multisort(): Argument #1 ($array) must be an array or a sort flag"}
			}
			last := cols[len(cols)-1]
			switch t {
			case sortAsc:
				last.desc = false
			case sortDesc:
				last.desc = true
			default:
				last.flags = t
			}
		default:
			return false, &valueError{message: "array_multisort(): Argument must be an array or a sort flag"}
		}
	}
	if len(cols) == 0 {
		return false, nil
	}
	n := len(cols[0].vals)
	for _, c := range cols {
		if len(c.vals) != n {
			return false, &valueError{message: "Array sizes are inconsistent"}
		}
	}
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(x, y int) bool {
		for _, c := range cols {
			if r := c.compare(order[x], order[y]); r != 0 {
				return r < 0
			}
		}
		return false
	})
	for _, c := range cols {
		c.target.Clear()
		for _, idx := range order {
			if k, isStr := c.keys[idx].(string); isStr {
				c.target.Set(k, c.vals[idx])
			} else {
				c.target.Append(c.vals[idx])
			}
		}
	}
	return true, nil
}
