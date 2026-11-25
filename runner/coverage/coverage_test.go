package coverage_test

import (
	"testing"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/parser"
	"github.com/titpetric/phpx/runner/coverage"
)

// collect returns a collector holding one registration of src under filename,
// with count hits reported for each registered statement.
func collect(t *testing.T, filename, src string, count int) *coverage.Collector {
	t.Helper()
	program, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	c := coverage.New()
	c.Register(filename, program)
	for statement := range program.SourceSpans {
		for range count {
			c.Hit(statement)
		}
	}
	return c
}

const collectorSource = `<?php
$a = 1;
echo $a;
function helper() { return 2; }
class Thing { public function method() { return 3; } }
interface Shape {}
`

// TestNew covers the state of a collector after one registration: every
// executable statement seeded, declarations skipped, and the file recorded
// even when nothing in it ran.
func TestNew(t *testing.T) {
	t.Parallel()
	c := collect(t, "app.php", collectorSource, 0)

	if files := c.Files(); len(files) != 1 || files[0] != "app.php" {
		t.Errorf("files = %v, want app.php", files)
	}
	for _, block := range c.Blocks() {
		if block.Count != 0 {
			t.Errorf("line %d count = %d, want the zero baseline", block.StartLine, block.Count)
		}
	}
	// The declarations are not executable: the class and the interface seed no
	// block at all, only the method body inside the class does.
	if got := len(c.Blocks()); got != 4 {
		t.Errorf("blocks = %+v, want the four executable statements", c.Blocks())
	}
}

// TestCollector covers the mark: a statement reached twice is covered once;
// one that nothing reached stays at zero, the row a report exists to show.
func TestCollector(t *testing.T) {
	t.Parallel()
	for _, hits := range []int{1, 2} {
		c := collect(t, "app.php", collectorSource, hits)
		for _, block := range c.Blocks() {
			if block.Count != 1 {
				t.Errorf("%d hits: line %d count = %d, want the covered mark", hits, block.StartLine, block.Count)
			}
		}
	}

	empty := collect(t, "app.php", collectorSource, 0)
	if got := coverage.Percent(coverage.Columns(empty.Blocks(), func(string) []string { return nil })); got != 0 {
		t.Errorf("uncovered percent = %v, want 0", got)
	}
}

// TestCollector_Functions covers the declaration spans a per-function report
// charges against: free functions by their name, methods by Class::name.
// An interface declares only signatures and spans nothing.
func TestCollector_Functions(t *testing.T) {
	t.Parallel()
	funcs := collect(t, "app.php", collectorSource, 1).Functions()

	names := map[string]coverage.FuncSpan{}
	for _, fn := range funcs {
		names[fn.Name] = fn
	}
	for _, want := range []string{"helper", "Thing::method"} {
		fn, ok := names[want]
		if !ok {
			t.Errorf("%s is not registered, got %v", want, funcs)
			continue
		}
		if fn.File != "app.php" || fn.StartLine < 1 || fn.EndLine < fn.StartLine {
			t.Errorf("%s span = %+v", want, fn)
		}
	}
	if len(funcs) != 2 {
		t.Errorf("functions = %+v, want the two declarations with a body", funcs)
	}
}

// TestCollector_Register pins that re-registering a cached program keeps the
// marks it already has: a server includes the same file every request. A
// reset baseline would report the previous run in place of the current one.
func TestCollector_Register(t *testing.T) {
	t.Parallel()
	program, err := parser.Parse(collectorSource)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	c := coverage.New()
	c.Register("app.php", program)
	for statement := range program.SourceSpans {
		c.Hit(statement)
	}
	c.Register("app.php", program)

	for _, block := range c.Blocks() {
		if block.Count != 1 {
			t.Errorf("line %d count = %d after a second Register, want 1", block.StartLine, block.Count)
		}
	}
}

// TestName covers the two spellings a collector sees for one file: the
// interpreter anchors an entrypoint at the source root for __FILE__, and names
// an include as it resolved it.
func TestName(t *testing.T) {
	t.Parallel()
	for _, tc := range [][2]string{
		{"/app.php", "app.php"},
		{"app.php", "app.php"},
		{"public/index.php", "public/index.php"},
		{"", ""},
	} {
		if got := coverage.Name(tc[0]); got != tc[1] {
			t.Errorf("Name(%q) = %q, want %q", tc[0], got, tc[1])
		}
	}

	a := coverage.NewAggregator()
	a.Add(collect(t, "/app.php", aggregatorSource, 1))
	a.Add(collect(t, "app.php", aggregatorSource, 1))
	if files := a.Files(); len(files) != 1 {
		t.Errorf("files = %v, want the two spellings folded into one", files)
	}
}

// benchSource holds a loop body long enough that a statement costs the same
// whether it is the first or the tenth statement of the body.
const benchSource = `<?php
function hot($n) {
	$a = 1;
	$b = 2;
	$c = 3;
	$d = 4;
	$e = 5;
	$f = 6;
	return $n + $a + $b + $c + $d + $e + $f;
}
for ($i = 0; $i < 3; $i++) {
	hot($i);
}
`

// benchmarkCollector parses benchSource into a collector and returns the
// statements it registered.
func benchmarkCollector(b *testing.B) (*coverage.Collector, []model.Stmt) {
	b.Helper()
	program, err := parser.Parse(benchSource)
	if err != nil {
		b.Fatalf("parse: %v", err)
	}
	c := coverage.New()
	c.Register("bench.php", program)
	stmts := make([]model.Stmt, 0, len(program.SourceSpans))
	for stmt := range program.SourceSpans {
		stmts = append(stmts, stmt)
	}
	return c, stmts
}

// BenchmarkCollectorHit measures the hot path: the interpreter reports a
// statement it has reported before, which is every statement after a loop's
// first iteration.
func BenchmarkCollectorHit(b *testing.B) {
	c, stmts := benchmarkCollector(b)
	if len(stmts) == 0 {
		b.Fatal("no statements registered")
	}
	stmt := stmts[0]
	c.Hit(stmt)
	b.ResetTimer()
	for range b.N {
		c.Hit(stmt)
	}
}

// BenchmarkCollectorHitMany measures the same path across the statements of a
// loop body, so the map lookup moves across keys and not one hot entry.
func BenchmarkCollectorHitMany(b *testing.B) {
	c, stmts := benchmarkCollector(b)
	if len(stmts) == 0 {
		b.Fatal("no statements registered")
	}
	for _, stmt := range stmts {
		c.Hit(stmt)
	}
	b.ResetTimer()
	for i := range b.N {
		c.Hit(stmts[i%len(stmts)])
	}
}

// BenchmarkCollectorBlocks measures rendering the profile once, after the run:
// the map it merges statements by and the sort it orders blocks with.
func BenchmarkCollectorBlocks(b *testing.B) {
	c, stmts := benchmarkCollector(b)
	for _, stmt := range stmts {
		c.Hit(stmt)
	}
	b.ResetTimer()
	for range b.N {
		if blocks := c.Blocks(); len(blocks) == 0 {
			b.Fatal("no blocks")
		}
	}
}
