//go:build !phpx

package parser

import "github.com/titpetric/phpx/model"

// phpxNewStaticProp answers the untagged build, where `new A::$b` is a parse error.
func (p *parser) phpxNewStaticProp(_ string) (model.Expr, bool) { return nil, false }
