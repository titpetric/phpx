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
// for a single one. kind is the import the statement opened with, `function`
// or `const`, and an item may carry its own. It returns the imports for the
// AST node the formatter prints.
func (p *parser) phpxGroupUse(kind string) ([]model.UseImport, error) {
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
		itemKind := kind
		if p.isKw("function") || p.isKw("const") {
			itemKind = p.cur().val
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
		switch itemKind {
		case "":
			if p.imports == nil {
				p.imports = map[string]string{}
			}
			p.imports[alias] = name
		case "function":
			if p.funcImports == nil {
				p.funcImports = map[string]string{}
			}
			p.funcImports[alias] = name
		case "const":
			if alias != name {
				if p.constAlias == nil {
					p.constAlias = map[string]string{}
				}
				p.constAlias[alias] = name
			}
		}
		imports = append(imports, model.UseImport{Path: name, Alias: aliasSpelling(name, alias)})
		if p.isOp(",") {
			p.next()
		}
	}
	return imports, p.eatOp("}")
}
