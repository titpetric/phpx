package formatter

import (
	"sort"
	"strings"

	"github.com/titpetric/phpx/model"
)

// printTrait writes `trait Name { ... }`. A trait body holds the members a
// class body holds and no heritage, so the same member list prints both.
func (p *printer) printTrait(n *model.TraitDecl) {
	p.line("trait " + shortName(n.Name) + " {")
	p.depth++
	p.printMembers(p.memberList(n), p.stmtEnd(n))
	p.depth--
	p.line("}")
}

// memberList collects the declarations of a class or trait body. The `use`
// line comes first, then constants, properties and methods, and each group
// keeps source order within itself.
func (p *printer) memberList(n model.Node) []classMember {
	var traits []string
	var rules []model.TraitRule
	var consts, fields, statics []model.Field
	var methods []*model.FuncDecl
	switch d := n.(type) {
	case *model.ClassDecl:
		traits, consts, fields, statics, methods = d.Traits, d.Consts, d.Fields, d.Statics, d.Methods
		rules = d.TraitRules
	case *model.TraitDecl:
		traits, consts, fields, statics, methods = d.Traits, d.Consts, d.Fields, d.Statics, d.Methods
		rules = d.TraitRules
	}
	members := make([]classMember, 0, len(consts)+len(fields)+len(statics)+len(methods)+len(traits))
	if len(traits) > 0 {
		names := make([]string, 0, len(traits))
		for _, name := range traits {
			names = append(names, p.typeName(name))
		}
		members = append(members, classMember{group: groupUse, print: func() {
			if len(rules) == 0 {
				p.line("use " + joinNames(names) + ";")
				return
			}
			p.line("use " + joinNames(names) + " {")
			p.depth++
			for _, r := range rules {
				p.line(p.traitRule(r) + ";")
			}
			p.depth--
			p.line("}")
		}})
	}
	for _, c := range consts {
		members = append(members, classMember{group: groupConst, span: c.Span, print: func() {
			vis := ""
			if c.Visibility != "" {
				vis = c.Visibility + " "
			}
			p.line(vis + "const " + c.Name + " = " + p.expr(c.Default) + ";")
		}})
	}
	for _, f := range fields {
		members = append(members, classMember{group: groupProp, span: f.Span, print: func() {
			p.line(p.field(f) + ";")
		}})
	}
	for _, f := range statics {
		members = append(members, classMember{group: groupProp, span: f.Span, print: func() {
			p.line(p.staticField(f) + ";")
		}})
	}
	for _, m := range methods {
		members = append(members, classMember{group: groupMethod, span: p.spans[m], print: func() {
			p.printFunc(m, true)
		}})
	}
	// Properties keep their declaration order; the parser collects
	// static ones separately because they are different storage.
	sort.SliceStable(members, func(i, j int) bool {
		if members[i].group != members[j].group {
			return members[i].group < members[j].group
		}
		return members[i].span.Start < members[j].span.Start
	})
	return members
}

// joinNames spells a comma-separated name list the way php writes one.
func joinNames(names []string) string {
	return strings.Join(names, ", ")
}

// traitRule spells one adaptation line without its semicolon.
func (p *printer) traitRule(r model.TraitRule) string {
	name := r.Method
	if r.Trait != "" {
		name = p.typeName(r.Trait) + "::" + r.Method
	}
	if r.IsInsteadof() {
		others := make([]string, 0, len(r.Insteadof))
		for _, o := range r.Insteadof {
			others = append(others, p.typeName(o))
		}
		return name + " insteadof " + joinNames(others)
	}
	out := name + " as"
	if r.Visibility != "" {
		out += " " + r.Visibility
	}
	if r.Alias != "" {
		out += " " + r.Alias
	}
	return out
}
