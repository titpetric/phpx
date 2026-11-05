//go:build phpx

package phpx

import (
	"reflect"
	"testing"
)

func TestCsvRecord(t *testing.T) {
	cases := []struct {
		in   string
		want []any
	}{
		{"a,b,c\n", []any{"a", "b", "c"}},
		{"a, b ,\"c d\"", []any{"a", " b ", "c d"}},
		{"\"a\"\"b\",c", []any{"a\"b", "c"}},
		{"\"a\\\"b\",c", []any{"a\\\"b", "c"}},
		{"a,,b", []any{"a", "", "b"}},
		{"a,", []any{"a", ""}},
		{"\n", []any{nil}},
		{"  \"x\" ,y", []any{"x ", "y"}},
	}
	for _, c := range cases {
		if got := csvRecord(c.in, nil, ',', '"', '\\'); !reflect.DeepEqual(got, c.want) {
			t.Errorf("csvRecord(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}
}

func TestCsvRecordMultiline(t *testing.T) {
	lines := []string{"b\",c\n"}
	more := func() (string, bool) {
		if len(lines) == 0 {
			return "", false
		}
		l := lines[0]
		lines = lines[1:]
		return l, true
	}
	got := csvRecord("a,\"x\n", more, ',', '"', '\\')
	want := []any{"a", "x\nb", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
