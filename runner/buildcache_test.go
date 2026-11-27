package runner

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func cacheFS() fstest.MapFS {
	return fstest.MapFS{
		"a.php":   &fstest.MapFile{Data: []byte("<?php\nfunction hit($n) { return $n * 2; }\necho hit(21);\n")},
		"inc.php": &fstest.MapFile{Data: []byte("<?php\nfunction twice() { return 2; }\n")},
	}
}

func cacheRuntime(t *testing.T, dir string) *Runtime {
	t.Helper()
	var out strings.Builder
	rt := New(&out, Options{RootFS: cacheFS(), BuildCache: dir, SAPI: "cli"})
	return rt
}

func TestBuildCacheMissThenHit(t *testing.T) {
	dir := t.TempDir()
	first := cacheRuntime(t, dir)
	if _, err := first.LoadFile("a.php"); err != nil {
		t.Fatal(err)
	}
	hits, misses, writes := first.BuildCacheStats()
	if hits != 0 || misses != 1 || writes != 1 {
		t.Fatalf("first load stats = %d/%d/%d, want 0/1/1", hits, misses, writes)
	}

	second := cacheRuntime(t, dir)
	prog, err := second.LoadFile("a.php")
	if err != nil {
		t.Fatal(err)
	}
	hits, misses, writes = second.BuildCacheStats()
	if hits != 1 || misses != 0 || writes != 0 {
		t.Fatalf("second load stats = %d/%d/%d, want 1/0/0", hits, misses, writes)
	}
	if len(prog.Stmts) != 2 {
		t.Fatalf("cached program holds %d statements", len(prog.Stmts))
	}
}

func TestBuildCachedProgramRuns(t *testing.T) {
	dir := t.TempDir()
	first := cacheRuntime(t, dir)
	if _, err := first.LoadFile("a.php"); err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	rt := New(&out, Options{RootFS: cacheFS(), BuildCache: dir, SAPI: "cli"})
	prog, err := rt.LoadFile("a.php")
	if err != nil {
		t.Fatal(err)
	}
	if err := rt.Run(prog); err != nil {
		t.Fatal(err)
	}
	if out.String() != "42" {
		t.Fatalf("cached program printed %q, want 42", out.String())
	}
}

func TestBuildCacheMissOnChangedSource(t *testing.T) {
	dir := t.TempDir()
	first := cacheRuntime(t, dir)
	if _, err := first.LoadFile("a.php"); err != nil {
		t.Fatal(err)
	}

	sources := cacheFS()
	sources["a.php"] = &fstest.MapFile{Data: []byte("<?php echo 1;\n")}
	var out strings.Builder
	rt := New(&out, Options{RootFS: sources, BuildCache: dir, SAPI: "cli"})
	if _, err := rt.LoadFile("a.php"); err != nil {
		t.Fatal(err)
	}
	hits, misses, writes := rt.BuildCacheStats()
	if hits != 0 || misses != 1 || writes != 1 {
		t.Fatalf("changed load stats = %d/%d/%d, want 0/1/1", hits, misses, writes)
	}
}

func TestBuildCacheSeparatesParseOptions(t *testing.T) {
	dir := t.TempDir()
	first := cacheRuntime(t, dir)
	if _, err := first.LoadFile("a.php"); err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	rt := New(&out, Options{RootFS: cacheFS(), BuildCache: dir, SAPI: "cli", Runtime: RuntimeGo})
	if _, err := rt.LoadFile("a.php"); err != nil {
		t.Fatal(err)
	}
	hits, misses, _ := rt.BuildCacheStats()
	if hits != 0 || misses != 1 {
		t.Fatalf("other options stats = %d/%d, want 0/1", hits, misses)
	}
}

func TestBuildCacheOffByDefault(t *testing.T) {
	var out strings.Builder
	rt := New(&out, Options{RootFS: cacheFS(), SAPI: "cli"})
	if _, err := rt.LoadFile("a.php"); err != nil {
		t.Fatal(err)
	}
	hits, misses, writes := rt.BuildCacheStats()
	if hits != 0 || misses != 0 || writes != 0 {
		t.Fatalf("stats = %d/%d/%d, want a disabled cache", hits, misses, writes)
	}
}

// benchSource is the prelude, a file of every ordinary shape: functions,
// classes, closures and constant tables.
func benchSource(b *testing.B) fstest.MapFS {
	b.Helper()
	src, err := os.ReadFile(filepath.Join("phpx_prelude", "spl.php"))
	if err != nil {
		b.Fatal(err)
	}
	return fstest.MapFS{"big.php": &fstest.MapFile{Data: src}}
}

// BenchmarkLoadParsed parses the file on every load, the work the cache
// replaces.
func BenchmarkLoadParsed(b *testing.B) {
	fsys := benchSource(b)
	rt := New(io.Discard, Options{RootFS: fsys, SAPI: "cli"})
	b.SetBytes(int64(len(fsys["big.php"].Data)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := rt.LoadFile("big.php"); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkLoadCached reads the program a previous load wrote.
func BenchmarkLoadCached(b *testing.B) {
	fsys := benchSource(b)
	dir := b.TempDir()
	warm := New(io.Discard, Options{RootFS: fsys, BuildCache: dir, SAPI: "cli"})
	if _, err := warm.LoadFile("big.php"); err != nil {
		b.Fatal(err)
	}
	rt := New(io.Discard, Options{RootFS: fsys, BuildCache: dir, SAPI: "cli"})
	b.SetBytes(int64(len(fsys["big.php"].Data)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := rt.LoadFile("big.php"); err != nil {
			b.Fatal(err)
		}
	}
}
