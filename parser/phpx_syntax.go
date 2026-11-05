//go:build phpx

package parser

import (
	"fmt"

	"github.com/titpetric/phpx/model"
)

// phpxByRefName consumes the `&` of `function &name()`, a method that returns by reference.
func (p *parser) phpxByRefName() bool {
	if p.isOp("&") {
		p.next()
		return true
	}
	return false
}

// phpxStaticDynamicMember reads `Class::{expr}(args)`: the braces hold the
// name of the static method to call.
func (p *parser) phpxStaticDynamicMember(class string) (model.Expr, bool, error) {
	if !p.isOp("{") {
		return nil, false, nil
	}
	p.next()
	name, err := p.parseExpr()
	if err != nil {
		return nil, true, err
	}
	if err := p.eatOp("}"); err != nil {
		return nil, true, err
	}
	if !p.isOp("(") {
		return nil, true, fmt.Errorf("line %d: expected ( after a dynamic static member", p.cur().line)
	}
	args, firstClass, err := p.parseCallArgs()
	if err != nil {
		return nil, true, err
	}
	return &model.StaticCall{Class: class, MethodExpr: name, Args: args, FirstClass: firstClass}, true, nil
}

// phpxLeadingDot lexes a float that begins with the point, `.1` and `.5e3`.
func (l *lexer) phpxLeadingDot() bool {
	start := l.pos
	i := l.pos + 1
	for i < len(l.src) && (l.src[i] >= '0' && l.src[i] <= '9' || l.src[i] == '_') {
		i++
	}
	if i < len(l.src) && (l.src[i] == 'e' || l.src[i] == 'E') {
		j := i + 1
		if j < len(l.src) && (l.src[j] == '+' || l.src[j] == '-') {
			j++
		}
		if j < len(l.src) && l.src[j] >= '0' && l.src[j] <= '9' {
			for j < len(l.src) && l.src[j] >= '0' && l.src[j] <= '9' {
				j++
			}
			i = j
		}
	}
	l.advance(i - start)
	l.emit(tFloat, l.src[start:l.pos])
	return true
}
