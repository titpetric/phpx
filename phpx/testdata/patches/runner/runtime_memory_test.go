package runner

import (
	"io"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

// TestRuntime_CompiledTree covers the cost of the parsed tree, which phpinfo()
// reports and memory_limit does not charge for: every request shares the tree,
// and the tree outlives each request. The heap figure is zero until a
// precompile pass measures it. Printing 0.00 MiB across a filled cache would
// read as a measurement, not as an absent one.
func TestRuntime_CompiledTree(t *testing.T) {
	t.Parallel()
	rt := New(io.Discard, Options{
		Precompile: true,
		RootFS: fstest.MapFS{
			"lib.php": &fstest.MapFile{Data: []byte(`<?php function helper($a) { return $a * 2; }`)},
		},
	})

	programs, expressions, heap := rt.CompiledTree()
	if programs != 0 || expressions != 0 {
		t.Fatalf("CompiledTree = %d programs, %d expressions on a runtime that loaded nothing", programs, expressions)
	}
	if heap != 0 {
		t.Errorf("CompiledTree heap = %d with no precompile pass behind it, want 0", heap)
	}

	if _, err := rt.LoadFile("lib.php"); err != nil {
		t.Fatal(err)
	}
	if programs, _, _ = rt.CompiledTree(); programs != 1 {
		t.Errorf("CompiledTree = %d programs after one load, want 1", programs)
	}
}

// TestRuntimeMemoryReporting covers the figures memory_get_usage() and
// memory_get_peak_usage() answer with. Both come from a walk of the live roots.
// The usage is the reachable total now; the peak is the highest a walk has
// seen. An allocation made and released between two walks does not raise the
// peak, while php's allocator-level peak counts it.
func TestRuntimeMemoryReporting(t *testing.T) {
	t.Parallel()
	rt := New(io.Discard, Options{})

	base := rt.MemoryUsage()
	if base <= 0 {
		t.Fatalf("MemoryUsage() = %d, want the baseline a runtime starts at", base)
	}
	if got := rt.MemoryWalk(); got != base {
		t.Errorf("MemoryWalk() = %d against MemoryUsage()'s %d, want one figure", got, base)
	}
	if got := rt.MemoryPeak(); got < base {
		t.Errorf("MemoryPeak() = %d, below the %d in use", got, base)
	}

	// A host folds the request's own cost into the baseline and the walk starts
	// from what the request already weighs.
	rt.AccountRequest(strings.Repeat("x", 4096))
	grown := rt.MemoryWalk()
	if grown <= base {
		t.Errorf("MemoryWalk() = %d after accounting a 4k value, want more than the %d baseline", grown, base)
	}

	// The peak is at least what the last walk saw.
	peak := rt.MemoryPeak()
	if peak < grown {
		t.Errorf("MemoryPeak() = %d, below the %d the last walk measured", peak, grown)
	}

	// A reset takes it back to the baseline, because the figure belonged to the
	// program that ran: a pooled runtime reporting the previous script's peak
	// would charge it to a page that never allocated any of it.
	rt.Reset(io.Discard, nil)
	if got := rt.MemoryPeak(); got >= peak {
		t.Errorf("MemoryPeak() = %d after a reset, want it back below the %d the last program reached", got, peak)
	}
}

// TestRuntimeMemorySamplesExpensiveWalk covers the bound on a walk that a host
// reads per event. The walk is linear in the live heap, so an expensive one is
// reused for a window instead of repeated; a cheap one is exact on every read,
// which keeps a short script's figures what php's are.
func TestRuntimeMemorySamplesExpensiveWalk(t *testing.T) {
	t.Parallel()
	rt := New(io.Discard, Options{})

	cached := runtimeBaseline + 1
	rt.memWalkCost = 2 * memSampleCost
	rt.memWalkAt = time.Now()
	rt.memUsage = cached
	rt.globals["big"] = strings.Repeat("x", 8192)

	if got := rt.MemoryUsage(); got != cached {
		t.Fatalf("MemoryUsage() = %d inside the sample window, want the cached %d", got, cached)
	}

	rt.memWalkAt = time.Now().Add(-time.Second)
	if got := rt.MemoryUsage(); got <= cached {
		t.Fatalf("MemoryUsage() = %d after the window, want a fresh walk above %d", got, cached)
	}

	// A cheap walk is not sampled: two reads with an allocation between them
	// see the allocation.
	rt.memWalkCost = 0
	before := rt.MemoryUsage()
	rt.globals["more"] = strings.Repeat("y", 8192)
	if after := rt.MemoryUsage(); after <= before {
		t.Errorf("MemoryUsage() = %d after an 8k allocation against %d before, want it to grow", after, before)
	}
}
