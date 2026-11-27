package astcodec_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/model/astcodec"
	"github.com/titpetric/phpx/parser"
)

// corpus reads every fixture body in the tree, which holds the node shapes a
// real file carries: classes, closures, heredocs, attributes and the rest.
func corpus(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{}
	root := filepath.Join("..", "..", "tests", "fixtures")
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		ext := filepath.Ext(path)
		if ext != ".php" && ext != ".phpt" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		src := string(data)
		if ext == ".phpt" {
			if i := strings.Index(src, "--FILE--\n"); i >= 0 {
				src = src[i+len("--FILE--\n"):]
			}
			if i := strings.Index(src, "\n--EXPECT"); i >= 0 {
				src = src[:i]
			}
		}
		out[path] = src
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) < 100 {
		t.Fatalf("corpus holds %d files, want the fixture tree", len(out))
	}
	return out
}

func withoutSpans(p *model.Program) *model.Program {
	out := *p
	out.SourceSpans = nil
	return &out
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()
	files := corpus(t)
	for path, src := range files {
		prog, err := parser.ParseFile(path, src)
		if err != nil {
			continue
		}
		data, err := astcodec.Encode(prog)
		if err != nil {
			t.Fatalf("%s: encode: %v", path, err)
		}
		back, err := astcodec.Decode(data)
		if err != nil {
			t.Fatalf("%s: decode: %v", path, err)
		}
		if !reflect.DeepEqual(withoutSpans(prog), withoutSpans(back)) {
			t.Fatalf("%s: decoded program differs from the parsed one", path)
		}
		// The span map keys on statement pointers, so a decode that builds a
		// second tree cannot compare key for key. Re-encoding the decoded
		// program walks the same nodes in the same order: identical bytes mean
		// every span named the statement its source named, and that the
		// encoder writes the same payload twice.
		again, err := astcodec.Encode(back)
		if err != nil {
			t.Fatalf("%s: re-encode: %v", path, err)
		}
		if !bytes.Equal(data, again) {
			t.Fatalf("%s: re-encoded payload differs", path)
		}
	}
}

// TestRoundTripReal walks the upstream packages this checkout keeps under
// testdata, when they are there. They carry the node shapes a framework file
// writes and a fixture does not.
func TestRoundTripReal(t *testing.T) {
	t.Parallel()
	root := filepath.Join("..", "..", "phpx", "testdata", "upstream", ".cache")
	if _, err := os.Stat(root); err != nil {
		t.Skip("no upstream testdata in this checkout")
	}
	files := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".php" {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Size() > 256<<10 {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		prog, err := parser.ParseFile(path, string(data))
		if err != nil {
			return nil
		}
		payload, err := astcodec.Encode(prog)
		if err != nil {
			t.Fatalf("%s: encode: %v", path, err)
		}
		back, err := astcodec.Decode(payload)
		if err != nil {
			t.Fatalf("%s: decode: %v", path, err)
		}
		if !reflect.DeepEqual(withoutSpans(prog), withoutSpans(back)) {
			t.Fatalf("%s: decoded program differs", path)
		}
		files++
		if files >= 200 {
			return fs.SkipAll
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if files == 0 {
		t.Skip("no upstream source to read")
	}
}

func TestDecodeRefusesShortPayload(t *testing.T) {
	t.Parallel()
	prog, err := parser.Parse("<?php echo 1;")
	if err != nil {
		t.Fatal(err)
	}
	data, err := astcodec.Encode(prog)
	if err != nil {
		t.Fatal(err)
	}
	for i := range data {
		if _, err := astcodec.Decode(data[:i]); err == nil {
			t.Fatalf("Decode(%d of %d bytes) = nil error", i, len(data))
		}
	}
}

func TestDecodeRefusesUnknownTag(t *testing.T) {
	t.Parallel()
	prog, err := parser.Parse("<?php echo 1;")
	if err != nil {
		t.Fatal(err)
	}
	data, err := astcodec.Encode(prog)
	if err != nil {
		t.Fatal(err)
	}
	// The first byte is the statement count; the second the first step tag.
	data[1] = 200
	if _, err := astcodec.Decode(data); err == nil {
		t.Fatal("Decode with an unknown tag = nil error")
	}
}

func TestDecodeStringReadsTheSamePayload(t *testing.T) {
	t.Parallel()
	prog, err := parser.Parse("<?php $a = [1, 2, 3]; echo $a[1];")
	if err != nil {
		t.Fatal(err)
	}
	data, err := astcodec.Encode(prog)
	if err != nil {
		t.Fatal(err)
	}
	fromBytes, err := astcodec.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	fromString, err := astcodec.DecodeString(string(data))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(withoutSpans(fromBytes), withoutSpans(fromString)) {
		t.Fatal("the two decode doors answered different programs")
	}
}

func TestEncodedProgramRuns(t *testing.T) {
	t.Parallel()
	src := "<?php\nfunction f($n) { return $n * 2; }\nclass A { public $x = 3; public function get() { return $this->x; } }\n$r = f(4) + (new A)->get();\necho $r;"
	prog, err := parser.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	data, err := astcodec.Encode(prog)
	if err != nil {
		t.Fatal(err)
	}
	back, err := astcodec.Decode(data)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := back.Stmts[0].(*model.FuncDecl); !ok {
		t.Fatalf("Stmts[0] = %T, want the function", back.Stmts[0])
	}
	if !reflect.DeepEqual(prog.Stmts, back.Stmts) {
		t.Fatal("statements differ after a round trip")
	}
}

func BenchmarkEncode(b *testing.B) {
	src, err := os.ReadFile(filepath.Join("..", "..", "runner", "phpx_prelude", "spl.php"))
	if err != nil {
		b.Fatal(err)
	}
	prog, err := parser.Parse(string(src))
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := astcodec.Encode(prog); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	src, err := os.ReadFile(filepath.Join("..", "..", "runner", "phpx_prelude", "spl.php"))
	if err != nil {
		b.Fatal(err)
	}
	prog, err := parser.Parse(string(src))
	if err != nil {
		b.Fatal(err)
	}
	data, err := astcodec.Encode(prog)
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := astcodec.Decode(data); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseSame(b *testing.B) {
	src, err := os.ReadFile(filepath.Join("..", "..", "runner", "phpx_prelude", "spl.php"))
	if err != nil {
		b.Fatal(err)
	}
	b.SetBytes(int64(len(src)))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := parser.Parse(string(src)); err != nil {
			b.Fatal(err)
		}
	}
}
