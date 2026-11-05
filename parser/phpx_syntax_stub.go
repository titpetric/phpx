//go:build !phpx

package parser

import (
	"github.com/titpetric/phpx/model"
)

// phpxByRefName answers the untagged build, which has no by-reference method name.
func (p *parser) phpxByRefName() bool { return false }

// phpxStaticDynamicMember answers the untagged build, which reads no `Class::{expr}()`.
func (p *parser) phpxStaticDynamicMember(_ string) (model.Expr, bool, error) { return nil, false, nil }

// phpxLeadingDot answers the untagged lexer, where a point starts an operator.
func (l *lexer) phpxLeadingDot() bool { return false }
