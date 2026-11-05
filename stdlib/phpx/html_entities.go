//go:build phpx

package phpx

import (
	"html"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerHTMLEntities)
}

// The ENT_* flags with php's values.
const (
	entCompat     = 2
	entQuotes     = 3
	entNoQuotes   = 0
	entHTML5      = 48
	entHTML401    = 0
	entSubstitute = 8
)

// latin1Names are the HTML 4.01 entity names of U+00A0 to U+00FF, in order.
var latin1Names = strings.Fields(`nbsp iexcl cent pound curren yen brvbar sect uml copy ordf laquo not shy reg macr
deg plusmn sup2 sup3 acute micro para middot cedil sup1 ordm raquo frac14 frac12 frac34 iquest
Agrave Aacute Acirc Atilde Auml Aring AElig Ccedil Egrave Eacute Ecirc Euml Igrave Iacute Icirc Iuml
ETH Ntilde Ograve Oacute Ocirc Otilde Ouml times Oslash Ugrave Uacute Ucirc Uuml Yacute THORN szlig
agrave aacute acirc atilde auml aring aelig ccedil egrave eacute ecirc euml igrave iacute icirc iuml
eth ntilde ograve oacute ocirc otilde ouml divide oslash ugrave uacute ucirc uuml yacute thorn yuml`)

// otherNames are the common HTML 4.01 entities beyond Latin-1.
var otherNames = map[rune]string{
	0x20ac: "euro", 0x2026: "hellip", 0x2014: "mdash", 0x2013: "ndash", 0x2018: "lsquo", 0x2019: "rsquo",
	0x201c: "ldquo", 0x201d: "rdquo", 0x2022: "bull", 0x2122: "trade", 0x2190: "larr", 0x2192: "rarr",
	0x2191: "uarr", 0x2193: "darr", 0x2264: "le", 0x2265: "ge", 0x2260: "ne", 0x221e: "infin",
	0x03b1: "alpha", 0x03b2: "beta", 0x03b3: "gamma", 0x03b4: "delta", 0x03c0: "pi", 0x03a9: "Omega",
	0x2020: "dagger", 0x2021: "Dagger", 0x2030: "permil", 0x2039: "lsaquo", 0x203a: "rsaquo",
	0x0152: "OElig", 0x0153: "oelig", 0x0160: "Scaron", 0x0161: "scaron", 0x0178: "Yuml", 0x0192: "fnof",
	0x02c6: "circ", 0x02dc: "tilde", 0x2212: "minus", 0x2211: "sum", 0x221a: "radic", 0x2248: "asymp",
}

func entityName(r rune) string {
	if r >= 0xa0 && r <= 0xff {
		return latin1Names[r-0xa0]
	}
	return otherNames[r]
}

// htmlEscape is htmlspecialchars and htmlentities: all selects the named
// entities beyond the five.
func htmlEscape(s string, flags int64, doubleEncode, all bool) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size <= 1 {
			if flags&entSubstitute != 0 {
				b.WriteString("�")
			}
			i++
			continue
		}
		switch {
		case r == '&':
			if !doubleEncode {
				if end := entityEnd(s[i:]); end > 0 {
					b.WriteString(s[i : i+end])
					i += end
					continue
				}
			}
			b.WriteString("&amp;")
		case r == '<':
			b.WriteString("&lt;")
		case r == '>':
			b.WriteString("&gt;")
		case r == '"' && flags&entCompat != 0:
			b.WriteString("&quot;")
		case r == '\'' && flags&entQuotes == entQuotes:
			if flags&entHTML5 != 0 {
				b.WriteString("&apos;")
			} else {
				b.WriteString("&#039;")
			}
		case all && entityName(r) != "":
			b.WriteString("&" + entityName(r) + ";")
		default:
			b.WriteString(s[i : i+size])
		}
		i += size
	}
	return b.String()
}

// entityEnd answers the length of a well-formed entity at the start of s, or 0.
func entityEnd(s string) int {
	if len(s) < 3 || s[0] != '&' {
		return 0
	}
	semi := strings.IndexByte(s, ';')
	if semi < 2 || semi > 32 {
		return 0
	}
	body := s[1:semi]
	if body[0] == '#' {
		if len(body) < 2 {
			return 0
		}
		digits := body[1:]
		base := 10
		if digits[0] == 'x' || digits[0] == 'X' {
			digits, base = digits[1:], 16
		}
		if _, err := strconv.ParseUint(digits, base, 32); err != nil {
			return 0
		}
		return semi + 1
	}
	for i := 0; i < len(body); i++ {
		c := body[i]
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
			return 0
		}
	}
	return semi + 1
}

// htmlDecode is htmlspecialchars_decode and html_entity_decode.
func htmlDecode(s string, flags int64, all bool) string {
	if !strings.Contains(s, "&") {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] != '&' {
			b.WriteByte(s[i])
			i++
			continue
		}
		end := entityEnd(s[i:])
		if end == 0 {
			b.WriteByte('&')
			i++
			continue
		}
		ent := s[i : i+end]
		var out string
		switch ent {
		case "&amp;":
			out = "&"
		case "&lt;":
			out = "<"
		case "&gt;":
			out = ">"
		case "&quot;":
			if flags&entCompat != 0 {
				out = "\""
			}
		case "&#039;", "&#39;", "&#x27;", "&apos;":
			if flags&entQuotes == entQuotes && (ent != "&apos;" || flags&entHTML5 != 0) {
				out = "'"
			}
		default:
			if all {
				if dec := html.UnescapeString(ent); dec != ent {
					out = dec
				}
			}
		}
		if out == "" {
			b.WriteString(ent)
		} else {
			b.WriteString(out)
		}
		i += end
	}
	return b.String()
}

func registerHTMLEntities(rt *runner.Runtime) {
	flagsOf := func(rest []any) int64 {
		if len(rest) > 0 && rest[0] != nil {
			return phpval.Int(rest[0])
		}
		return entQuotes | entSubstitute
	}
	doubleOf := func(rest []any) bool { return len(rest) < 3 || rest[2] == nil || phpval.Truthy(rest[2]) }
	rt.RegisterFunc("htmlspecialchars", func(s string, rest ...any) string {
		return htmlEscape(s, flagsOf(rest), doubleOf(rest), false)
	})
	rt.RegisterFunc("htmlentities", func(s string, rest ...any) string {
		return htmlEscape(s, flagsOf(rest), doubleOf(rest), true)
	})
	rt.RegisterFunc("htmlspecialchars_decode", func(s string, rest ...any) string {
		return htmlDecode(s, flagsOf(rest), false)
	})
	rt.RegisterFunc("html_entity_decode", func(s string, rest ...any) string {
		return htmlDecode(s, flagsOf(rest), true)
	})
}
