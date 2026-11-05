package parser

import (
	"fmt"

	"github.com/titpetric/phpx/model"
)

// Hook methods carry this prefix and the property name. The runtime looks the
// pair up when a property is read or written.
const (
	phpxGetHook = "__phpx_get_"
	phpxSetHook = "__phpx_set_"
)

// phpxIsSetVisibility reports whether the modifier at the cursor is followed
// by `(set)`, the asymmetric form that narrows writes and not reads.
func (p *parser) phpxIsSetVisibility() bool {
	return p.i+3 < len(p.toks) && p.toks[p.i+1].val == "(" && p.toks[p.i+2].val == "set" && p.toks[p.i+3].val == ")"
}

// phpxSkipSetVisibility consumes the `(set)` of an asymmetric modifier
// (`private(set)`) after the keyword and reports whether it was there. The
// modifier narrows writes only; the fork reads it and does not enforce it.
func (p *parser) phpxSkipSetVisibility() bool {
	if !p.isOp("(") || p.i+2 >= len(p.toks) || p.toks[p.i+1].val != "set" || p.toks[p.i+2].val != ")" {
		return false
	}
	p.i += 3
	return true
}

// phpxParseHooks consumes the `{ get {...} set($v) => expr; }` block after a
// property declaration (php 8.4) and lowers each hook to a method on cd named
// for the property: `get` returns, `set` receives `$value` unless it names its
// own parameter. `set => expr` stores expr in the property.
func (p *parser) phpxParseHooks(cd *model.ClassDecl, field model.Field) error {
	p.next() // {
	for !p.isOp("}") && !p.atEOF() {
		if p.isOp("&") {
			p.next()
		}
		kind := p.cur().val
		if kind != "get" && kind != "set" {
			return fmt.Errorf("line %d: expected get or set hook, got %s", p.cur().line, p.cur())
		}
		p.next()
		fd := &model.FuncDecl{Name: phpxGetHook + field.Name, Visibility: "public", Filename: p.file}
		if kind == "set" {
			fd.Name = phpxSetHook + field.Name
			fd.Params = []model.Param{{Name: "value"}}
		}
		if p.isOp("(") {
			params, err := p.parseParams()
			if err != nil {
				return err
			}
			fd.Params = params
		}
		was := p.function
		p.function = fd.Name
		err := p.phpxHookBody(fd, kind, field.Name)
		p.function = was
		if err != nil {
			return err
		}
		cd.Methods = append(cd.Methods, fd)
	}
	return p.eatOp("}")
}

// phpxHookBody reads `=> expr;`, `{ ... }` or the `;` of a hook with no body.
func (p *parser) phpxHookBody(fd *model.FuncDecl, kind, prop string) error {
	switch {
	case p.isOp("=>"):
		p.next()
		x, err := p.parseExpr()
		if err != nil {
			return err
		}
		p.optSemi()
		if kind == "get" {
			fd.Body = []model.Stmt{&model.Return{Value: x}}
			return nil
		}
		target := &model.PropAccess{Base: &model.Var{Name: "this"}, Name: prop}
		fd.Body = []model.Stmt{&model.Assign{Target: target, Op: "=", Value: x}}
		return nil
	case p.isOp(";"):
		p.next()
		return nil
	}
	body, err := p.parseBlock()
	fd.Body = body
	return err
}
