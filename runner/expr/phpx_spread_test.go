package expr

import (
	"testing"
)

func TestNamedArgCarriesItsName(t *testing.T) {
	t.Parallel()
	arg := NamedArg{Name: "limit", Val: 3}
	if arg.Name != "limit" || arg.Val != 3 {
		t.Fatalf("got %+v", arg)
	}
	if !sameArgs(nil, nil) {
		t.Fatal("two empty argument lists are the same list")
	}
}
