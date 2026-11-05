//go:build phpx

package parser

import (
	"strings"

	"github.com/titpetric/phpx/model"
)

// phpxBuiltinTypes are the type names php reads as themselves and no class.
var phpxBuiltinTypes = map[string]bool{
	"int": true, "float": true, "string": true, "bool": true, "array": true,
	"callable": true, "iterable": true, "object": true, "mixed": true,
	"void": true, "null": true, "never": true, "false": true, "true": true,
	"self": true, "static": true, "parent": true,
}

// phpxResolveTypeHint rewrites the class names in a type spelling to the names
// the namespace and the `use` imports give them, as reflection reports them:
// `?Bar` in namespace Foo reads `?Foo\Bar`.
func (p *parser) phpxResolveTypeHint(hint string) string {
	if hint == "" {
		return hint
	}
	var b strings.Builder
	start := 0
	flush := func(end int) {
		atom := hint[start:end]
		switch {
		case atom == "":
		case strings.HasPrefix(atom, "\\"):
			b.WriteString(atom[1:])
		case phpxBuiltinTypes[strings.ToLower(atom)]:
			b.WriteString(atom)
		default:
			b.WriteString(p.qualify(atom, false))
		}
	}
	for i := 0; i < len(hint); i++ {
		switch hint[i] {
		case '?', '|', '&', '(', ')':
			flush(i)
			b.WriteByte(hint[i])
			start = i + 1
		}
	}
	flush(len(hint))
	return phpxNullableUnion(b.String())
}

// phpxNullableUnion spells `T|null` and `null|T` as php prints them: `?T`.
func phpxNullableUnion(hint string) string {
	if strings.ContainsAny(hint, "&(?") || strings.Count(hint, "|") != 1 {
		return hint
	}
	left, right, _ := strings.Cut(hint, "|")
	switch {
	case strings.EqualFold(right, "null") && !strings.EqualFold(left, "mixed"):
		return "?" + left
	case strings.EqualFold(left, "null") && !strings.EqualFold(right, "mixed"):
		return "?" + right
	}
	return hint
}

// phpxCatchName reads the class names of a catch clause through the file's
// imports: `use A\Ex\RuntimeException; catch (RuntimeException $e)` names that
// class and not the global one. A name no import covers is left as written,
// which keeps `catch (Exception $e)` in a namespace meaning the global class,
// as the runtime has always read it.
func (p *parser) phpxCatchName(hint string) string {
	if len(p.imports) == 0 || hint == "" {
		return hint
	}
	parts := strings.Split(hint, "|")
	for i, part := range parts {
		head, _, _ := strings.Cut(part, "\\")
		if _, ok := p.imports[head]; ok && !strings.HasPrefix(part, "\\") {
			parts[i] = p.qualify(part, false)
		}
	}
	return strings.Join(parts, "|")
}

// phpxQualifyConsts adds, for a `const` in a namespace, the entry under the
// qualified name a `\Ns\NAME` reference reads. The bare name stays, which is how
// the runtime's one constant table has always been read.
func (p *parser) phpxQualifyConsts(consts []model.Field) []model.Field {
	if p.namespace == "" {
		return consts
	}
	out := make([]model.Field, 0, len(consts)*2)
	for _, c := range consts {
		out = append(out, c)
		qualified := c
		qualified.Name = p.namespace + "\\" + c.Name
		out = append(out, qualified)
	}
	return out
}
