package runner

import (
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/titpetric/phpx/model"
)

// memorySink keeps a measured figure reachable, so the compiler cannot drop it.
var memorySink int64

// TestPhpxRealMemoryUsage covers the real form of memory_get_usage(): the bytes
// the Go heap has taken from the system, which is the quantity php's
// $real_usage names for its allocator. The figure is at least the live estimate
// and never below the peak already read.
func TestPhpxRealMemoryUsage(t *testing.T) {
	t.Parallel()
	rt := New(io.Discard, Options{})

	real := rt.RealMemoryUsage()
	if real <= 0 {
		t.Fatalf("RealMemoryUsage() = %d, want the heap's system bytes", real)
	}
	if used := rt.MemoryUsage(); real < used {
		t.Errorf("RealMemoryUsage() = %d, below the %d in use", real, used)
	}
	if peak := rt.RealMemoryPeak(); peak < real {
		t.Errorf("RealMemoryPeak() = %d, below RealMemoryUsage()'s %d", peak, real)
	}

	rt.globals["big"] = strings.Repeat("x", 1<<16)
	if got, used := rt.RealMemoryUsage(), rt.MemoryUsage(); got < used {
		t.Errorf("RealMemoryUsage() = %d after 64 KiB, below the %d in use", got, used)
	}
}

// TestPhpxMemoryWalkReusesVisitedSet covers the visited set a walk keeps: a
// repeated walk reuses the two maps instead of allocating them, and a walk that
// saw more containers than memVisitKeep drops them rather than keeping the
// buckets a large heap made.
func TestPhpxMemoryWalkReusesVisitedSet(t *testing.T) {
	rt := New(io.Discard, Options{})

	rt.MemoryWalk()
	if got := testing.AllocsPerRun(10, func() { memorySink = rt.MemoryWalk() }); got != 0 {
		t.Errorf("MemoryWalk allocated %v times after the first walk, want 0", got)
	}

	for i := 0; i <= memVisitKeep; i++ {
		item := model.NewArray()
		item.Append(i)
		rt.globals["big"+strconv.Itoa(i)] = item
	}
	rt.MemoryWalk()
	if rt.memVisit.arrays != nil || rt.memVisit.objects != nil {
		t.Errorf("a walk over %d arrays kept its visited maps", memVisitKeep+1)
	}
}

// BenchmarkPhpxRealMemoryUsage measures the read behind memory_get_usage(true),
// which a queue worker calls once per job.
func BenchmarkPhpxRealMemoryUsage(b *testing.B) {
	rt := New(io.Discard, Options{})
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		memorySink = rt.RealMemoryUsage()
	}
}

// BenchmarkPhpxRealMemoryUsageReadMemStats measures the fallback path, which
// stops the world and answers the same figure.
func BenchmarkPhpxRealMemoryUsageReadMemStats(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		memorySink = int64(realUsageFallback())
	}
}

// BenchmarkPhpxMemoryWalk measures a walk over a thousand live arrays.
func BenchmarkPhpxMemoryWalk(b *testing.B) {
	rt := New(io.Discard, Options{})
	for i := 0; i < 1000; i++ {
		item := model.NewArray()
		for j := 0; j < 4; j++ {
			item.Append(j)
		}
		rt.globals["k"+strconv.Itoa(i)] = item
	}
	rt.MemoryWalk() // the first walk makes the visited set the rest reuse
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memorySink = rt.MemoryWalk()
	}
}
