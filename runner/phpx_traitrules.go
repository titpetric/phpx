//go:build phpx

package runner

import (
	"fmt"
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxAdaptTraits applies the adaptation block of a `use` line to the bodies
// of the traits it names, in the order php does: an alias is taken from the
// method as the trait declared it, then `insteadof` removes the losing
// methods. An `as` rule that names a visibility alone changes the method in
// place; one that names an alias adds a second method beside the original.
func phpxAdaptTraits(rules []model.TraitRule, names []string, bodies []*phpxTraitBody) error {
	if len(rules) == 0 {
		return nil
	}
	index := func(trait string) int {
		for i, name := range names {
			if strings.EqualFold(name, trait) {
				return i
			}
		}
		return -1
	}
	for _, r := range rules {
		if r.Trait != "" && index(r.Trait) < 0 {
			return fmt.Errorf("Required Trait %s wasn't added to the class", r.Trait)
		}
		for _, other := range r.Insteadof {
			if index(other) < 0 {
				return fmt.Errorf("Required Trait %s wasn't added to the class", other)
			}
		}
	}
	// Aliases first, so one is taken from the method as declared and not from a
	// visibility an earlier rule gave the original.
	for _, aliased := range []bool{true, false} {
		for _, r := range rules {
			if r.IsInsteadof() || (r.Alias != "") != aliased {
				continue
			}
			if err := phpxAliasMethod(r, bodies, index); err != nil {
				return err
			}
		}
	}
	for _, r := range rules {
		for _, other := range r.Insteadof {
			bodies[index(other)].dropMethod(r.Method)
		}
	}
	return nil
}

// phpxAliasMethod gives one `as` rule effect on the body that holds the method.
func phpxAliasMethod(r model.TraitRule, bodies []*phpxTraitBody, index func(string) int) error {
	var owner *phpxTraitBody
	if r.Trait != "" {
		owner = bodies[index(r.Trait)]
	} else {
		for _, b := range bodies {
			if _, ok := phpxFieldFold(b.methods, r.Method); ok {
				owner = b
				break
			}
		}
	}
	var method *model.FuncDecl
	if owner != nil {
		method, _ = phpxFieldFold(owner.methods, r.Method)
	}
	if method == nil {
		return fmt.Errorf("An alias was defined for method %s(), but this method does not exist", r.Method)
	}
	adapted := *method
	if r.Visibility != "" {
		adapted.Visibility = r.Visibility
	}
	if r.Alias == "" {
		owner.methods[method.Name] = &adapted
		return nil
	}
	adapted.Name = r.Alias
	owner.dropMethod(r.Alias)
	owner.methods[r.Alias] = &adapted
	owner.order = append(owner.order, r.Alias)
	owner.from[phpxTraitKey("method", r.Alias)] = owner.source("method", method.Name)
	return nil
}
