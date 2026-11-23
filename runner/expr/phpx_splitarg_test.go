package expr

import (
	"testing"

	"github.com/titpetric/phpx/model"
)

// TestPhpxSplitArgFallsBackWithoutHelper covers the untagged build, where no
// __phpx_splitbase helper is installed and the argument is a value.
func TestPhpxSplitArgFallsBackWithoutHelper(t *testing.T) {
	dc := &directCompiler{}
	c, ok := dc.phpxSplitArg(&model.Lit{Value: int64(7)})
	if !ok {
		t.Fatal("a literal argument did not compile")
	}
	got, err := c(&Env{})
	if err != nil {
		t.Fatalf("split argument: %v", err)
	}
	if got != int64(7) {
		t.Errorf("argument read %#v, want the literal's value", got)
	}
}

// TestPhpxSplitArgUsesTheHelper covers the phpx build, where the helper
// resolves the argument's container and answers the value it should bind.
func TestPhpxSplitArgUsesTheHelper(t *testing.T) {
	dc := &directCompiler{}
	node := &model.Lit{Value: int64(7)}
	c, ok := dc.phpxSplitArg(node)
	if !ok {
		t.Fatal("a literal argument did not compile")
	}
	var seen any
	env := &Env{Base: map[string]any{
		"__phpx_splitbase": func(args ...any) (any, error) {
			seen = args[0]
			return "split", nil
		},
	}}
	got, err := c(env)
	if err != nil {
		t.Fatalf("split argument: %v", err)
	}
	if got != "split" {
		t.Errorf("argument read %#v, want the helper's answer", got)
	}
	if seen != node {
		t.Errorf("the helper resolved %#v, want the argument node", seen)
	}
}
