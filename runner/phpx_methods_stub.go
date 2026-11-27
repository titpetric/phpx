//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxMethodKey names one memoized lookup: a class and the method spelling.
type phpxMethodKey struct {
	class  *model.Class
	method string
}

// phpxMethodHit is what one lookup answered, nil for a name no class declares.
type phpxMethodHit struct {
	decl *model.FuncDecl
	ok   bool
}
