package parser

import (
	"fmt"

	"github.com/titpetric/phpx/model"
)

// parseTrait consumes `trait Name { ... }`. A trait body holds what a class
// body holds, so it is read by the same parser; the members land on the
// temporary ClassDecl first because parseClassBody writes there, and they are
// copied to the TraitDecl after. The parser records a trait so the formatter
// prints it back; the untagged build runs it as the no-op it always was and
// the fork flattens the members into each class that says `use`.
func (p *parser) parseTrait() (model.Stmt, error) {
	p.next() // trait
	if p.cur().kind != tIdent {
		return nil, fmt.Errorf("line %d: expected trait name", p.cur().line)
	}
	td := &model.TraitDecl{Name: p.qualify(p.next().val, false)}
	body := &model.ClassDecl{Name: td.Name}
	wasClass := p.class
	wasTrait := p.trait
	p.class, p.trait = td.Name, td.Name
	err := p.parseClassBody(body)
	p.class, p.trait = wasClass, wasTrait
	if err != nil {
		return nil, err
	}
	td.Fields, td.Statics, td.Consts, td.Methods, td.Traits = body.Fields, body.Statics, body.Consts, body.Methods, body.Traits
	td.TraitRules = body.TraitRules
	return td, nil
}

// parseTraitUse consumes `use A, B;` in a class or trait body and records the
// names on dst and the adaptation block (`A::f insteadof B; B::f as g;`) on
// rules.
func (p *parser) parseTraitUse(dst *[]string, rules *[]model.TraitRule) error {
	p.next() // use
	for {
		name, absolute, err := p.parseQualifiedName(true)
		if err != nil {
			return fmt.Errorf("line %d: expected trait name after use: %w", p.cur().line, err)
		}
		*dst = append(*dst, p.qualify(name, absolute))
		if !p.isOp(",") {
			break
		}
		p.next()
	}
	if p.isOp("{") {
		return p.parseTraitRules(rules)
	}
	return p.eatOp(";")
}

// parseTraitRules reads the `{ ... }` block of a `use` line.
func (p *parser) parseTraitRules(rules *[]model.TraitRule) error {
	p.next() // {
	for !p.isOp("}") {
		rule, err := p.parseTraitRule()
		if err != nil {
			return err
		}
		*rules = append(*rules, rule)
	}
	p.next() // }
	return nil
}

// parseTraitRule reads `[Trait::]method insteadof A, B;` or
// `[Trait::]method as [visibility] [alias];`.
func (p *parser) parseTraitRule() (model.TraitRule, error) {
	var rule model.TraitRule
	name, absolute, err := p.parseQualifiedName(true)
	if err != nil {
		return rule, fmt.Errorf("line %d: expected trait method in adaptation: %w", p.cur().line, err)
	}
	rule.Method = name
	if p.isOp("::") {
		p.next()
		rule.Trait = p.qualify(name, absolute)
		if p.cur().kind != tIdent {
			return rule, fmt.Errorf("line %d: expected method name after %s::", p.cur().line, name)
		}
		rule.Method = p.next().val
	}
	switch {
	case p.isKw("insteadof"):
		p.next()
		if rule.Trait == "" {
			return rule, fmt.Errorf("line %d: insteadof needs Trait::method", p.cur().line)
		}
		for {
			other, abs, err := p.parseQualifiedName(true)
			if err != nil {
				return rule, fmt.Errorf("line %d: expected trait name after insteadof: %w", p.cur().line, err)
			}
			rule.Insteadof = append(rule.Insteadof, p.qualify(other, abs))
			if !p.isOp(",") {
				break
			}
			p.next()
		}
	case p.isKw("as"):
		p.next()
		if p.isKw("public", "protected", "private") {
			rule.Visibility = p.next().val
		}
		if p.cur().kind == tIdent {
			rule.Alias = p.next().val
		}
		if rule.Alias == "" && rule.Visibility == "" {
			return rule, fmt.Errorf("line %d: expected alias or visibility after as", p.cur().line)
		}
	default:
		return rule, fmt.Errorf("line %d: expected insteadof or as, got %s", p.cur().line, p.cur())
	}
	return rule, p.eatOp(";")
}

// phpxMagicClass compiles __CLASS__. Inside a trait body the frame's own
// __class__ is read at run time, which hoist set to the class that used the
// trait; anywhere else the name is a literal, as upstream compiles it.
func (p *parser) phpxMagicClass() model.Expr {
	if p.trait != "" {
		return p.newVar("__class__")
	}
	return p.newLit(p.class)
}
