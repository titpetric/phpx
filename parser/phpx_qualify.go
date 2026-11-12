package parser

import (
	"strings"
)

// qualifyCall resolves the unqualified name of a function call. A `use` alias
// answers in an exact spelling only: the class table reads an alias in any
// case, and a call `config()` beside `use Pkg\Config;` would name the class.
func (p *parser) qualifyCall(name string) string {
	if full, ok := p.imports[name]; ok {
		return full
	}
	if p.namespace == "" || strings.ContainsRune(name, '\\') {
		return name
	}
	return p.namespace + "\\" + name
}
