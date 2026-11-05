//go:build phpx

package runner

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/titpetric/phpx/model"
)

// phpxApplyTraits folds the class's `use A, B;` list into a class that already
// holds its own members. A method the class declares itself wins and no trait
// contributes it, which is php's rule and the one that lets a class override a
// trait method with its own body. A property or constant it declares itself
// takes part in the composition as an ordinary definition: two declarations of
// one name compose when they agree and php calls the rest a collision.
func (rt *Runtime) phpxApplyTraits(c *model.Class, cd *model.ClassDecl, traits map[string]*model.TraitDecl) ([]string, error) {
	merged := newPhpxTraitBody()
	merged.seed(cd)
	bodies := make([]*phpxTraitBody, 0, len(cd.Traits))
	for _, name := range cd.Traits {
		body, err := rt.phpxTraitBody(name, traits, map[string]bool{})
		if err != nil {
			return nil, err
		}
		bodies = append(bodies, body)
	}
	if err := phpxAdaptTraits(cd.TraitRules, cd.Traits, bodies); err != nil {
		return nil, err
	}
	for _, body := range bodies {
		if err := merged.merge(body, cd.Name, phpxOwnNames(cd)); err != nil {
			return nil, err
		}
	}
	c.Fields = merged.fields
	c.Statics = merged.statics
	c.Consts = merged.consts
	var abstract []string
	for _, name := range merged.order {
		method := merged.methods[name]
		if _, ok := phpxFieldFold(c.Methods, name); ok {
			continue
		}
		if method.Abstract {
			// php attributes a trait's abstract method to the class that used
			// the trait, which is the name its message prints.
			abstract = append(abstract, phpxMethodRef(cd.Name, method.Name))
		}
		c.Methods[name] = method
		// php resolves `self::` inside a trait body to the class that used the
		// trait and not to the trait, which declares no storage of its own.
		c.MethodOwner[name] = cd.Name
	}
	return abstract, nil
}

// seed lays the composing class's own properties and constants down as the
// first definitions, so a trait repeating one of them is checked against it.
func (b *phpxTraitBody) seed(cd *model.ClassDecl) {
	for _, f := range cd.Fields {
		b.fields = append(b.fields, f)
		b.from[phpxTraitKey("field", f.Name)] = cd.Name
	}
	for _, f := range cd.Statics {
		b.statics = append(b.statics, f)
		b.from[phpxTraitKey("field", f.Name)] = cd.Name
	}
	for _, f := range cd.Consts {
		b.consts = append(b.consts, f)
		b.from[phpxTraitKey("const", f.Name)] = cd.Name
	}
}

// phpxTraitBody is a trait's members with its own nested `use` lines already
// folded in, and the trait that declared each name kept for the collision
// message. A method the trait declares itself replaces the one a nested trait
// contributed; a property or constant is checked against it, the same rule a
// class composing two traits follows.
type phpxTraitBody struct {
	fields  []model.Field
	statics []model.Field
	consts  []model.Field
	methods map[string]*model.FuncDecl
	order   []string          // method names in the order php composes them
	from    map[string]string // canonical member key -> trait that declared it
}

func newPhpxTraitBody() *phpxTraitBody {
	return &phpxTraitBody{methods: map[string]*model.FuncDecl{}, from: map[string]string{}}
}

// phpxTraitKey names a member for the collision bookkeeping. php compares
// method, property and constant names case-insensitively.
func phpxTraitKey(kind, name string) string {
	return kind + ":" + strings.ToLower(name)
}

// phpxOwnNames reports whether the class declares the method itself, which
// keeps the trait's version out of the composition. Only a method is exempt:
// php checks a property or constant the class repeats against the trait's.
func phpxOwnNames(cd *model.ClassDecl) func(kind, name string) bool {
	return func(kind, name string) bool {
		if kind != "method" {
			return false
		}
		for _, m := range cd.Methods {
			if strings.EqualFold(m.Name, name) {
				return true
			}
		}
		return false
	}
}

// phpxTraitBody resolves one trait and its nested `use` lines. A name the
// trait declares itself wins over the one a nested trait offers, and two
// nested traits declaring one name must agree or the error names both.
func (rt *Runtime) phpxTraitBody(name string, traits map[string]*model.TraitDecl, stack map[string]bool) (*phpxTraitBody, error) {
	key := strings.ToLower(name)
	if stack[key] {
		return nil, fmt.Errorf("trait %s: circular use", name)
	}
	td := phpxLookupTrait(traits, name)
	if td == nil {
		td = rt.phpxFindTrait(name)
	}
	if td == nil {
		return nil, &phpxTraitMissing{Name: name}
	}
	stack[key] = true
	defer delete(stack, key)

	body := newPhpxTraitBody()
	subs := make([]*phpxTraitBody, 0, len(td.Traits))
	for _, nested := range td.Traits {
		sub, err := rt.phpxTraitBody(nested, traits, stack)
		if err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	if err := phpxAdaptTraits(td.TraitRules, td.Traits, subs); err != nil {
		return nil, err
	}
	for _, sub := range subs {
		if err := body.merge(sub, td.Name, nil); err != nil {
			return nil, err
		}
	}
	if err := body.overlay(td); err != nil {
		return nil, err
	}
	return body, nil
}

// phpxLookupTrait finds a trait declaration by name, case-insensitively as php
// resolves every class and trait name.
func phpxLookupTrait(traits map[string]*model.TraitDecl, name string) *model.TraitDecl {
	if td, ok := traits[name]; ok {
		return td
	}
	for key, td := range traits {
		if strings.EqualFold(key, name) {
			return td
		}
	}
	return nil
}

// overlay lays the trait's own declarations over the ones its nested traits
// contributed. A method of the trait's own wins, as in php; a property or
// constant is another definition of the name and the two must agree or the
// composition fails.
func (b *phpxTraitBody) overlay(td *model.TraitDecl) error {
	for _, f := range td.Fields {
		if err := b.overlayField(td, f); err != nil {
			return err
		}
	}
	for _, f := range td.Statics {
		if err := b.overlayField(td, f); err != nil {
			return err
		}
	}
	for _, f := range td.Consts {
		existing, existingFrom, found := b.findConst(f.Name)
		if found {
			if !phpxSameField(existing, f) {
				return fmt.Errorf("%s and %s define the same constant (%s) in the composition of %s. However, the definition differs and is considered incompatible. Class was composed",
					td.Name, existingFrom, f.Name, td.Name)
			}
			continue
		}
		b.consts = append(b.consts, f)
		b.from[phpxTraitKey("const", f.Name)] = td.Name
	}
	for _, m := range td.Methods {
		b.dropMethod(m.Name)
		b.methods[m.Name] = m
		b.order = append(b.order, m.Name)
		b.from[phpxTraitKey("method", m.Name)] = td.Name
	}
	return nil
}

// overlayField lays one property of a trait's own body over the name a nested
// trait contributed, and reports php's collision when the definitions differ.
func (b *phpxTraitBody) overlayField(td *model.TraitDecl, f model.Field) error {
	existing, existingFrom, found := b.findField(f.Name)
	if found {
		if !phpxSameField(existing, f) {
			return fmt.Errorf("%s and %s define the same property ($%s) in the composition of %s. However, the definition differs and is considered incompatible. Class was composed",
				td.Name, existingFrom, f.Name, td.Name)
		}
		return nil
	}
	if isStaticField(td.Statics, f.Name) {
		b.statics = append(b.statics, f)
	} else {
		b.fields = append(b.fields, f)
	}
	b.from[phpxTraitKey("field", f.Name)] = td.Name
	return nil
}

// isStaticField reports whether name is declared in the static list.
func isStaticField(statics []model.Field, name string) bool {
	for _, f := range statics {
		if strings.EqualFold(f.Name, name) {
			return true
		}
	}
	return false
}

func (b *phpxTraitBody) dropMethod(name string) {
	for key := range b.methods {
		if strings.EqualFold(key, name) {
			delete(b.methods, key)
		}
	}
	kept := b.order[:0]
	for _, have := range b.order {
		if !strings.EqualFold(have, name) {
			kept = append(kept, have)
		}
	}
	b.order = kept
	delete(b.from, phpxTraitKey("method", name))
}

// merge composes another trait's body into b and reports php's collision when
// the two disagree. A name skip reports as the composing declaration's own is
// not a collision at all.
func (b *phpxTraitBody) merge(other *phpxTraitBody, compose string, skip func(kind, name string) bool) error {
	skipName := func(kind, name string) bool {
		return skip != nil && skip(kind, name)
	}
	for _, f := range other.fields {
		if skipName("field", f.Name) {
			continue
		}
		if err := b.addField("field", f, other.source("field", f.Name), compose); err != nil {
			return err
		}
	}
	for _, f := range other.statics {
		if skipName("field", f.Name) {
			continue
		}
		if err := b.addField("static", f, other.source("field", f.Name), compose); err != nil {
			return err
		}
	}
	for _, f := range other.consts {
		if skipName("const", f.Name) {
			continue
		}
		if err := b.addConst(f, other.source("const", f.Name), compose); err != nil {
			return err
		}
	}
	return b.addMethods(other, compose, skipName)
}

// source names the trait a member came from, which the collision message and
// the next merge both read.
func (b *phpxTraitBody) source(kind, name string) string {
	return b.from[phpxTraitKey(kind, name)]
}

// addField appends a property the body does not hold. A name it already holds
// composes when the definitions agree and is php's fatal when they differ.
func (b *phpxTraitBody) addField(kind string, f model.Field, from, compose string) error {
	existing, existingFrom, found := b.findField(f.Name)
	if !found {
		if kind == "static" {
			b.statics = append(b.statics, f)
		} else {
			b.fields = append(b.fields, f)
		}
		b.from[phpxTraitKey("field", f.Name)] = from
		return nil
	}
	if phpxSameField(existing, f) {
		return nil
	}
	return fmt.Errorf("%s and %s define the same property ($%s) in the composition of %s. However, the definition differs and is considered incompatible. Class was composed",
		existingFrom, from, f.Name, compose)
}

func (b *phpxTraitBody) addConst(f model.Field, from, compose string) error {
	existing, existingFrom, found := b.findConst(f.Name)
	if !found {
		b.consts = append(b.consts, f)
		b.from[phpxTraitKey("const", f.Name)] = from
		return nil
	}
	if phpxSameField(existing, f) {
		return nil
	}
	return fmt.Errorf("%s and %s define the same constant (%s) in the composition of %s. However, the definition differs and is considered incompatible. Class was composed",
		existingFrom, from, f.Name, compose)
}

// addMethods composes the method table. A method name two declarations share
// is php's collision whatever the bodies say, and the message names the later
// trait first.
func (b *phpxTraitBody) addMethods(other *phpxTraitBody, compose string, skip func(kind, name string) bool) error {
	for _, name := range other.order {
		if skip("method", name) {
			continue
		}
		existing, found := phpxFieldFold(b.methods, name)
		if !found {
			b.methods[name] = other.methods[name]
			b.order = append(b.order, name)
			b.from[phpxTraitKey("method", name)] = other.source("method", name)
			continue
		}
		// An abstract declaration asks for a method and a concrete one answers it:
		// the pair composes, with the concrete body.
		if incoming := other.methods[name]; incoming.Abstract {
			continue
		} else if existing.Abstract {
			b.methods[name] = incoming
			b.from[phpxTraitKey("method", name)] = other.source("method", name)
			continue
		}
		return fmt.Errorf("Trait method %s::%s has not been applied as %s::%s, because of collision with %s::%s",
			other.source("method", name), other.methods[name].Name, compose, other.methods[name].Name,
			b.source("method", name), existing.Name)
	}
	return nil
}

func (b *phpxTraitBody) findField(name string) (model.Field, string, bool) {
	for _, f := range b.fields {
		if strings.EqualFold(f.Name, name) {
			return f, b.from[phpxTraitKey("field", f.Name)], true
		}
	}
	for _, f := range b.statics {
		if strings.EqualFold(f.Name, name) {
			return f, b.from[phpxTraitKey("field", f.Name)], true
		}
	}
	return model.Field{}, "", false
}

func (b *phpxTraitBody) findConst(name string) (model.Field, string, bool) {
	for _, f := range b.consts {
		if strings.EqualFold(f.Name, name) {
			return f, b.from[phpxTraitKey("const", f.Name)], true
		}
	}
	return model.Field{}, "", false
}

// phpxSameField reports whether two property or constant declarations are the
// same definition. php composes two traits holding one name only when they
// agree on visibility, type and initial value, and reports the rest as a
// collision; the declared line is not part of the definition.
func phpxSameField(a, b model.Field) bool {
	a.Span, b.Span = model.SourceSpan{}, model.SourceSpan{}
	visibility := func(f model.Field) string {
		if f.Visibility == "" {
			return "public"
		}
		return f.Visibility
	}
	return strings.EqualFold(a.Name, b.Name) && visibility(a) == visibility(b) &&
		a.Type == b.Type && reflect.DeepEqual(a.Default, b.Default)
}

var (
	phpxTraitMu  sync.Mutex
	phpxTraitAll = map[string]*model.TraitDecl{}
)

// phpxRememberTrait keeps every trait a hoisted file declares, so a class in
// another file finds it.
func phpxRememberTrait(td *model.TraitDecl) {
	phpxTraitMu.Lock()
	phpxTraitAll[strings.ToLower(td.Name)] = td
	phpxTraitMu.Unlock()
}

// phpxFindTrait resolves a trait declared in a file other than the class's,
// asking the autoload queue for it when no hoisted file has.
func (rt *Runtime) phpxFindTrait(name string) *model.TraitDecl {
	key := strings.ToLower(strings.TrimPrefix(name, "\\"))
	lookup := func() *model.TraitDecl {
		phpxTraitMu.Lock()
		defer phpxTraitMu.Unlock()
		return phpxTraitAll[key]
	}
	if td := lookup(); td != nil {
		return td
	}
	if err := rt.autoload(strings.TrimPrefix(name, "\\"), rt.newScope()); err != nil {
		return nil
	}
	return lookup()
}

// PhpxTraitExists is trait_exists(): true when name is a declared trait, loaded
// through the autoloaders when autoload is set.
func (rt *Runtime) PhpxTraitExists(name string, autoload bool) bool {
	key := strings.ToLower(strings.TrimPrefix(name, "\\"))
	phpxTraitMu.Lock()
	_, known := phpxTraitAll[key]
	phpxTraitMu.Unlock()
	if known {
		return true
	}
	if !autoload {
		return false
	}
	return rt.phpxFindTrait(name) != nil
}

// PhpxClassUses is class_uses(): the traits a class names in its own body, or
// false for a class that does not exist.
func (rt *Runtime) PhpxClassUses(v any, autoload bool) any {
	var name string
	switch t := v.(type) {
	case *model.Object:
		if t.Class == nil {
			return false
		}
		name = t.Class.Name
	case string:
		name = strings.TrimPrefix(t, "\\")
	default:
		return false
	}
	if ok, _ := rt.ClassExists(name, autoload); !ok {
		return false
	}
	c, found := rt.lookupClass(name)
	if !found {
		return false
	}
	out := model.NewArray()
	if cd := phpxDeclOf(c); cd != nil {
		for _, tn := range cd.Traits {
			tn = strings.TrimPrefix(tn, "\\")
			out.Set(tn, tn)
		}
	}
	return out
}
