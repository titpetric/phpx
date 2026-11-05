//go:build phpx

package compat

import (
	"errors"
	"strconv"
	"strings"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

// phpxExpandReplacement writes the replacement for one match the way php reads
// it: \n, $n and ${n} name a group (one or two digits), a doubled backslash is
// one backslash, and every other $ or \ is itself.
func phpxExpandReplacement(b *strings.Builder, repl, subject string, idx []int) {
	group := func(n int) {
		if 2*n+1 < len(idx) && idx[2*n] >= 0 {
			b.WriteString(subject[idx[2*n]:idx[2*n+1]])
		}
	}
	digits := func(i int) (int, int, bool) {
		j := i
		for j < len(repl) && j < i+2 && repl[j] >= '0' && repl[j] <= '9' {
			j++
		}
		if j == i {
			return 0, i, false
		}
		n, _ := strconv.Atoi(repl[i:j])
		return n, j, true
	}
	for i := 0; i < len(repl); i++ {
		switch c := repl[i]; {
		case c == '\\' && i+1 < len(repl) && repl[i+1] == '\\':
			b.WriteByte('\\')
			i++
		case c == '\\' || c == '$':
			if n, end, ok := digits(i + 1); ok {
				group(n)
				i = end - 1
			} else if c == '$' && i+1 < len(repl) && repl[i+1] == '{' {
				if n, end, ok := digits(i + 2); ok && end < len(repl) && repl[end] == '}' {
					group(n)
					i = end
				} else {
					b.WriteByte(c)
				}
			} else {
				b.WriteByte(c)
			}
		default:
			b.WriteByte(c)
		}
	}
}

// phpxReplaceOne applies one pattern to one subject with at most limit
// replacements (-1 for all) and answers the new subject and the count.
func (c *regexpCache) phpxReplaceOne(pattern, repl, subject string, limit int64) (string, int64, bool) {
	re, err := c.compilePCRE(pattern)
	if err != nil {
		return "", 0, false
	}
	var b strings.Builder
	last, n := 0, int64(0)
	for _, idx := range re.findAllIndex(subject, 0) {
		if limit >= 0 && n >= limit {
			break
		}
		b.WriteString(subject[last:idx[0]])
		phpxExpandReplacement(&b, repl, subject, idx)
		last = idx[1]
		n++
	}
	b.WriteString(subject[last:])
	return b.String(), n, true
}

func phpxStrings(v any) []string {
	var out []string
	if model.IsCollection(v) {
		model.RangeValues(v, func(_, item any) bool {
			out = append(out, phpval.String(item))
			return true
		})
		return out
	}
	return []string{phpval.String(v)}
}

// phpxReplaceSubject runs every pattern over one subject. The replacement is a
// string for all of them, or a list matched to the patterns by position.
func (c *regexpCache) phpxReplaceSubject(patterns []string, replacements []string, listRepl bool, subject string, limit int64) (string, int64, bool) {
	total := int64(0)
	for i, p := range patterns {
		repl := ""
		switch {
		case !listRepl:
			repl = replacements[0]
		case i < len(replacements):
			repl = replacements[i]
		}
		next, n, ok := c.phpxReplaceOne(p, repl, subject, limit)
		if !ok {
			return "", 0, false
		}
		subject = next
		total += n
	}
	return subject, total, true
}

func phpxSetCount(rest []any, i int, total int64) {
	if i < len(rest) {
		if set, ok := rest[i].(func(any)); ok {
			set(total)
		}
	}
}

func phpxLimit(rest []any) int64 {
	if len(rest) > 0 && rest[0] != nil {
		return phpval.Int(rest[0])
	}
	return -1
}

// phpxPregReplace is preg_replace and preg_filter over strings and lists.
func (c *regexpCache) phpxPregReplace(filter bool, pattern, replacement, subject any, rest []any) (any, error) {
	patterns := phpxStrings(pattern)
	listRepl := model.IsCollection(replacement)
	if listRepl && !model.IsCollection(pattern) {
		return nil, errors.New("preg_replace(): Argument #1 ($pattern) must be of type array when argument #2 ($replacement) is an array, string given")
	}
	replacements := phpxStrings(replacement)
	limit := phpxLimit(rest)
	total := int64(0)
	if !model.IsCollection(subject) {
		out, n, ok := c.phpxReplaceSubject(patterns, replacements, listRepl, phpval.String(subject), limit)
		phpxSetCount(rest, 1, n)
		if !ok || filter && n == 0 {
			return nil, nil
		}
		return out, nil
	}
	result := model.NewArray()
	model.RangeValues(subject, func(k, item any) bool {
		out, n, ok := c.phpxReplaceSubject(patterns, replacements, listRepl, phpval.String(item), limit)
		total += n
		if ok && (!filter || n > 0) {
			result.Set(k, out)
		}
		return true
	})
	phpxSetCount(rest, 1, total)
	return result, nil
}

// phpxPregReplaceCallback runs the callback over every pattern and subject.
func (c *regexpCache) phpxPregReplaceCallback(rt *runner.Runtime, pattern, callback, subject any, rest []any) (any, error) {
	fn, ok := rt.Callable(callback)
	if !ok {
		return nil, errors.New("preg_replace_callback(): argument #2 ($callback) must be a valid callback")
	}
	var limit any
	if len(rest) > 0 {
		limit = rest[0]
	}
	var flags any
	if len(rest) > 2 {
		flags = rest[2]
	}
	total := int64(0)
	one := func(s string) (any, error) {
		cur := any(s)
		for _, p := range phpxStrings(pattern) {
			var err error
			cur, err = c.phpPregReplaceCallback(p, fn, phpval.String(cur), limit, func(v any) { total += phpval.Int(v) }, flags)
			if err != nil || cur == nil {
				return nil, err
			}
		}
		return cur, nil
	}
	var out any
	if !model.IsCollection(subject) {
		v, err := one(phpval.String(subject))
		if err != nil {
			return nil, err
		}
		out = v
	} else {
		result := model.NewArray()
		var failure error
		model.RangeValues(subject, func(k, item any) bool {
			v, err := one(phpval.String(item))
			if err != nil {
				failure = err
				return false
			}
			result.Set(k, v)
			return true
		})
		if failure != nil {
			return nil, failure
		}
		out = result
	}
	phpxSetCount(rest, 1, total)
	return out, nil
}

// phpxRegisterReplace installs the replace family over the shared pattern cache.
func (c *regexpCache) phpxRegisterReplace(rt *runner.Runtime) {
	rt.RegisterFunc("preg_replace", func(pattern, replacement, subject any, rest ...any) (any, error) {
		return c.phpxPregReplace(false, pattern, replacement, subject, rest)
	})
	rt.RegisterFunc("preg_filter", func(pattern, replacement, subject any, rest ...any) (any, error) {
		return c.phpxPregReplace(true, pattern, replacement, subject, rest)
	})
	rt.RegisterFunc("preg_replace_callback", func(pattern, callback, subject any, rest ...any) (any, error) {
		return c.phpxPregReplaceCallback(rt, pattern, callback, subject, rest)
	})
	rt.RegisterFunc("preg_replace_callback_array", func(patterns, subject any, rest ...any) (any, error) {
		total := int64(0)
		cur := subject
		var failure error
		model.RangeValues(patterns, func(p, cb any) bool {
			counted := []any{nil, func(v any) { total += phpval.Int(v) }}
			if len(rest) > 0 {
				counted[0] = rest[0]
			}
			if len(rest) > 2 {
				counted = append(counted, rest[2])
			}
			cur, failure = c.phpxPregReplaceCallback(rt, p, cb, cur, counted)
			return failure == nil && cur != nil
		})
		phpxSetCount(rest, 1, total)
		return cur, failure
	})
}
