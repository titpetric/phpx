//go:build phpx

package model

import "strings"

// PhpxEffectiveMethodNames returns every method name a class has once the
// traits it uses and the ancestors it extends are folded in, lower-cased. It
// is the AST answer the linter uses; the fork's hoist answers from the built
// class, where a parent the program autoloads has already contributed.
//
// A name the class declares itself is reported once. A parent or trait no
// declaration in stmts defines contributes nothing: a vendor class has no AST
// here, and the caller that needs its members is the hoist and not the
// linter.
func PhpxEffectiveMethodNames(cd *ClassDecl, stmts []Stmt) map[string]bool {
	decls := map[string]*ClassDecl{}
	traits := map[string]*TraitDecl{}
	funcs := map[string][]string{}
	for _, s := range stmts {
		switch d := s.(type) {
		case *ClassDecl:
			decls[strings.ToLower(d.Name)] = d
		case *TraitDecl:
			traits[strings.ToLower(d.Name)] = d
		case *FuncDecl:
			// `function Class::method` is a method the class does not write
			// in its body, and the hoist attaches it to the same table.
			if d.Class != "" {
				key := strings.ToLower(d.Class)
				funcs[key] = append(funcs[key], d.Name)
			}
		}
	}
	out := map[string]bool{}
	phpxAddClassMethods(cd, decls, traits, funcs, map[string]bool{}, out)
	return out
}

// phpxAddClassMethods folds one class's methods into out: its own body first,
// then its traits, then its parent's, which is the order the hoist uses and
// the one place where the order does not change which names are present.
func phpxAddClassMethods(cd *ClassDecl, decls map[string]*ClassDecl, traits map[string]*TraitDecl, funcs map[string][]string, chain map[string]bool, out map[string]bool) {
	key := strings.ToLower(cd.Name)
	if chain[key] {
		return
	}
	chain[key] = true
	for _, m := range cd.Methods {
		out[strings.ToLower(m.Name)] = true
	}
	for _, name := range funcs[key] {
		out[strings.ToLower(name)] = true
	}
	traitSeen := map[string]bool{}
	for _, name := range cd.Traits {
		phpxAddTraitMethods(name, traits, traitSeen, out)
	}
	if cd.Parent == "" {
		return
	}
	if parent, ok := decls[strings.ToLower(cd.Parent)]; ok {
		phpxAddClassMethods(parent, decls, traits, funcs, chain, out)
	}
}

// phpxAddTraitMethods folds a trait's methods, and a trait it composes, into
// out. seen holds the traits already walked so a cycle ends.
func phpxAddTraitMethods(name string, traits map[string]*TraitDecl, seen map[string]bool, out map[string]bool) {
	key := strings.ToLower(name)
	if seen[key] {
		return
	}
	seen[key] = true
	t, ok := traits[key]
	if !ok {
		return
	}
	for _, m := range t.Methods {
		out[strings.ToLower(m.Name)] = true
	}
	for _, nested := range t.Traits {
		phpxAddTraitMethods(nested, traits, seen, out)
	}
}

// PhpxContractResolved reports whether every ancestor a class names is
// declared in stmts. A caller with an AST and no runtime uses it to tell a
// class that lacks a method from one whose parent is in a file the caller
// cannot see: the method may be the parent's, and the runtime autoloads and
// checks the finished class.
func PhpxContractResolved(cd *ClassDecl, stmts []Stmt) bool {
	decls := map[string]*ClassDecl{}
	for _, s := range stmts {
		if d, ok := s.(*ClassDecl); ok {
			decls[strings.ToLower(d.Name)] = d
		}
	}
	return phpxAncestorsResolved(cd, decls, map[string]bool{})
}

// phpxAncestorsResolved walks the `extends` chain and reports whether every
// name on it is declared. seen ends a cycle the parser accepted.
func phpxAncestorsResolved(cd *ClassDecl, decls map[string]*ClassDecl, seen map[string]bool) bool {
	key := strings.ToLower(cd.Name)
	if seen[key] {
		return true
	}
	seen[key] = true
	if cd.Parent == "" {
		return true
	}
	parent, ok := decls[strings.ToLower(cd.Parent)]
	if !ok {
		return false
	}
	return phpxAncestorsResolved(parent, decls, seen)
}

// PhpxCheckInterfaces reports what a class declaring `implements` lacks once
// the methods it actually has are counted. The fork's hoist passes a function
// over the class table, where traits and parents are already folded in; the
// linter passes PhpxEffectiveMethodNames over the AST.
//
// A name no interface declaration in prog defines is not a contract, exactly
// as CheckInterfaces treats an interface from another file.
func PhpxCheckInterfaces(prog *Program, declared func(*ClassDecl) map[string]bool) []InterfaceViolation {
	interfaces := interfaceIndex(prog.Stmts)
	if len(interfaces) == 0 {
		return nil
	}
	var out []InterfaceViolation
	check := func(cd *ClassDecl) {
		if len(cd.Implements) == 0 {
			return
		}
		have := declared(cd)
		for _, name := range cd.Implements {
			for _, req := range contractOf(name, interfaces) {
				if have[strings.ToLower(req.Method)] {
					continue
				}
				out = append(out, InterfaceViolation{
					Decl:      cd,
					Class:     cd.Name,
					Interface: req.Interface,
					Method:    req.Method,
				})
			}
		}
	}
	for _, s := range prog.Stmts {
		if cd, ok := s.(*ClassDecl); ok {
			check(cd)
		}
	}
	for _, cd := range prog.AnonClasses {
		check(cd)
	}
	return out
}
