//go:build phpx

package phpx

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\MbstringExtra`, registerMbstringExtra)
}

// php's MB_CASE_* modes.
const (
	mbCaseUpper = iota
	mbCaseLower
	mbCaseTitle
	mbCaseFold
	mbCaseUpperSimple
	mbCaseLowerSimple
	mbCaseTitleSimple
	mbCaseFoldSimple
)

// mbConvertCase applies a case mode with the full Unicode mappings php uses
// (ß upper-cases to SS).
func mbConvertCase(s string, mode int64) string {
	switch int(mode) {
	case mbCaseUpper, mbCaseUpperSimple:
		return cases.Upper(language.Und).String(s)
	case mbCaseLower, mbCaseLowerSimple:
		return cases.Lower(language.Und).String(s)
	case mbCaseTitle, mbCaseTitleSimple:
		return cases.Title(language.Und).String(s)
	case mbCaseFold, mbCaseFoldSimple:
		return cases.Fold().String(s)
	}
	return s
}

func registerMbstringExtra(rt *runner.Runtime) {
	rt.RegisterFunc("mb_convert_case", func(s string, mode int64, _ ...any) string { return mbConvertCase(s, mode) })
	// php 8.4 takes the encoding as a second argument.
	rt.RegisterFunc("mb_ucfirst", func(s string, _ ...any) string { return mbFirst(s, unicode.ToUpper) })
	rt.RegisterFunc("mb_lcfirst", func(s string, _ ...any) string { return mbFirst(s, unicode.ToLower) })
	rt.RegisterFunc("mb_strtoupper", func(s string, _ ...any) string { return mbConvertCase(s, mbCaseUpper) })
	rt.RegisterFunc("mb_strtolower", func(s string, _ ...any) string { return mbConvertCase(s, mbCaseLower) })
	rt.RegisterFunc("mb_ord", func(s string, _ ...any) any {
		if s == "" {
			return false
		}
		r, size := utf8.DecodeRuneInString(s)
		if r == utf8.RuneError && size <= 1 {
			return false
		}
		return int64(r)
	})
	rt.RegisterFunc("mb_chr", func(cp int64, _ ...any) any {
		if cp < 0 || cp > unicode.MaxRune || (cp >= 0xD800 && cp <= 0xDFFF) {
			return false
		}
		return string(rune(cp))
	})
	rt.RegisterFunc("mb_scrub", func(s string, _ ...any) string { return strings.ToValidUTF8(s, "?") })
	rt.RegisterFunc("mb_internal_encoding", func(enc ...any) any {
		if len(enc) == 0 || enc[0] == nil {
			return "UTF-8"
		}
		return true
	})
	rt.RegisterFunc("mb_regex_encoding", func(enc ...any) any {
		if len(enc) == 0 || enc[0] == nil {
			return "UTF-8"
		}
		return true
	})
	rt.RegisterFunc("mb_language", func(lang ...any) any {
		if len(lang) == 0 || lang[0] == nil {
			return "neutral"
		}
		return true
	})
	rt.RegisterFunc("mb_substitute_character", func(c ...any) any {
		if len(c) == 0 || c[0] == nil {
			return int64('?')
		}
		return true
	})
	rt.RegisterFunc("mb_detect_order", func(order ...any) any {
		if len(order) == 0 || order[0] == nil {
			return stringList("ASCII", "UTF-8")
		}
		return true
	})
	rt.RegisterFunc("mb_http_input", func(_ ...any) any { return false })
	rt.RegisterFunc("mb_http_output", func(enc ...any) any {
		if len(enc) == 0 || enc[0] == nil {
			return "UTF-8"
		}
		return true
	})
	rt.RegisterFunc("mb_list_encodings", func() any { return stringList(mbEncodingNames()...) })
	for _, trim := range []struct {
		name        string
		left, right bool
	}{{"mb_trim", true, true}, {"mb_ltrim", true, false}, {"mb_rtrim", false, true}} {
		rt.RegisterFunc(trim.name, func(s string, rest ...any) string {
			return mbTrim(s, trim.left, trim.right, rest)
		})
	}
	rt.RegisterFunc("mb_strcut", func(s string, start int64, rest ...any) string { return mbStrcut(s, start, rest) })
	registerMbSearch(rt)
}

func stringList(items ...string) *model.Array {
	out := model.NewArraySize(len(items))
	for _, s := range items {
		out.Append(s)
	}
	return out
}

// mbTrim is mb_trim: it strips Unicode white space, or the characters of the
// second argument, from the chosen ends.
func mbTrim(s string, left, right bool, rest []any) string {
	pred := unicode.IsSpace
	if len(rest) > 0 && rest[0] != nil {
		set := phpval.String(rest[0])
		pred = func(r rune) bool { return strings.ContainsRune(set, r) }
	} else {
		pred = func(r rune) bool { return unicode.IsSpace(r) || r == 0 || r == 0x85 || r == 0xA0 || r == 0x180E }
	}
	if left {
		s = strings.TrimLeftFunc(s, pred)
	}
	if right {
		s = strings.TrimRightFunc(s, pred)
	}
	return s
}

// mbStrcut cuts at byte offsets that fall on character boundaries.
func mbStrcut(s string, start int64, rest []any) string {
	n := int64(len(s))
	if start < 0 {
		start = max(0, n+start)
	}
	if start > n {
		return ""
	}
	for start > 0 && !utf8.RuneStart(s[start]) {
		start--
	}
	end := n
	if len(rest) > 0 && rest[0] != nil {
		l := phpval.Int(rest[0])
		if l < 0 {
			end = max(start, n+l)
		} else {
			end = min(n, start+l)
		}
	}
	for end < n && end > start && !utf8.RuneStart(s[end]) {
		end--
	}
	return s[start:end]
}

// registerMbSearch installs the mb_str* searches that return a part of the
// haystack: mb_strstr, mb_stristr, mb_strrchr, mb_strrichr and mb_strripos.
func registerMbSearch(rt *runner.Runtime) {
	part := func(fold, last bool) func(h, n string, rest ...any) any {
		return func(h, n string, rest ...any) any {
			before := len(rest) > 0 && phpval.Truthy(rest[0])
			hay, needle := h, n
			if fold {
				hay, needle = mbConvertCase(h, mbCaseFold), mbConvertCase(n, mbCaseFold)
			}
			var i int
			if last {
				i = strings.LastIndex(hay, needle)
			} else {
				i = strings.Index(hay, needle)
			}
			if i < 0 {
				return false
			}
			if fold {
				// Fold may change byte lengths; locate the match by rune count.
				i = len(string([]rune(h)[:utf8.RuneCountInString(hay[:i])]))
			}
			if before {
				return h[:i]
			}
			return h[i:]
		}
	}
	rt.RegisterFunc("mb_strstr", part(false, false))
	rt.RegisterFunc("mb_stristr", part(true, false))
	rt.RegisterFunc("mb_strrchr", part(false, true))
	rt.RegisterFunc("mb_strrichr", part(true, true))
	rt.RegisterFunc("mb_strripos", func(h, n string, rest ...any) any {
		hay, needle := []rune(mbConvertCase(h, mbCaseFold)), []rune(mbConvertCase(n, mbCaseFold))
		offset := 0
		if len(rest) > 0 {
			offset = int(phpval.Int(rest[0]))
		}
		end := len(hay) - len(needle)
		start := 0
		if offset > 0 {
			start = offset
		} else if offset < 0 {
			end = min(end, len(hay)+offset)
		}
		for i := end; i >= start; i-- {
			if i >= 0 && string(hay[i:i+len(needle)]) == string(needle) {
				return int64(i)
			}
		}
		return false
	})
}

// mbFirst maps the first character of s.
func mbFirst(s string, mapping func(rune) rune) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError {
		return s
	}
	return string(mapping(r)) + s[size:]
}
