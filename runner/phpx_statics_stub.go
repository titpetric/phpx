//go:build !phpx

package runner

import (
	"github.com/titpetric/phpx/model"
)

// staticStorageFor answers the untagged build: each class keeps its own copy.
func (rt *Runtime) staticStorageFor(class, _ string, scope *Scope) (map[string]any, error) {
	return rt.staticStorage(class, scope)
}

// phpxStaticName answers the untagged build, where the literal name stands.
func (rt *Runtime) phpxStaticName(tgt *model.StaticProp, _ *Scope) (string, error) {
	return tgt.Name, nil
}
