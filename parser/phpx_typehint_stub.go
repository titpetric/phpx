//go:build !phpx

package parser

import (
	"github.com/titpetric/phpx/model"
)

// phpxResolveTypeHint answers the untagged build, which keeps a type as written.
func (p *parser) phpxResolveTypeHint(hint string) string { return hint }

// phpxCatchName answers the untagged build, which keeps a catch type as written.
func (p *parser) phpxCatchName(hint string) string { return hint }

// phpxQualifyConsts answers the untagged build, which declares a constant once.
func (p *parser) phpxQualifyConsts(consts []model.Field) []model.Field { return consts }

// phpxIntersection answers the untagged build, which reads no intersection type.
func (p *parser) phpxIntersection() bool { return false }

// phpxTypedConst answers the untagged build, which reads no typed constant.
func (p *parser) phpxTypedConst() bool { return false }
