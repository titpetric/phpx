package model

// Class is the resolved, runnable form of a ClassDecl: field defaults plus a
// method table keyed by method name.
//
// Statics are the declarations of `static $name = expr` properties; their
// values live in the runtime (one bag per class, created on first access) so
// that every instance and every static call observes the same storage.
type Class struct {
	Name string
	// Implements is every interface name the declaration listed, plus the names
	// those interfaces extend, lower-cased. It records the contracts the class
	// checks against, and is what `instanceof` answers an interface name from. No member arrives through it; see docs/design.md.
	Implements []string
	Fields     []Field
	Statics    []Field // static property declarations (Name + default Expr)
	Consts     []Field // class constants (Name + value Expr)
	Methods    map[string]*FuncDecl
	// Ancestors is the chain `extends` names, nearest first: Ancestors[0] is
	// the direct parent, and the last is the root of the chain. It is a list
	// of names and not storage; the members themselves were flattened into this
	// class when it was hoisted (runner/phpx.go in the fork).
	Ancestors []string
	// MethodOwner maps a method name to the class that declared it, so `self::`
	// inside an inherited body resolves where php resolves it. A class that
	// declares its own method owns no entry different from its name.
	MethodOwner map[string]string
}

// declares reports whether name is one of the class's declared fields.
func (c *Class) declares(name string) bool {
	for _, field := range c.Fields {
		if field.Name == name {
			return true
		}
	}
	return false
}

// Field returns the declaration of the named field, for a caller that needs the
// visibility or type a property declares.
func (c *Class) Field(name string) (Field, bool) {
	for _, field := range c.Fields {
		if field.Name == name {
			return field, true
		}
	}
	return Field{}, false
}
