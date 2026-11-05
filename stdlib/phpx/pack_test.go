//go:build phpx

package phpx

import "testing"

func TestPack(t *testing.T) {
	cases := []struct {
		format string
		args   []any
		want   string
	}{
		{"n", []any{int64(0x1234)}, "\x12\x34"},
		{"n*", []any{int64(1), int64(2)}, "\x00\x01\x00\x02"},
		{"v", []any{int64(0x1234)}, "\x34\x12"},
		{"N", []any{int64(1)}, "\x00\x00\x00\x01"},
		{"J", []any{int64(1)}, "\x00\x00\x00\x00\x00\x00\x00\x01"},
		{"a4", []any{"ab"}, "ab\x00\x00"},
		{"A4", []any{"ab"}, "ab  "},
		{"Z*", []any{"ab"}, "ab\x00"},
		{"H*", []any{"deadbeef"}, "\xde\xad\xbe\xef"},
		{"Cx2C", []any{int64(1), int64(2)}, "\x01\x00\x00\x02"},
		{"E", []any{1.0}, "\x3f\xf0\x00\x00\x00\x00\x00\x00"},
	}
	for _, c := range cases {
		got, err := pack(c.format, c.args)
		if err != nil || got != c.want {
			t.Errorf("pack(%q) = %q, %v; want %q", c.format, got, err, c.want)
		}
	}
	if _, err := pack("n", nil); err == nil {
		t.Error("too few arguments should fail")
	}
	if _, err := pack("y", nil); err == nil {
		t.Error("unknown code should fail")
	}
}

func TestUnpackNative(t *testing.T) {
	if got := unpackNative('l', []byte{0xff, 0xff, 0xff, 0xff}); got != int64(-1) {
		t.Errorf("l = %v", got)
	}
	if got := unpackNative('E', []byte{0x3f, 0xf0, 0, 0, 0, 0, 0, 0}); got != 1.0 {
		t.Errorf("E = %v", got)
	}
	if got := unpackNative('S', []byte{0x34, 0x12}); got != int64(0x1234) {
		t.Errorf("S = %v", got)
	}
}
