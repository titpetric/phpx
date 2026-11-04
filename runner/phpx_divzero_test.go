//go:build phpx

package runner

import "testing"

// TestPhpxDivZero pins which divisors raise and the message each operator carries.
func TestPhpxDivZero(t *testing.T) {
	t.Parallel()
	cases := []struct {
		op   string
		b    any
		want string
	}{
		{"/", int64(0), "Division by zero"},
		{"/", 0.0, "Division by zero"},
		{"/", "0", "Division by zero"},
		{"%", int64(0), "Modulo by zero"},
		{"%", 0.5, "Modulo by zero"},
		{"/", 0.5, ""},
		{"%", int64(2), ""},
		{"+", int64(0), ""},
	}
	for _, c := range cases {
		err := phpxDivZero(c.op, int64(1), c.b)
		got := ""
		if err != nil {
			got = err.Error()
		}
		if got != c.want {
			t.Errorf("%s %v: %q, want %q", c.op, c.b, got, c.want)
		}
	}
}
