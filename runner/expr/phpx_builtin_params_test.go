package expr

import (
	"testing"
)

func TestBuiltinParamsNameTheThrowables(t *testing.T) {
	t.Parallel()
	got := builtinParams["exception"]
	if len(got) != 3 || got[2] != "previous" {
		t.Fatalf("got %v, want message, code, previous", got)
	}
}
