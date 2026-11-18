//go:build phpx

package phpx

import (
	"testing"

	"github.com/titpetric/phpx/runner"
)

func TestPDOConstants(t *testing.T) {
	checks := map[string]any{
		"ATTR_ERRMODE":      int64(3),
		"ATTR_DRIVER_NAME":  int64(16),
		"ERRMODE_EXCEPTION": int64(2),
		"FETCH_ASSOC":       int64(2),
		"FETCH_OBJ":         int64(5),
		"FETCH_KEY_PAIR":    int64(12),
		"PARAM_INT":         int64(1),
		"PARAM_STR":         int64(2),
		"ERR_NONE":          "00000",
	}
	for name, want := range checks {
		if got := pdoConsts[name]; got != want {
			t.Errorf("%s = %v, want %v", name, got, want)
		}
	}
	if len(pdoConsts) != 103 {
		t.Errorf("constants = %d, want 103", len(pdoConsts))
	}
}

func TestPDOClassConstantsResolve(t *testing.T) {
	installPDOClass()
	rt := runner.New(nil, runner.Options{})
	if got, err := rt.PhpxClassConst("PDO", "FETCH_OBJ"); err != nil || got != int64(5) {
		t.Fatalf("PDO::FETCH_OBJ = %v, %v", got, err)
	}
	if got, err := rt.PhpxClassConst("PDO", "ERR_NONE"); err != nil || got != "00000" {
		t.Fatalf("PDO::ERR_NONE = %v, %v", got, err)
	}
}
