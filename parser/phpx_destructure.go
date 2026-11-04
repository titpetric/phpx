package parser

import (
	"fmt"

	"github.com/titpetric/phpx/model"
)

// phpxDestructure lowers `[$a, [$b, $c], 'k' => $d] = src`. A flat positional
// pattern is a ListExpr, which the runtime already assigns. Any other shape
// stores src in a temporary and reads each element off it, in an array literal
// so the steps run in order inside one expression: `[($t = src), ($a = $t[0]),
// ...][0]`, which also leaves src as the expression's value as php does.
func phpxDestructure(pattern, src model.Expr, line int) (model.Expr, error) {
	if list, ok := phpxFlatPattern(pattern); ok {
		return &model.AssignExpr{Target: list, Op: "=", Value: src, Line: line}, nil
	}
	tmp := phpxTemp("d")
	steps := []model.ArrayItem{{Val: &model.Parenthesized{X: &model.AssignExpr{Target: tmp, Op: "=", Value: src, Line: line}}}}
	steps, err := phpxPatternSteps(steps, pattern, tmp, line)
	if err != nil {
		return nil, err
	}
	return &model.Index{Base: &model.ArrayLit{Items: steps}, Index: &model.Lit{Value: int64(0), Raw: "0"}}, nil
}

// phpxFlatPattern reports a pattern of positional plain targets as a ListExpr.
func phpxFlatPattern(pattern model.Expr) (*model.ListExpr, bool) {
	if l, ok := pattern.(*model.ListExpr); ok {
		for _, e := range l.Elems {
			if _, nested := e.(*model.ListExpr); nested {
				return nil, false
			}
		}
		return l, true
	}
	arr, ok := pattern.(*model.ArrayLit)
	if !ok {
		return nil, false
	}
	elems := make([]model.Expr, len(arr.Items))
	for i, it := range arr.Items {
		if it.Key != nil || it.Val == nil || !phpxPlainTarget(it.Val) {
			return nil, false
		}
		elems[i] = it.Val
	}
	return &model.ListExpr{Elems: elems}, true
}

func phpxPlainTarget(e model.Expr) bool {
	switch e.(type) {
	case *model.Var, *model.Index, *model.PropAccess, *model.StaticProp:
		return true
	}
	return false
}

// phpxPatternSteps appends one assignment per target of pattern, read off src.
func phpxPatternSteps(steps []model.ArrayItem, pattern model.Expr, src *model.Var, line int) ([]model.ArrayItem, error) {
	var items []model.ArrayItem
	switch pt := pattern.(type) {
	case *model.ArrayLit:
		items = pt.Items
	case *model.ListExpr:
		for _, e := range pt.Elems {
			items = append(items, model.ArrayItem{Val: e})
		}
	}
	for i, it := range items {
		if it.Val == nil {
			continue
		}
		key := it.Key
		if key == nil {
			key = &model.Lit{Value: int64(i), Raw: fmt.Sprint(i)}
		}
		read := &model.Index{Base: src, Index: key}
		switch it.Val.(type) {
		case *model.ArrayLit, *model.ListExpr:
			tmp := phpxTemp("d")
			steps = append(steps, model.ArrayItem{Val: &model.Parenthesized{X: &model.AssignExpr{Target: tmp, Op: "=", Value: read, Line: line}}})
			var err error
			if steps, err = phpxPatternSteps(steps, it.Val, tmp, line); err != nil {
				return nil, err
			}
		default:
			if !phpxPlainTarget(it.Val) {
				return nil, fmt.Errorf("line %d: cannot assign to this element of a destructuring pattern", line)
			}
			steps = append(steps, model.ArrayItem{Val: &model.Parenthesized{X: &model.AssignExpr{Target: it.Val, Op: "=", Value: read, Line: line}}})
		}
	}
	return steps, nil
}

func phpxTemp(kind string) *model.Var {
	return &model.Var{Name: fmt.Sprintf("__phpx_%s%d", kind, phpxTemps.Add(1))}
}

// phpxForeachPattern rewrites a foreach whose key or value target is an array
// pattern or a list(). The target binds a temporary, which the first statement
// of the body destructures; the runtime assigns a list element by element and
// has no foreach path for a pattern of its own.
func (p *parser) phpxForeachPattern(node *model.Foreach) error {
	var prologue []model.Stmt
	for _, slot := range []*model.Expr{&node.KeyTarget, &node.ValTarget} {
		switch (*slot).(type) {
		case *model.ArrayLit, *model.ListExpr:
		default:
			continue
		}
		tmp := phpxTemp("f")
		lowered, err := phpxDestructure(*slot, tmp, 0)
		if err != nil {
			return err
		}
		*slot = tmp
		if as, ok := lowered.(*model.AssignExpr); ok {
			prologue = append(prologue, p.newAssign(as.Target, as.Op, as.Value))
			continue
		}
		prologue = append(prologue, p.newExprStmt(lowered))
	}
	if len(prologue) > 0 {
		node.Body = append(prologue, node.Body...)
	}
	if v, ok := node.ValTarget.(*model.Var); ok {
		node.ValVar = v.Name
	}
	if v, ok := node.KeyTarget.(*model.Var); ok {
		node.KeyVar = v.Name
	}
	return nil
}
