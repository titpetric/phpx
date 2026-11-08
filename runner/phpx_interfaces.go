//go:build phpx

package runner

import (
	"strings"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/modelutil"
)

// phpxCheckInterfaceContracts holds a class declaring `implements` to the
// methods it actually has, which the hoist has folded together by the time
// this runs: its own, what a trait contributed, and what a parent carried,
// an autoloaded parent included. php answers `implements` from the finished
// class, so `class B extends A implements I` is satisfied by A's method and
// the check must not call it missing.
//
// The untagged build keeps upstream's AST-only check in the stub beside this.
func phpxCheckInterfaceContracts(prog *model.Program, classes map[string]*model.Class) error {
	violations := modelutil.PhpxCheckInterfaces(prog, func(cd *model.ClassDecl) map[string]bool {
		out := make(map[string]bool, 8)
		if c, ok := classes[cd.Name]; ok {
			for name := range c.Methods {
				out[strings.ToLower(name)] = true
			}
		}
		return out
	})
	if len(violations) == 0 {
		return nil
	}
	return &model.InterfaceContractError{Violations: violations}
}

// phpxMergeInterfaces folds an interface's inherited constants into its own
// declaration, so `interface B extends A` answers `B::X` for a constant A
// declares. The hoist registers these declarations and `Interface::CONST`
// reads them by name; php resolves an inherited interface constant through
// the extender's name. The merge is idempotent, so hoisting one program twice
// does not add a constant twice.
//
// An interface A declares in another included file answers through the
// runtime's registered declarations; frameworks put one interface per file,
// and a class reads a constant across that boundary.
func (rt *Runtime) phpxMergeInterfaces(stmts []model.Stmt) {
	index := rt.phpxInterfaceIndex(stmts)
	for _, id := range index {
		id.Consts = rt.phpxInterfaceConsts(id, index, map[string]bool{})
	}
}

// phpxInterfaceIndex maps the lower-cased name of every interface the file
// declares to its declaration.
func (rt *Runtime) phpxInterfaceIndex(stmts []model.Stmt) map[string]*model.InterfaceDecl {
	index := map[string]*model.InterfaceDecl{}
	for _, s := range stmts {
		if id, ok := s.(*model.InterfaceDecl); ok {
			index[strings.ToLower(id.Name)] = id
		}
	}
	return index
}

// phpxInterfaceConsts is an interface's own constants followed by the
// constants of every interface it extends, nearest declaration first. A name
// already present is not added again.
func (rt *Runtime) phpxInterfaceConsts(id *model.InterfaceDecl, index map[string]*model.InterfaceDecl, seen map[string]bool) []model.Field {
	key := strings.ToLower(id.Name)
	if seen[key] {
		return nil
	}
	seen[key] = true
	out := append([]model.Field(nil), id.Consts...)
	for _, parent := range id.Extends {
		p, ok := index[strings.ToLower(parent)]
		if !ok {
			p, ok = rt.phpxLookupInterface(parent)
		}
		if !ok {
			continue
		}
		for _, k := range rt.phpxInterfaceConsts(p, index, seen) {
			if phpxHasConst(out, k.Name) {
				continue
			}
			out = append(out, k)
		}
	}
	return out
}

// phpxAddInterfaceConsts appends the constants the class's interfaces
// contribute, in the order the class lists them. A name the class declares
// itself, or one a parent carried, is already in consts and wins: php reads
// the class's own constant and not the interface's.
func (rt *Runtime) phpxAddInterfaceConsts(consts []model.Field, interfaces []string, stmts []model.Stmt) []model.Field {
	if len(interfaces) == 0 {
		return consts
	}
	index := rt.phpxInterfaceIndex(stmts)
	out := append([]model.Field(nil), consts...)
	for _, name := range interfaces {
		id, ok := index[strings.ToLower(name)]
		if !ok {
			id, ok = rt.phpxLookupInterface(name)
		}
		if _, builtin := phpxBuiltinInterfaces[strings.ToLower(strings.TrimPrefix(name, "\\"))]; builtin && !ok {
			continue
		}
		if !ok {
			// An interface in a file nothing has included yet: the autoload
			// queue loads it, and its constants join the class.
			if loaded, _ := rt.InterfaceExists(strings.TrimPrefix(name, "\\"), true); loaded {
				id, ok = rt.phpxLookupInterface(name)
			}
		}
		if !ok {
			continue
		}
		for _, k := range rt.phpxInterfaceConstTree(id, nil) {
			if phpxHasConst(out, k.Name) {
				continue
			}
			out = append(out, k)
		}
	}
	return out
}

// phpxLookupInterface finds an interface the runtime registered from an
// included file. The name a declaration lists is compared without regard to
// case, as php compares an interface name.
func (rt *Runtime) phpxLookupInterface(name string) (*model.InterfaceDecl, bool) {
	if id, ok := rt.interfaces[name]; ok {
		return id, true
	}
	for key, id := range rt.interfaces {
		if strings.EqualFold(key, name) {
			return id, true
		}
	}
	return nil, false
}

// phpxHasConst reports whether consts already declares name, comparing
// case-insensitively as php compares a constant name.
func phpxHasConst(consts []model.Field, name string) bool {
	for _, k := range consts {
		if strings.EqualFold(k.Name, name) {
			return true
		}
	}
	return false
}

// phpxInterfaceConstTree lists the constants an interface holds: its own, then
// those of every interface it extends, transitively, loading the ones no file
// has declared yet. A name the interface repeats keeps its own value.
func (rt *Runtime) phpxInterfaceConstTree(id *model.InterfaceDecl, seen map[string]bool) []model.Field {
	if len(id.Extends) == 0 {
		return id.Consts
	}
	if seen == nil {
		seen = map[string]bool{}
	}
	seen[strings.ToLower(id.Name)] = true
	out := append([]model.Field(nil), id.Consts...)
	for _, ext := range id.Extends {
		ext = strings.TrimPrefix(ext, "\\")
		if seen[strings.ToLower(ext)] {
			continue
		}
		if _, builtin := phpxBuiltinInterfaces[strings.ToLower(ext)]; builtin {
			// The engine's own interfaces hold no constants and are no file to load.
			if _, declared := rt.phpxLookupInterface(ext); !declared {
				continue
			}
		}
		parent, ok := rt.phpxLookupInterface(ext)
		if !ok {
			if loaded, _ := rt.InterfaceExists(ext, true); !loaded {
				continue
			}
			if parent, ok = rt.phpxLookupInterface(ext); !ok {
				continue
			}
		}
		for _, k := range rt.phpxInterfaceConstTree(parent, seen) {
			if !phpxHasConst(out, k.Name) {
				out = append(out, k)
			}
		}
	}
	return out
}
