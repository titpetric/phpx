//go:build phpx

package runner

import "github.com/titpetric/phpx/internal/phpval"

func init() {
	phpval.PhpxResourceID = func(v any) (int64, bool) {
		if PhpxIsResource(v) {
			return PhpxResourceID(v), true
		}
		return 0, false
	}
}
