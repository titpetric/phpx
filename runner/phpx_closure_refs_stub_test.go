//go:build !phpx

package runner

import (
	"testing"
)

func TestPhpxSyncRefsStub(t *testing.T) {
	t.Parallel()
	New(nil, Options{}).phpxSyncRefs(closureEnv{}, nil)()
}
