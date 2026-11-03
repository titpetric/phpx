package expr

import (
	"errors"
	"testing"
)

type quietErr struct{}

func (quietErr) Error() string { return "refused" }
func (quietErr) Quiet() bool   { return true }

// TestQuietOnRefusal keeps an error that marks itself Quiet from reaching
// isset, and lets any other error through.
func TestQuietOnRefusal(t *testing.T) {
	quiet := quietOnRefusal(func(*Env) (any, error) { return 1, quietErr{} })
	if v, err := quiet(nil); v != nil || err != nil {
		t.Fatalf("quiet: %v, %v", v, err)
	}
	plain := errors.New("boom")
	loud := quietOnRefusal(func(*Env) (any, error) { return nil, plain })
	if _, err := loud(nil); !errors.Is(err, plain) {
		t.Fatalf("loud: %v", err)
	}
	if !isQuietCall("ISSET") || isQuietCall("count") {
		t.Fatal("isQuietCall")
	}
}
