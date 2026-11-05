//go:build phpx

package runner

import (
	"os"
	"testing"

	"github.com/titpetric/phpx/internal/phpval"
)

func TestResourceHookInstalled(t *testing.T) {
	h := os.Stdout
	if got := phpval.String(h); got != phpxResourceString(h) {
		t.Fatalf("phpval.String = %q", got)
	}
}
