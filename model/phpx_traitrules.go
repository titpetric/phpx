package model

// TraitRule is one line of the adaptation block that follows a `use` line:
// `A::f insteadof B, C;` names the trait whose method wins, and
// `A::f as protected g;` gives a method another name, another visibility, or
// both. Trait is empty when the line names the method alone.
type TraitRule struct {
	Trait      string
	Method     string
	Insteadof  []string // traits that lose Method; empty for an `as` rule
	Alias      string   // the new name of an `as` rule; may be empty
	Visibility string   // the visibility of an `as` rule; may be empty
}

// IsInsteadof reports whether the rule picks a winner and not an alias.
func (r TraitRule) IsInsteadof() bool { return len(r.Insteadof) > 0 }
