package coverage_test

import (
	"io"
	"strings"
	"testing"

	"github.com/titpetric/phpx/runner/coverage"
)

var profileBlocks = []coverage.Block{
	{File: "app.php", StartLine: 3, EndLine: 3, NumStmt: 1, Count: 0},
	{File: "app.php", StartLine: 2, EndLine: 2, NumStmt: 1, Count: 1},
	{File: "lib.php", StartLine: 2, EndLine: 2, NumStmt: 2, Count: 1},
}

var profileSource = map[string]string{
	"app.php": "<?php\n  echo \"a\";\n  echo \"b\";\n",
	"lib.php": "<?php\nfunction f() { return 1; }\n",
}

func testSource(file string) []string {
	src, ok := profileSource[file]
	if !ok {
		return nil
	}
	return strings.Split(src, "\n")
}

// TestColumns covers the half the collector cannot answer: it records line
// ranges. A profile lists columns spanning the statement text and not the
// indentation around it.
func TestColumns(t *testing.T) {
	t.Parallel()
	blocks := coverage.Columns(profileBlocks, testSource)
	if len(blocks) != 3 {
		t.Fatalf("blocks = %d, want 3", len(blocks))
	}
	// Sorted by file, then position, the order a profile lists them in.
	if blocks[0].File != "app.php" || blocks[0].StartLine != 2 {
		t.Errorf("first block = %+v, want app.php line 2", blocks[0])
	}
	// `  echo "a";` sits two spaces in and runs eleven characters.
	if blocks[0].StartCol != 3 || blocks[0].EndCol != 12 {
		t.Errorf("columns = %d..%d, want 3..12", blocks[0].StartCol, blocks[0].EndCol)
	}

	t.Run("unreadable source", func(t *testing.T) {
		blocks := coverage.Columns(profileBlocks, func(string) []string { return nil })
		for _, block := range blocks {
			if block.StartCol != 1 || block.EndCol != 1 {
				t.Errorf("block %+v, want column 1 on both ends", block)
			}
		}
	})
}

// TestWriteProfile pins the format go test writes and go tool cover reads.
func TestWriteProfile(t *testing.T) {
	t.Parallel()
	var out strings.Builder
	if err := coverage.WriteProfile(&out, coverage.Columns(profileBlocks, testSource), coverage.DefaultMode); err != nil {
		t.Fatalf("WriteProfile: %v", err)
	}
	want := "mode: " + string(coverage.DefaultMode) + "\n" +
		"app.php:2.3,2.12 1 1\n" +
		"app.php:3.3,3.12 1 0\n" +
		"lib.php:2.1,2.27 2 1\n"
	if out.String() != want {
		t.Errorf("profile =\n%s\nwant\n%s", out.String(), want)
	}
}

// TestWriteProfileAtomicCapsCount pins that an atomic profile carries the
// mark, not a sum. The test command merges per-fixture profiles, and a shared
// include two fixtures ran would otherwise write count 2 under mode atomic.
func TestWriteProfileAtomicCapsCount(t *testing.T) {
	t.Parallel()
	blocks := []coverage.ProfileBlock{{File: "app.php", StartLine: 2, StartCol: 1, EndLine: 2, EndCol: 4, NumStmt: 1, Count: 3}}
	var out strings.Builder
	if err := coverage.WriteProfile(&out, blocks, coverage.ModeAtomic); err != nil {
		t.Fatalf("WriteProfile: %v", err)
	}
	if want := "app.php:2.1,2.4 1 1\n"; !strings.HasSuffix(out.String(), want) {
		t.Errorf("atomic profile = %q, want the mark", out.String())
	}
	out.Reset()
	if err := coverage.WriteProfile(&out, blocks, coverage.ModeCount); err != nil {
		t.Fatalf("WriteProfile: %v", err)
	}
	if want := "app.php:2.1,2.4 1 3\n"; !strings.HasSuffix(out.String(), want) {
		t.Errorf("count profile = %q, want the count", out.String())
	}
}

// TestPercent covers the statement-weighted percentage go test reports.
func TestPercent(t *testing.T) {
	t.Parallel()
	blocks := coverage.Columns(profileBlocks, testSource)
	// Three of four statements ran: app.php line 2 and lib.php's pair.
	if got := coverage.Percent(blocks); got != 75 {
		t.Errorf("Percent = %v, want 75", got)
	}
	if got := coverage.Percent(nil); got != 0 {
		t.Errorf("Percent(nil) = %v, want 0", got)
	}
}

// BenchmarkWriteProfile measures writing a profile the size of a real suite:
// one line per block, tens of thousands of them.
func BenchmarkWriteProfile(b *testing.B) {
	blocks := make([]coverage.ProfileBlock, 40000)
	for i := range blocks {
		blocks[i] = coverage.ProfileBlock{
			File:      "vendor/phpunit/phpunit/src/Framework/TestCase.php",
			StartLine: i + 1,
			StartCol:  9,
			EndLine:   i + 1,
			EndCol:    42,
			NumStmt:   1,
			Count:     i % 2,
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if err := coverage.WriteProfile(io.Discard, blocks, coverage.ModeAtomic); err != nil {
			b.Fatal(err)
		}
	}
}
