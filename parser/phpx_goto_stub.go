//go:build !phpx

package parser

import "github.com/titpetric/phpx/model"

// phpxGotoStmt answers the untagged build, which reads no goto and no label.
func (p *parser) phpxGotoStmt(_ token) (model.Stmt, bool) { return nil, false }
