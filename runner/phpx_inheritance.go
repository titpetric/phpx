//go:build phpx

package runner

import (
	"fmt"
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxClass builds the runtime form of one class declaration, with the members
// `extends` carries flattened in. The child's own declaration of a field,
// static, constant or method wins over the parent's, which is php's rule and
// the one place where the merge order matters.
//
// Flattening is what keeps model.Class free of a parent field: the guard in
// runner.TestNoInheritanceAtRuntime refuses one and a parent contributes its
// members to the child's own lists rather than a pointer to itself. The chain
// stays in Ancestors for `parent::` and `instanceof`, and each method remembers
// the class that declared it so `self::` inside an inherited body resolves as
// php resolves it.
func (rt *Runtime) phpxClass(cd *model.ClassDecl, decls map[string]*model.ClassDecl, stmts []model.Stmt) (*model.Class, error) {
	return rt.phpxClassChain(cd, decls, stmts, map[string]bool{})
}

// phpxClassChain is phpxClass with the declarations already on the stack, so
// `class A extends B` and `class B extends A` end in an error and not in a
// recursion.
func (rt *Runtime) phpxClassChain(cd *model.ClassDecl, decls map[string]*model.ClassDecl, stmts []model.Stmt, chain map[string]bool) (*model.Class, error) {
	if chain[cd.Name] {
		return nil, fmt.Errorf("class %s: circular inheritance", cd.Name)
	}
	chain[cd.Name] = true
	defer delete(chain, cd.Name)

	c := &model.Class{
		Name:        cd.Name,
		Implements:  model.InterfaceNames(cd, stmts),
		Fields:      cd.Fields,
		Statics:     cd.Statics,
		Consts:      cd.Consts,
		Methods:     make(map[string]*model.FuncDecl, len(cd.Methods)),
		MethodOwner: make(map[string]string, len(cd.Methods)),
	}
	for _, m := range cd.Methods {
		c.Methods[m.Name] = m
		c.MethodOwner[m.Name] = cd.Name
	}
	if cd.Parent == "" {
		return c, nil
	}

	parent, err := rt.phpxParentClass(cd.Name, cd.Parent, decls, stmts, chain)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		// A host-backed parent such as Exception confers no member: there is
		// no model.Class behind the Go constructor. The name stays in the
		// chain, so `instanceof Exception` and a catch clause answer php.
		c.Ancestors = []string{cd.Parent}
		return c, nil
	}

	c.Ancestors = append([]string{parent.Name}, parent.Ancestors...)
	c.Implements = phpxMergeNames(c.Implements, parent.Implements...)
	c.Fields = phpxMergeFields(parent.Fields, c.Fields)
	c.Statics = phpxMergeFields(parent.Statics, c.Statics)
	c.Consts = phpxMergeFields(parent.Consts, c.Consts)
	for name, method := range parent.Methods {
		if _, ok := phpxFieldFold(c.Methods, name); ok {
			continue
		}
		c.Methods[name] = method
		c.MethodOwner[name] = parent.MethodOwner[name]
	}
	return c, nil
}

// phpxParentClass resolves the name `extends` wrote. The declaration list of
// the program under hoisting comes first, then the class table, which holds a
// parent hoisted from an included file, then a host constructor, which is a
// class the runtime builds in Go and has no members to carry. A name none of
// them answers is the fatal php raises for an unknown parent.
func (rt *Runtime) phpxParentClass(child, name string, decls map[string]*model.ClassDecl, stmts []model.Stmt, chain map[string]bool) (*model.Class, error) {
	if cd, ok := decls[name]; ok {
		return rt.phpxClassChain(cd, decls, stmts, chain)
	}
	if c, ok := rt.lookupClass(name); ok {
		return c, nil
	}
	if _, ok := rt.lookupConstructor(name); ok {
		return nil, nil
	}
	return nil, fmt.Errorf("class %s: parent class %s not found", child, name)
}

// phpxMergeFields lays the parent's declarations first and each child
// declaration over the name it repeats, case-insensitively as php compares
// property and constant names. A child declaration the parent does not name is
// appended.
func phpxMergeFields(parent, child []model.Field) []model.Field {
	merged := make([]model.Field, len(parent), len(parent)+len(child))
	copy(merged, parent)
	for _, field := range child {
		replaced := false
		for i := range merged {
			if strings.EqualFold(merged[i].Name, field.Name) {
				merged[i] = field
				replaced = true
				break
			}
		}
		if !replaced {
			merged = append(merged, field)
		}
	}
	return merged
}

// phpxMergeNames appends the names own does not already hold, comparing
// case-insensitively because every name here is a class name.
func phpxMergeNames(own []string, inherited ...string) []string {
	for _, name := range inherited {
		held := false
		for _, have := range own {
			if strings.EqualFold(have, name) {
				held = true
				break
			}
		}
		if !held {
			own = append(own, name)
		}
	}
	return own
}

// phpxFieldFold returns the entry of a field or method table whose name
// matches case-insensitively.
func phpxFieldFold[T any](table map[string]T, name string) (T, bool) {
	if value, ok := table[name]; ok {
		return value, true
	}
	for key, value := range table {
		if strings.EqualFold(key, name) {
			return value, true
		}
	}
	var zero T
	return zero, false
}

// phpxOwnerClass names the class a running method body was declared in. php
// resolves `self::` there and not in the class of the instance, so a method
// inherited from a parent keeps the parent's name even when it runs on a child.
func (rt *Runtime) phpxOwnerClass(class *model.Class, decl *model.FuncDecl) string {
	if class == nil {
		return ""
	}
	if owner := class.MethodOwner[decl.Name]; owner != "" {
		return owner
	}
	return class.Name
}

// phpxNoParentError is the error `parent` raises when the frame records no
// parent: outside a class body, or in one that declares no `extends`.
func phpxNoParentError() error {
	return fmt.Errorf(`cannot use "parent" when current class scope has no parent`)
}

// phpxParentScope names the parent of the class a body was declared in. The
// frame records it as `__parent__` when a method starts, and `parent::`
// reads it from there. The class lookup walks the chain the hoist recorded.
func (rt *Runtime) phpxParentScope(owner string) string {
	if owner == "" {
		return ""
	}
	if class, ok := rt.lookupClass(owner); ok && len(class.Ancestors) > 0 {
		return class.Ancestors[0]
	}
	return ""
}

// phpxCarries reports whether a call to a method declared by target may run on
// an instance of class. It is the check `$this` forwarding makes through
// `self::`, `parent::` and `static::`: php forwards the receiver when the
// instance's class is the declaring class or derives from it.
func phpxCarries(class, target *model.Class) bool {
	if class == nil || target == nil {
		return false
	}
	if class == target || strings.EqualFold(class.Name, target.Name) {
		return true
	}
	for _, name := range class.Ancestors {
		if strings.EqualFold(name, target.Name) {
			return true
		}
	}
	return false
}

// phpxCatchExtends reports whether err is a thrown object whose class chain
// names want, which is what `catch (Parent)` takes for a thrown child.
func phpxCatchExtends(want string, err error) bool {
	thrown, ok := err.(*objectError)
	if !ok || thrown.object == nil || thrown.object.Class == nil {
		return false
	}
	class := thrown.object.Class
	for _, name := range class.Ancestors {
		if strings.EqualFold(want, name) {
			return true
		}
	}
	for _, name := range class.Implements {
		if strings.EqualFold(want, name) {
			return true
		}
	}
	return false
}
