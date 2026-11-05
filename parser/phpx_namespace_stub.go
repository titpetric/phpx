//go:build !phpx

package parser

import (
	"fmt"

	"github.com/titpetric/phpx/model"
)

// phpxBracedNamespace answers the untagged build, which refuses a braced
// namespace as it always has.
func (p *parser) phpxBracedNamespace(_ string, line int) (model.Stmt, error) {
	return nil, fmt.Errorf("line %d: braced namespaces are not supported", line)
}
