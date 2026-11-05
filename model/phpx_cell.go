package model

import "sync/atomic"

// Cell holds a value that two names share, which is what a php reference is: a
// variable, a property or a by-reference parameter that was bound to it reads
// and writes the one slot. An array element is not given one here.
type Cell struct {
	V any
	// Mirror, when set, receives each value written through Set: a cell made
	// around a `static` variable keeps the variable's persistent slot current.
	Mirror func(any)
}

// Set writes the value the cell holds.
func (c *Cell) Set(v any) {
	c.V = v
	if c.Mirror != nil {
		c.Mirror(v)
	}
}

// Deref answers the value a cell holds, and any other value as it is.
func Deref(v any) any {
	if c, ok := v.(*Cell); ok {
		return c.V
	}
	return v
}

// PhpxAnyByRef is set once a function, method or closure that takes a parameter
// by reference has been parsed. Until then a call passes values alone and pays
// nothing for references.
var PhpxAnyByRef atomic.Bool

// RefArg is an argument a call site could pass by reference: the value, and a
// way to make the cell behind the lvalue that wrote it. The callee that takes
// the parameter by reference makes the cell; any other callee reads the value.
type RefArg struct {
	Value any
	Make  func() *Cell
}
