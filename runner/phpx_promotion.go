//go:build phpx

package runner

import (
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxPromote returns the declaration with each promoted constructor parameter
// (`public function __construct(private int $x)`) turned into the property it
// declares and the `$this->x = $x;` that fills it. The parser keeps the
// modifiers on the parameter for the formatter, so the original declaration is
// left alone and a shallow copy carries the change.
func phpxPromote(cd *model.ClassDecl) *model.ClassDecl {
	var fields []model.Field
	methods := cd.Methods
	for i, m := range cd.Methods {
		if !strings.EqualFold(m.Name, "__construct") {
			continue
		}
		var body []model.Stmt
		for _, p := range m.Params {
			if p.Modifiers == "" {
				continue
			}
			fields = append(fields, model.Field{Name: p.Name, Visibility: phpxPromotedVisibility(p.Modifiers), Type: p.Type})
			body = append(body, &model.Assign{
				Target: &model.PropAccess{Base: &model.Var{Name: "this"}, Name: p.Name},
				Op:     "=",
				Value:  &model.Var{Name: p.Name},
			})
		}
		if len(body) == 0 {
			break
		}
		clone := *m
		clone.Body = append(body, m.Body...)
		methods = append([]*model.FuncDecl(nil), cd.Methods...)
		methods[i] = &clone
		break
	}
	if len(fields) == 0 {
		return cd
	}
	out := *cd
	out.Methods = methods
	out.Fields = phpxMergeFields(cd.Fields, fields)
	return &out
}

// phpxPromotedVisibility reads the visibility out of a parameter's modifiers;
// `readonly` alone is public, as php has it.
func phpxPromotedVisibility(modifiers string) string {
	for _, word := range strings.Fields(modifiers) {
		switch word {
		case "private", "protected", "public":
			return word
		}
	}
	return "public"
}
