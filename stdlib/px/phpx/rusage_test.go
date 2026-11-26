//go:build phpx && !windows

package phpx

import "testing"

// TestPhpxRusage pins the record getrusage() answers: a real reading with
// php's keys, and false for a selector the host refuses.
func TestPhpxRusage(t *testing.T) {
	v := phpxRusage(rusageSelf)
	arr, ok := v.(interface {
		Get(any) (any, bool)
	})
	if !ok {
		t.Fatalf("getrusage(RUSAGE_SELF) = %T, want an array", v)
	}
	for _, key := range []string{"ru_utime.tv_sec", "ru_utime.tv_usec", "ru_stime.tv_sec", "ru_stime.tv_usec", "ru_maxrss"} {
		if _, ok := arr.Get(key); !ok {
			t.Errorf("getrusage() has no %q", key)
		}
	}
	if got := phpxRusage(99); got != false {
		t.Errorf("getrusage(99) = %v, want false", got)
	}
}
