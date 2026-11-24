//go:build !phpx

package parser

import (
	"fmt"

	"github.com/titpetric/phpx/model"
)

// phpxStaticPropName answers the untagged build, which refuses the form.
func (p *parser) phpxStaticPropName() (model.Expr, error) {
	return nil, fmt.Errorf("line %d: a static property named by an expression is not supported", p.cur().line)
}
