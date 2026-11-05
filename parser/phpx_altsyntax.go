//go:build phpx

package parser

import "github.com/titpetric/phpx/model"

// phpxAltEnds are the keywords that close the body of an alternative-syntax
// block (`if (c): ... endif;`). elseif and else end the body without closing
// the construct.
var phpxAltEnds = []string{"endif", "endwhile", "endfor", "endforeach", "endswitch", "elseif", "else"}

// phpxAltBody reads the body of `if (c):`, `while (c):`, `for (;;):` and
// `foreach ($a as $b):` after the condition. It stops at the keyword that ends
// the body and consumes an `endX;` keyword, which closes the construct; an
// elseif or else is left to the if that is parsing. ok is false when the
// next token is no colon.
func (p *parser) phpxAltBody() (body []model.Stmt, ok bool, err error) {
	if !p.isOp(":") {
		return nil, false, nil
	}
	p.next()
	mark := p.stmts.mark()
	for !p.atEOF() && !p.isKw(phpxAltEnds...) {
		s, err := p.parseStmt()
		if err != nil {
			p.stmts.drop(mark)
			return nil, true, err
		}
		if s != nil {
			p.pushStmt(s)
		}
	}
	body = p.stmts.take(mark)
	if p.isKw("endif", "endwhile", "endfor", "endforeach", "endswitch") {
		p.next()
		p.optSemi()
	}
	return body, true, nil
}

// phpxSwitchOpen reads the `{` or the `:` that opens a switch.
func (p *parser) phpxSwitchOpen() (alt bool, err error) {
	if p.isOp(":") {
		p.next()
		return true, nil
	}
	return false, p.eatOp("{")
}

// phpxSwitchAtEnd reports whether the switch body is at its closing token.
func (p *parser) phpxSwitchAtEnd() bool {
	return p.isOp("}") || p.isKw("endswitch")
}

// phpxSwitchClose reads the `}` or the `endswitch;` that closes a switch.
func (p *parser) phpxSwitchClose() error {
	if p.isKw("endswitch") {
		p.next()
		p.optSemi()
		return nil
	}
	return p.eatOp("}")
}
