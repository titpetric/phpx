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
	return td, nil
}

// parseTraitUse consumes `use A, B;` in a class or trait body and records the
// names on dst. The adaptation block (`A::f insteadof B; B::f as g;`) is not
// supported: the fork's flattening stops at composing the members, and a
// program that needs the block is told so instead of losing it silently.
func (p *parser) parseTraitUse(dst *[]string) error {
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
		return fmt.Errorf("line %d: trait adaptations (`insteadof`, `as`) are not supported", p.cur().line)
	}
	return p.eatOp(";")
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
