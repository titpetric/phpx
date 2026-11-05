//go:build phpx

package runner

import (
	"errors"
	"fmt"

	"github.com/titpetric/phpx/model"
)

// phpxPendingClass is a declaration hoist set aside, with what it needs to be
// built later.
type phpxPendingClass struct {
	decls    map[string]*model.ClassDecl
	traits   map[string]*model.TraitDecl
	stmts    []model.Stmt
	filename string
}

// phpxDeferClass sets cd aside when its parent is not there yet and reports
// whether it did. php builds a class whose parent it cannot see at compile
// time when the declaration statement runs, so a script that requires the
// autoloader first and declares `extends Vendor\Base` after it works.
func (rt *Runtime) phpxDeferClass(cd *model.ClassDecl, err error, decls map[string]*model.ClassDecl, traits map[string]*model.TraitDecl, stmts []model.Stmt, filename string) bool {
	var missing *phpxParentMissing
	var noTrait *phpxTraitMissing
	if !errors.As(err, &missing) && !errors.As(err, &noTrait) {
		return false
	}
	if rt.phpxPending == nil {
		rt.phpxPending = map[*model.ClassDecl]*phpxPendingClass{}
	}
	rt.phpxPending[cd] = &phpxPendingClass{decls: decls, traits: traits, stmts: stmts, filename: filename}
	return true
}

// phpxDeclareDeferred builds a class hoist set aside, when the statement that
// declares it runs. The parent is asked for again, now that the loader a
// require registered is in place.
func (rt *Runtime) phpxDeclareDeferred(cd *model.ClassDecl) error {
	pending, ok := rt.phpxPending[cd]
	if !ok {
		return nil
	}
	c, err := rt.phpxClass(cd, pending.decls, pending.traits, pending.stmts)
	if err != nil {
		return err
	}
	delete(rt.phpxPending, cd)
	for _, m := range cd.Methods {
		if m.Filename == "" {
			m.Filename = pending.filename
		}
	}
	rt.RegisterClass(c)
	return nil
}

// phpxTraitMissing is the fatal for a trait no source answers. A class using
// a trait that a later include declares is built where its statement runs.
type phpxTraitMissing struct{ Name string }

func (e *phpxTraitMissing) Error() string { return fmt.Sprintf("Trait %q not found", e.Name) }

// phpxDeclareNested declares a class whose statement sits inside a block, an
// `if (!class_exists('X')) { class X {} }` or a function body: hoist only reads
// the statements at the top of a file, and php declares such a class when the
// statement runs. A class hoist already declared is left alone.
func (rt *Runtime) phpxDeclareNested(cd *model.ClassDecl) error {
	if _, pending := rt.phpxPending[cd]; pending {
		return nil
	}
	if _, declared := rt.lookupClass(cd.Name); declared {
		return nil
	}
	traits := map[string]*model.TraitDecl{}
	c, err := rt.phpxClass(cd, map[string]*model.ClassDecl{cd.Name: cd}, traits, nil)
	if err != nil {
		return err
	}
	rt.RegisterClass(c)
	return nil
}
