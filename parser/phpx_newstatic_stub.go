//go:build !phpx

package parser

import "github.com/titpetric/phpx/model"

// phpxNewStaticProp keeps the untagged parse, which names no static property as a class.
func (p *parser) phpxNewStaticProp(_ string) (model.Expr, bool, error) { return nil, false, nil }
