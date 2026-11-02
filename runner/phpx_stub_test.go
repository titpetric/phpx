//go:build !phpx

package runner

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// The untagged build is phpscript, and this pins that: a fork commit stays
// reviewable against upstream only while an untagged build answers what
// upstream answers.
func TestPhpxGlobalBindsNothingWithoutTheTag(t *testing.T) {
	t.Parallel()
	globals := NewScope()
	globals.vars["counter"] = 1
	rt := &Runtime{frames: []*Scope{globals}}
	local := NewScope()

	if err := rt.phpxGlobal(&model.Global{Names: []string{"counter"}}, local); err != nil {
		t.Fatalf("phpxGlobal returned an error: %v", err)
	}
	if _, ok := local.Get("counter"); ok {
		t.Error("the untagged build bound a name the statement names")
	}
}
