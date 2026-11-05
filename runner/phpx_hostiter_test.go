package runner

import (
	"testing"
)

type countIter struct{ n, max int }

func (c *countIter) Rewind() error         { c.n = 0; return nil }
func (c *countIter) Valid() (bool, error)  { return c.n < c.max, nil }
func (c *countIter) Current() (any, error) { return c.n * 10, nil }
func (c *countIter) Key() (any, error)     { return c.n, nil }
func (c *countIter) Next() error           { c.n++; return nil }

func TestPhpxForeachHost(t *testing.T) {
	t.Parallel()
	var got []any
	_ = phpxForeachHost(&countIter{max: 3}, func(_, v any) bool { got = append(got, v); return true })
	if len(got) != 3 || got[2] != 20 {
		t.Fatalf("got %v", got)
	}
}
