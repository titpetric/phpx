//go:build phpx

package parser

import (
	"fmt"

	"github.com/titpetric/phpx/model"
)

// phpxBracedNamespace parses `namespace Name { ... }`, and `namespace { ... }`
// for the global one. The statements inside are collected into the program as
// they are read, under the namespace and the imports of the block, which end
// with it.
func (p *parser) phpxBracedNamespace(name string, line int) (model.Stmt, error) {
	p.next() // {
	p.phpxBraced = true
	savedNS, savedLine, savedImports := p.namespace, p.namespaceLine, p.imports
	p.namespace, p.imports = name, nil
	if name != "" {
		p.namespaceLine = line
	}
	for !p.atEOF() && !p.isOp("}") {
		s, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if s != nil {
			p.pushStmt(s)
		}
	}
	if err := p.eatOp("}"); err != nil {
		return nil, fmt.Errorf("line %d: namespace block is not closed: %w", line, err)
	}
	if name == "" {
		p.namespace, p.namespaceLine = savedNS, savedLine
	}
	p.imports = savedImports
	return nil, nil
}
