//go:build phpx

package runner

import (
	"fmt"
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxClass builds the runtime form of one class declaration, with the members
// `use` takes from a trait and `extends` carries from a parent flattened in.
// The child's own declaration of a field, static, constant or method wins over
// both, which is php's rule and the one place where the merge order matters.
//
// Flattening is what keeps model.Class free of a parent field: the guard in
// runner.TestNoInheritanceAtRuntime refuses one and a parent contributes its
// members to the child's own lists rather than a pointer to itself. The chain
// stays in Ancestors for `parent::` and `instanceof`, and each method remembers
// the class that declared it so `self::` inside an inherited body resolves as
// php resolves it. A trait contributes no ancestor: `instanceof TraitName` is
// false in php, and a trait is a bag of members and not a type.
func (rt *Runtime) phpxClass(cd *model.ClassDecl, decls map[string]*model.ClassDecl, traits map[string]*model.TraitDecl, stmts []model.Stmt) (*model.Class, error) {
	return rt.phpxClassChain(cd, decls, traits, stmts, map[string]bool{})
}

// phpxClassChain is phpxClass with the declarations already on the stack, so
// `class A extends B` and `class B extends A` end in an error and not in a
// recursion.
func (rt *Runtime) phpxClassChain(cd *model.ClassDecl, decls map[string]*model.ClassDecl, traits map[string]*model.TraitDecl, stmts []model.Stmt, chain map[string]bool) (*model.Class, error) {
	// decls holds the last declaration of a name, so a different one here
	// means the program declared the name twice. stdClass is a name php owns.
	if prior, ok := decls[cd.Name]; (ok && prior != cd) || cd.Name == "stdClass" {
		return nil, fmt.Errorf("Cannot redeclare class %s", cd.Name)
	}
	if chain[cd.Name] {
		return nil, fmt.Errorf("class %s: circular inheritance", cd.Name)
	}
	chain[cd.Name] = true
	defer delete(chain, cd.Name)
	cd = phpxPromote(cd)

	c := &model.Class{
		Name:        cd.Name,
		Implements:  model.InterfaceNames(cd, stmts),
		Fields:      cd.Fields,
		Statics:     cd.Statics,
		Consts:      cd.Consts,
		Methods:     make(map[string]*model.FuncDecl, len(cd.Methods)),
		MethodOwner: make(map[string]string, len(cd.Methods)),
	}
	phpxRecordImplements(c, cd)
	rt.phpxWidenImplements(c, cd)
	phpxRecordDecl(c, cd)
	for _, m := range cd.Methods {
		// php refuses a private declaration without a body before it counts
		// the class abstract; a trait may declare one, because the class that
		// uses it supplies the body.
		if m.Abstract && m.Visibility == "private" {
			return nil, fmt.Errorf("Abstract function %s() cannot be declared private", phpxMethodRef(cd.Name, m.Name))
		}
		phpxSetMethod(c, m, cd.Name)
	}
	// php raises a class's own abstract methods before traits bind and a
	// parent's members arrive: a concrete class that declares one hears about
	// that one and about no other.
	own := phpxOwnAbstract(cd)
	if !cd.Abstract && len(own) > 0 {
		return nil, phpxAbstractError(cd.Name, own)
	}
	var traitAbstract []string
	if len(cd.Traits) > 0 {
		applied, err := rt.phpxApplyTraits(c, cd, traits)
		if err != nil {
			return nil, err
		}
		traitAbstract = applied
	}
	var inherited []string
	var parentClass *model.Class
	ownFields := c.Fields
	if cd.Parent != "" {
		parent, err := rt.phpxParentClass(cd.Name, cd.Parent, decls, traits, stmts, chain)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			// A host-backed parent such as Exception confers no member: there
			// is no model.Class behind the Go constructor. The name stays in
			// the chain, so `instanceof Exception` and a catch clause answer
			// php.
			c.Ancestors = []string{cd.Parent}
		} else {
			parentClass = parent
			c.Ancestors = append([]string{parent.Name}, parent.Ancestors...)
			c.Implements = phpxMergeNames(c.Implements, parent.Implements...)
			c.Fields = phpxMergeFields(parent.Fields, c.Fields)
			phpxShareStatics(c, parent)
			c.Statics = phpxMergeFields(parent.Statics, c.Statics)
			c.Consts = phpxMergeFields(parent.Consts, c.Consts)
			inherited = parent.AbstractMethods
			for name, method := range parent.Methods {
				existing, ok := phpxFieldFold(c.Methods, name)
				if !ok {
					phpxSetMethod(c, method, parent.MethodOwner[name])
					continue
				}
				if !existing.Abstract {
					continue
				}
				if phpxDeclaresMethod(cd, name) {
					// A class may not turn a method its parent implements back
					// into a declaration without a body.
					if !method.Abstract {
						return nil, fmt.Errorf("Cannot make non abstract method %s() abstract in class %s",
							phpxMethodRef(parent.MethodOwner[name], method.Name), cd.Name)
					}
					continue
				}
				// A trait's abstract method is satisfied by the parent's
				// concrete one, and two abstract declarations keep the
				// parent's owner, which is the name php reports.
				phpxSetMethod(c, method, parent.MethodOwner[name])
			}
		}
	}
	// An interface contributes its constants to the class that names it, which
	// is what php answers for `Class::CONST`. They are appended, so a constant
	// the class declares, or one a parent carried, is the one that wins.
	c.Consts = rt.phpxAddInterfaceConsts(c.Consts, cd.Implements, stmts)
	rt.phpxBindStringer(c)
	phpxBindProps(c, ownFields, parentClass)
	return c, phpxResolveAbstract(c, cd, own, inherited, traitAbstract)
}

// phpxSetMethod writes a method into the class table under the name the
// declaration spells, dropping any entry the table holds for the same name in
// another case.
func phpxSetMethod(c *model.Class, method *model.FuncDecl, owner string) {
	for key := range c.Methods {
		if key != method.Name && strings.EqualFold(key, method.Name) {
			delete(c.Methods, key)
			delete(c.MethodOwner, key)
		}
	}
	c.Methods[method.Name] = method
	c.MethodOwner[method.Name] = owner
}

// phpxDeclaresMethod reports whether the class declaration writes the method
// itself, which is what keeps a trait's version and a parent's out of the
// composition.
func phpxDeclaresMethod(cd *model.ClassDecl, name string) bool {
	for _, m := range cd.Methods {
		if strings.EqualFold(m.Name, name) {
			return true
		}
	}
	return false
}

// phpxParentClass resolves the name `extends` wrote. The declaration list of
// the program under hoisting comes first, then the class table, which holds a
// parent hoisted from an included file, then a host constructor, which is a
// class the runtime builds in Go and has no members to carry. The autoload
// queue answers a name none of the three holds, which is where composer maps
// it. A name no source answers is the fatal php raises for an unknown parent.
func (rt *Runtime) phpxParentClass(child, name string, decls map[string]*model.ClassDecl, traits map[string]*model.TraitDecl, stmts []model.Stmt, chain map[string]bool) (*model.Class, error) {
	if cd, ok := decls[name]; ok {
		return rt.phpxClassChain(cd, decls, traits, stmts, chain)
	}
	if c, ok := rt.lookupClass(name); ok {
		return c, nil
	}
	// A built-in throwable has members a subclass reaches: getMessage() and
	// `parent::__construct($message, $code)`, which the Go constructor behind
	// `new Exception` does not carry (runner/phpx_throwable.go).
	if c, ok, err := rt.phpxThrowableParent(name, traits, stmts, chain); ok {
		return c, err
	}
	// An internal class answers from the runtime table and php never asks a
	// loader for one, so the host constructor comes before the queue.
	if _, ok := rt.lookupConstructor(name); ok {
		return nil, nil
	}
	// php resolves `extends Parent` by autoloading a name no declaration in
	// the file answers, and composer's PSR-4 loader is the way a framework
	// hands a parent to the class that extends it. `--include
	// vendor/autoload.php` registers it before this program is hoisted, so
	// the callback is in place when the declaration is built.
	if err := rt.autoload(name, rt.newScope()); err != nil {
		return nil, err
	}
	if c, ok := rt.lookupClass(name); ok {
		return c, nil
	}
	return nil, &phpxParentMissing{Child: child, Parent: name}
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
	// A body that is not the one the class table holds under its name is an
	// ancestor's, reached through parent::. The owner is then the class whose
	// table holds this very declaration, nearest first, and not the class that
	// overrode it: parent::f() from B::f() reaches A::f(), which names A.
	if own, ok := class.Methods[decl.Name]; ok && own != decl {
		for _, name := range class.Ancestors {
			if anc, found := rt.lookupClass(name); found {
				if m, has := anc.Methods[decl.Name]; has && m == decl {
					if owner := anc.MethodOwner[decl.Name]; owner != "" {
						return owner
					}
					return anc.Name
				}
			}
		}
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
		// A built-in throwable a Go constructor made: its class and the
		// ancestors php gives it (UnexpectedValueException is a RuntimeException).
		if class, isThrowable := throwableClassOf(err); isThrowable {
			return phpxBuiltinDescends(class, want)
		}
		return false
	}
	class := thrown.object.Class
	for _, name := range class.Ancestors {
		if phpxSameClassName(want, name) || phpxBuiltinDescends(name, want) {
			return true
		}
	}
	for _, name := range class.Implements {
		if phpxSameClassName(want, name) {
			return true
		}
	}
	return false
}

// phpxSameClassName compares a catch clause's name with a class name. The
// parser leaves a clause as it was written, so inside a namespace `catch (Bad
// $e)` names Ns\Bad; a clause with no namespace separator therefore also
// matches the last segment of the class.
func phpxSameClassName(want, class string) bool {
	if strings.EqualFold(want, class) {
		return true
	}
	if strings.Contains(want, "\\") {
		return false
	}
	if i := strings.LastIndexByte(class, '\\'); i >= 0 {
		return strings.EqualFold(want, class[i+1:])
	}
	return false
}

// phpxParentMissing is the fatal for a parent no source answers. hoist defers
// the class when it can still arrive: php declares such a class where the
// statement runs, after an include has registered the loader.
type phpxParentMissing struct{ Child, Parent string }

func (e *phpxParentMissing) Error() string {
	return fmt.Sprintf("class %s: parent class %s not found", e.Child, e.Parent)
}

// phpxBuiltinDescends reports whether the built-in throwable class has want
// among its ancestors in php's hierarchy.
func phpxBuiltinDescends(class, want string) bool {
	want = strings.TrimPrefix(want, "\\")
	for class != "" {
		next := ""
		for _, pair := range phpxThrowableParents {
			if strings.EqualFold(pair[0], class) {
				next = pair[1]
				break
			}
		}
		if next != "" && strings.EqualFold(next, want) {
			return true
		}
		class = next
	}
	return false
}
