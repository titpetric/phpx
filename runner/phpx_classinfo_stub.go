//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// phpxAliasMatches answers the untagged build, which has no interface aliases.
func phpxAliasMatches(_, _ string) bool { return false }

// phpxForgetClassInfo answers the untagged build, which caches no class description.
func phpxForgetClassInfo(_ *model.Class) {}
