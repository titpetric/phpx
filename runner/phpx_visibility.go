//go:build phpx

package runner

import (
	"fmt"

	"github.com/titpetric/phpx/model"
)

// phpxCheckVisible refuses a call php refuses: a private method from outside
// its declaring class, a protected one from outside its family.
func (rt *Runtime) phpxCheckVisible(caller *Scope, class *model.Class, decl *model.FuncDecl) error {
	if decl.Visibility != "private" && decl.Visibility != "protected" {
		return nil
	}
	owner := rt.phpxOwnerClass(class, decl)
	from := phpxCallerClass(caller)
	if from != "" && (rt.phpxMayCall(decl.Visibility, owner, from) || rt.phpxSharesRoot(decl, owner, from)) {
		return nil
	}
	scope := "global scope"
	if from != "" {
		scope = "scope " + from
	}
	if decl.Name == "__construct" {
		return fmt.Errorf("Call to %s %s::__construct() from %s", decl.Visibility, owner, scope)
	}
	return fmt.Errorf("Call to %s method %s::%s() from %s", decl.Visibility, owner, decl.Name, scope)
}

// phpxCallerClass names the class of the body running in caller, empty at the
// top level and in a plain function.
func phpxCallerClass(caller *Scope) string {
	if caller == nil {
		return ""
	}
	if v, ok := caller.Get("__class__"); ok {
		if name, ok := v.(string); ok {
			return name
		}
	}
	return ""
}

// phpxMayCall reports whether a body of class from reaches a member of owner.
func (rt *Runtime) phpxMayCall(visibility, owner, from string) bool {
	if from == owner {
		return true
	}
	if visibility == "private" {
		return false
	}
	related := func(a, b string) bool {
		c, ok := rt.lookupClass(a)
		if !ok {
			return false
		}
		for _, name := range c.Ancestors {
			if name == b {
				return true
			}
		}
		return false
	}
	return related(from, owner) || related(owner, from)
}

// phpxSharesRoot answers php's rule for a protected method: it is reachable
// from any class related to the class that first declared it, so a sibling
// subclass overriding it is reachable from the parent's other subclasses.
func (rt *Runtime) phpxSharesRoot(decl *model.FuncDecl, owner, from string) bool {
	if decl.Visibility != "protected" {
		return false
	}
	c, ok := rt.lookupClass(owner)
	if !ok {
		return false
	}
	root := owner
	for _, name := range c.Ancestors {
		if p, found := rt.lookupClass(name); found {
			if _, declares := p.Methods[decl.Name]; declares {
				root = name
			}
		}
	}
	return root == from || rt.phpxMayCall("protected", root, from)
}
