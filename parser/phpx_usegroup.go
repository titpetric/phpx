package parser

import (
	"fmt"
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxAtGroupUse reports whether the tokens ahead spell the prefix of a group
// use, `Ns\Sub\{`, so parseUse can read it as one.
func (p *parser) phpxAtGroupUse() bool {
	i := 0
	if p.isOpAt(i, "\\") {
		i++
	}
	for p.peek(i).kind == tIdent && p.isOpAt(i+1, "\\") {
		if p.isOpAt(i+2, "{") {
			return true
		}
		i += 2
	}
	return false
}

// phpxGroupUse reads `Ns\{A, B as C, Sub\D}` and binds each name, as `use` does
// for a single one. It returns the imports for the AST node the formatter prints.
func (p *parser) phpxGroupUse() ([]model.UseImport, error) {
	var prefix []string
	if p.isOp("\\") {
		p.next()
	}
	for p.cur().kind == tIdent && p.isOpAt(1, "\\") {
		prefix = append(prefix, p.next().val)
		p.next() // \
		if p.isOp("{") {
			break
		}
	}
	if err := p.eatOp("{"); err != nil {
		return nil, err
	}
	var imports []model.UseImport
	for !p.isOp("}") {
		if p.isKw("function") || p.isKw("const") {
			p.next()
		}
		var parts []string
		for {
			if p.cur().kind != tIdent {
				return nil, fmt.Errorf("line %d: expected a name in the group use, got %s", p.cur().line, p.cur())
			}
			parts = append(parts, p.next().val)
			if !p.isOp("\\") {
				break
			}
			p.next()
		}
		name := strings.Join(append(append([]string(nil), prefix...), parts...), "\\")
		alias := parts[len(parts)-1]
		if p.isKw("as") {
			p.next()
			if p.cur().kind != tIdent {
				return nil, fmt.Errorf("line %d: expected alias name after as", p.cur().line)
			}
			alias = p.next().val
		}
		if p.imports == nil {
			p.imports = map[string]string{}
		}
		p.imports[alias] = name
		imports = append(imports, model.UseImport{Path: name, Alias: aliasSpelling(name, alias)})
		if p.isOp(",") {
			p.next()
		}
	}
	return imports, p.eatOp("}")
}
