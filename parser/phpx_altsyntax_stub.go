//go:build !phpx

package parser

import (
	"github.com/titpetric/phpx/model"
)

// phpxAltBody answers the untagged build, which reads no alternative syntax.
func (p *parser) phpxAltBody() ([]model.Stmt, bool, error) { return nil, false, nil }

// phpxSwitchOpen reads the `{` that opens a switch.
func (p *parser) phpxSwitchOpen() (bool, error) { return false, p.eatOp("{") }

// phpxSwitchAtEnd reports whether the switch body is at its closing brace.
func (p *parser) phpxSwitchAtEnd() bool { return p.isOp("}") }

// phpxSwitchClose reads the `}` that closes a switch.
func (p *parser) phpxSwitchClose() error { return p.eatOp("}") }
