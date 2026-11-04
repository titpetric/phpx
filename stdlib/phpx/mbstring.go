//go:build phpx

package phpx

import (
	"strings"
	"unicode/utf8"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxMbstring)
}

// phpxEncodingName spells an encoding the way mbstring reports it.
func phpxEncodingName(name string) string {
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "UTF8", "UTF-8":
		return "UTF-8"
	case "ASCII", "US-ASCII":
		return "ASCII"
	case "ISO-8859-1", "LATIN1", "ISO8859-1":
		return "ISO-8859-1"
	}
	return strings.ToUpper(strings.TrimSpace(name))
}

// phpxEncodingList reads the encodings argument: a comma list or an array.
func phpxEncodingList(v any) []string {
	switch t := v.(type) {
	case string:
		var out []string
		for _, part := range strings.Split(t, ",") {
			if part = strings.TrimSpace(part); part != "" {
				out = append(out, phpxEncodingName(part))
			}
		}
		return out
	case *model.Array:
		var out []string
		t.Range(func(_, item any) bool {
			if s, ok := item.(string); ok {
				out = append(out, phpxEncodingName(s))
			}
			return true
		})
		return out
	}
	return nil
}

// phpxValidIn reports whether s is a well-formed string of the encoding.
func phpxValidIn(s, encoding string) bool {
	switch encoding {
	case "ASCII":
		for i := 0; i < len(s); i++ {
			if s[i] >= 0x80 {
				return false
			}
		}
		return true
	case "UTF-8":
		return utf8.ValidString(s)
	}
	return true
}

// phpxConvertEncoding converts s between the encodings it knows by name and
// leaves the rest as they came.
func phpxConvertEncoding(s, to, from string) string {
	if to == from || from == "" {
		if to == "ASCII" && !phpxValidIn(s, "ASCII") {
			return strings.Map(func(r rune) rune {
				if r >= 0x80 {
					return '?'
				}
				return r
			}, s)
		}
		return s
	}
	switch {
	case from == "ISO-8859-1" && (to == "UTF-8"):
		var b strings.Builder
		for i := 0; i < len(s); i++ {
			b.WriteRune(rune(s[i]))
		}
		return b.String()
	case (from == "UTF-8" || from == "ASCII") && to == "ISO-8859-1":
		var b strings.Builder
		for _, r := range s {
			if r < 0x100 {
				b.WriteByte(byte(r))
			} else {
				b.WriteByte('?')
			}
		}
		return b.String()
	}
	return s
}

// registerPhpxMbstring installs the encoding detection and conversion the
// console and string libraries ask about. Only UTF-8, ASCII and ISO-8859-1
// are known by name.
func registerPhpxMbstring(rt *runner.Runtime) {
	rt.RegisterFunc("mb_detect_encoding", func(s string, rest ...any) any {
		candidates := []string{"ASCII", "UTF-8"}
		if len(rest) > 0 && rest[0] != nil {
			if list := phpxEncodingList(rest[0]); len(list) > 0 {
				candidates = list
			}
		}
		strict := len(rest) > 1 && rest[1] == true
		for _, name := range candidates {
			if phpxValidIn(s, name) {
				return name
			}
		}
		if strict {
			return false
		}
		return candidates[len(candidates)-1]
	})
	var convert func(v any, to, from string) any
	convert = func(v any, to, from string) any {
		switch t := v.(type) {
		case string:
			return phpxConvertEncoding(t, to, from)
		case *model.Array:
			out := model.NewArray()
			t.Range(func(k, item any) bool {
				out.Set(k, convert(item, to, from))
				return true
			})
			return out
		}
		return v
	}
	rt.RegisterFunc("mb_convert_encoding", func(v any, toEncoding string, rest ...any) any {
		from := ""
		if len(rest) > 0 && rest[0] != nil {
			if list := phpxEncodingList(rest[0]); len(list) > 0 {
				from = list[0]
				if s, ok := v.(string); ok && len(list) > 1 {
					for _, name := range list {
						if phpxValidIn(s, name) {
							from = name
							break
						}
					}
				}
			}
		}
		return convert(v, phpxEncodingName(toEncoding), from)
	})
}
