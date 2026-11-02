//go:build phpx

package runner

import (
	"fmt"
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxAbstractError is the fatal php raises when a class declares itself
// concrete and leaves a method without a body. The count and the names are
// php's, in php's order, and the singular reads "method" as php writes it.
func phpxAbstractError(class string, methods []string) error {
	noun := "methods"
	if len(methods) == 1 {
		noun = "method"
	}
	return fmt.Errorf("Class %s contains %d abstract %s and must therefore be declared abstract or implement the remaining methods (%s)",
		class, len(methods), noun, strings.Join(methods, ", "))
}

// phpxOwnAbstract names the abstract methods a declaration writes itself, in
// declaration order. php raises this list before traits bind and a parent's
// members arrive, so a class that declares one itself hears about that one
// first and hears about no other.
func phpxOwnAbstract(cd *model.ClassDecl) []string {
	var names []string
	for _, m := range cd.Methods {
		if m.Abstract {
			names = append(names, phpxMethodRef(cd.Name, m.Name))
		}
	}
	return names
}

// phpxMethodRef spells a method the way php's message does.
func phpxMethodRef(class, method string) string {
	return class + "::" + method
}

// phpxResolveAbstract fills in a class's abstract flag and the methods it
// leaves unimplemented. own, inherited and trait arrive in php's order; a name
// repeated across them is kept once, and a name a concrete method now answers
// falls out. A class that is not abstract and keeps a name is the fatal.
func phpxResolveAbstract(c *model.Class, cd *model.ClassDecl, own, inherited, trait []string) error {
	c.Abstract = cd.Abstract
	var seen []string
	remaining := make([]string, 0, len(own)+len(inherited)+len(trait))
	for _, group := range [][]string{own, inherited, trait} {
		for _, entry := range group {
			name := phpxMethodName(entry)
			decl, ok := phpxFieldFold(c.Methods, name)
			if !ok || !decl.Abstract || phpxAbstractSeen(seen, name) {
				continue
			}
			seen = append(seen, name)
			remaining = append(remaining, phpxMethodRef(c.MethodOwner[decl.Name], decl.Name))
		}
	}
	c.AbstractMethods = remaining
	if !c.Abstract && len(remaining) > 0 {
		return phpxAbstractError(c.Name, remaining)
	}
	return nil
}

// phpxMethodName reads the method name out of a "Class::name" reference.
func phpxMethodName(ref string) string {
	if i := strings.LastIndex(ref, "::"); i >= 0 {
		return ref[i+2:]
	}
	return ref
}

// phpxAbstractSeen reports whether a name already sits in the list, comparing
// case-insensitively as php compares method names.
func phpxAbstractSeen(seen []string, name string) bool {
	for _, have := range seen {
		if strings.EqualFold(have, name) {
			return true
		}
	}
	return false
}

// phpxCallAbstract refuses a call to a method with no body. php reaches this
// through parent:: from a child that implemented the method, because a call on
// the class itself never resolves to an abstract declaration.
//
// The declaring class is resolved only for the fatal: every ordinary call pays
// one nil check and one bool, and the class-table scan sits behind them.
func (rt *Runtime) phpxCallAbstract(class *model.Class, decl *model.FuncDecl) error {
	if decl == nil || !decl.Abstract {
		return nil
	}
	return fmt.Errorf("Cannot call abstract method %s()", phpxMethodRef(rt.phpxDeclOwner(class, decl), decl.Name))
}

// phpxDeclOwner names the class that declared a method. The instance's class
// names the wrong one when a child reaches an abstract parent body through
// parent::, and the class table records the declaring class beside the method
// it flattened into each descendant.
func (rt *Runtime) phpxDeclOwner(class *model.Class, decl *model.FuncDecl) string {
	if decl != nil {
		for _, c := range rt.classes {
			method, ok := phpxFieldFold(c.Methods, decl.Name)
			if !ok || method != decl {
				continue
			}
			if owner := c.MethodOwner[decl.Name]; owner != "" {
				return owner
			}
			return c.Name
		}
	}
	return rt.phpxOwnerClass(class, decl)
}

// phpxInstantiable refuses `new` on an abstract class. php's error is the same
// for a class that declared the modifier and for one that carries an
// unimplemented method, and both arrive here with Abstract set.
func phpxInstantiable(c *model.Class) error {
	if c == nil || !c.Abstract {
		return nil
	}
	return fmt.Errorf("Cannot instantiate abstract class %s", c.Name)
}
