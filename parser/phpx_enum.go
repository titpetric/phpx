package parser

import (
	"strings"
)

// phpxRewriteEnums turns each `enum Name[: type] [implements A] { ... }` in the
// token stream into the class php's enum is made of, before the parser sees it:
//
//	final class Name implements A, UnitEnum[, BackedEnum] {
//	    public readonly string $name; public readonly mixed $value;
//	    public static function cases(): array { return [self::X, self::Y]; }
//	    public static function from($v) ... tryFrom($v) ...
//	    const X = __phpx_enum_case('Name', 'X', <backing value or null>);
//	    ...the body's own constants, methods and traits
//	}
//
// A case is a class constant whose value is a function that builds the object
// once and answers the same one afterwards, so `Name::X === Name::X`.
func phpxRewriteEnums(toks []token) []token {
	has := false
	for i := range toks {
		if toks[i].kind == tIdent && toks[i].val == "enum" && i+2 < len(toks) && toks[i+1].kind == tIdent &&
			!(i > 0 && toks[i-1].kind == tOp && (toks[i-1].val == "::" || toks[i-1].val == "->" || toks[i-1].val == "?->")) &&
			(toks[i+2].kind == tOp && (toks[i+2].val == ":" || toks[i+2].val == "{") || toks[i+2].kind == tIdent && toks[i+2].val == "implements") {
			has = true
			break
		}
	}
	if !has {
		return toks
	}
	out := make([]token, 0, len(toks)+64)
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		if !(t.kind == tIdent && t.val == "enum" && i+2 < len(toks) && toks[i+1].kind == tIdent &&
			!(i > 0 && toks[i-1].kind == tOp && (toks[i-1].val == "::" || toks[i-1].val == "->" || toks[i-1].val == "?->")) &&
			(toks[i+2].kind == tOp && (toks[i+2].val == ":" || toks[i+2].val == "{") || toks[i+2].kind == tIdent && toks[i+2].val == "implements")) {
			out = append(out, t)
			continue
		}
		name := toks[i+1].val
		j := i + 2
		backing := ""
		if toks[j].kind == tOp && toks[j].val == ":" {
			backing = toks[j+1].val
			j += 2
		}
		var implements []token
		if toks[j].kind == tIdent && toks[j].val == "implements" {
			j++
			for !(toks[j].kind == tOp && toks[j].val == "{") {
				implements = append(implements, toks[j])
				j++
			}
		}
		// toks[j] is the opening brace. Find the closing one and the cases.
		depth, end := 0, j
		var cases []string
		for k := j; k < len(toks); k++ {
			if toks[k].kind == tOp && toks[k].val == "{" {
				depth++
			} else if toks[k].kind == tOp && toks[k].val == "}" {
				depth--
				if depth == 0 {
					end = k
					break
				}
			} else if depth == 1 && toks[k].kind == tIdent && toks[k].val == "case" && k+1 < len(toks) && toks[k+1].kind == tIdent {
				cases = append(cases, toks[k+1].val)
			}
		}
		out = append(out, phpxLex(t.line, "final class "+name+" implements ")...)
		for _, impl := range implements {
			out = append(out, impl)
		}
		if len(implements) > 0 {
			out = append(out, phpxLex(t.line, ", ")...)
		}
		out = append(out, phpxLex(t.line, phpxEnumHeader(name, backing, cases))...)
		// The body, after the brace.
		depth = 1
		for k := j + 1; k <= end; k++ {
			tk := toks[k]
			if tk.kind == tOp && tk.val == "{" {
				depth++
			} else if tk.kind == tOp && tk.val == "}" {
				depth--
			}
			if depth == 1 && tk.kind == tIdent && tk.val == "case" && k+1 < len(toks) && toks[k+1].kind == tIdent {
				caseName := toks[k+1].val
				k += 2
				var value []token
				if toks[k].kind == tOp && toks[k].val == "=" {
					k++
					for !(toks[k].kind == tOp && toks[k].val == ";") {
						value = append(value, toks[k])
						k++
					}
				}
				out = append(out, phpxLex(tk.line, "const "+caseName+" = __phpx_enum_case('"+name+"', '"+caseName+"', ")...)
				if len(value) == 0 {
					out = append(out, phpxLex(tk.line, "null")...)
				} else {
					out = append(out, value...)
				}
				out = append(out, phpxLex(tk.line, ");")...)
				continue
			}
			out = append(out, tk)
		}
		i = end
	}
	return out
}

// phpxEnumHeader is the members every enum has, as source text.
func phpxEnumHeader(name, backing string, cases []string) string {
	var b strings.Builder
	b.WriteString("UnitEnum")
	if backing != "" {
		b.WriteString(", BackedEnum")
	}
	b.WriteString(" { public readonly string $name; public readonly mixed $value;")
	b.WriteString(" public static function cases(): array { return [")
	for i, c := range cases {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString("self::" + c)
	}
	b.WriteString("]; }")
	if backing != "" {
		b.WriteString(` public static function from($value): static { return __phpx_enum_from('` + name + `', $value, true); }`)
		b.WriteString(` public static function tryFrom($value): ?static { return __phpx_enum_from('` + name + `', $value, false); }`)
	}
	return b.String()
}

// phpxLex lexes a snippet of source into tokens on the given line.
func phpxLex(line int, src string) []token {
	toks, err := newLexer("<?php " + src).run()
	if err != nil {
		return nil
	}
	defer releaseTokens(toks)
	var out []token
	for _, t := range toks {
		if t.kind == tEOF {
			continue
		}
		t.line = line
		out = append(out, t)
	}
	return out
}
