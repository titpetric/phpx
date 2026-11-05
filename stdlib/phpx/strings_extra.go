//go:build phpx

package phpx

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\StringsExtra`, registerStringsExtra)
}

// tokState is the position strtok() keeps between calls.
type tokState struct {
	s   string
	pos int
}

func (t *tokState) next(delims string) any {
	for t.pos < len(t.s) && strings.IndexByte(delims, t.s[t.pos]) >= 0 {
		t.pos++
	}
	if t.pos >= len(t.s) {
		return false
	}
	start := t.pos
	for t.pos < len(t.s) && strings.IndexByte(delims, t.s[t.pos]) < 0 {
		t.pos++
	}
	tok := t.s[start:t.pos]
	if t.pos < len(t.s) {
		t.pos++
	}
	return tok
}

// similarText is php_similar_char: the longest common substring, then the same
// on what lies either side of it.
func similarText(a, b string) int {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	max, pa, pb := 0, 0, 0
	for i := 0; i < len(a); i++ {
		for j := 0; j < len(b); j++ {
			k := 0
			for i+k < len(a) && j+k < len(b) && a[i+k] == b[j+k] {
				k++
			}
			if k > max {
				max, pa, pb = k, i, j
			}
		}
	}
	if max == 0 {
		return 0
	}
	return max + similarText(a[:pa], b[:pb]) + similarText(a[pa+max:], b[pb+max:])
}

// strIncrement is str_increment: the perl-style successor of an alphanumeric
// string.
func strIncrement(s string) (string, error) {
	if s == "" {
		return "", &valueError{message: "str_increment(): Argument #1 ($string) cannot be empty"}
	}
	b := []byte(s)
	for _, c := range b {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
			return "", &valueError{message: "str_increment(): Argument #1 ($string) must be composed only of alphanumeric ASCII characters"}
		}
	}
	for i := len(b) - 1; i >= 0; i-- {
		switch {
		case b[i] == 'z':
			b[i] = 'a'
		case b[i] == 'Z':
			b[i] = 'A'
		case b[i] == '9':
			b[i] = '0'
		default:
			b[i]++
			return string(b), nil
		}
	}
	prefix := map[byte]byte{'a': 'a', 'A': 'A', '0': '1'}[b[0]]
	return string(prefix) + string(b), nil
}

// strDecrement is str_decrement, the inverse of strIncrement.
func strDecrement(s string) (string, error) {
	if s == "" {
		return "", &valueError{message: "str_decrement(): Argument #1 ($string) cannot be empty"}
	}
	b := []byte(s)
	for _, c := range b {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') {
			return "", &valueError{message: "str_decrement(): Argument #1 ($string) must be composed only of alphanumeric ASCII characters"}
		}
	}
	for i := len(b) - 1; i >= 0; i-- {
		switch b[i] {
		case 'a':
			b[i] = 'z'
		case 'A':
			b[i] = 'Z'
		case '0':
			b[i] = '9'
		default:
			b[i]--
			if i == 0 && len(b) > 1 && b[0] == '0' {
				return string(b[1:]), nil
			}
			return string(b), nil
		}
	}
	if len(b) > 1 {
		return string(b[1:]), nil
	}
	return "", &valueError{message: "str_decrement(): Argument #1 ($string) \"" + s + "\" is out of decrement range"}
}

func registerStringsExtra(rt *runner.Runtime) {
	tok := &tokState{}
	rt.RegisterFunc("strtok", func(a string, rest ...any) any {
		if len(rest) > 0 {
			tok.s, tok.pos = a, 0
			return tok.next(phpval.String(rest[0]))
		}
		return tok.next(a)
	})
	rt.RegisterFunc("similar_text", func(a, b string, rest ...any) int64 {
		n := similarText(a, b)
		if len(rest) > 0 {
			if set, ok := rest[0].(func(any)); ok {
				pct := 0.0
				if len(a)+len(b) > 0 {
					pct = float64(n) * 2 * 100 / float64(len(a)+len(b))
				}
				set(pct)
			}
		}
		return int64(n)
	})
	rt.RegisterFunc("str_increment", strIncrement)
	rt.RegisterFunc("str_decrement", strDecrement)
	rt.RegisterFunc("str_rot13", func(s string) string {
		return strings.Map(func(r rune) rune {
			switch {
			case r >= 'a' && r <= 'z':
				return 'a' + (r-'a'+13)%26
			case r >= 'A' && r <= 'Z':
				return 'A' + (r-'A'+13)%26
			}
			return r
		}, s)
	})
	rt.RegisterFunc("stristr", func(h, n string, rest ...any) any {
		i := strings.Index(strings.ToLower(h), strings.ToLower(n))
		if i < 0 {
			return false
		}
		if len(rest) > 0 && phpval.Truthy(rest[0]) {
			return h[:i]
		}
		return h[i:]
	})
	rt.RegisterFunc("strrchr", func(h, n string) any {
		if n == "" {
			return false
		}
		i := strings.LastIndexByte(h, n[0])
		if i < 0 {
			return false
		}
		return h[i:]
	})
	rt.RegisterFunc("chunk_split", func(s string, rest ...any) (string, error) {
		size, end := int64(76), "\r\n"
		if len(rest) > 0 && rest[0] != nil {
			size = phpval.Int(rest[0])
		}
		if len(rest) > 1 {
			end = phpval.String(rest[1])
		}
		if size < 1 {
			return "", &valueError{message: "chunk_split(): Argument #2 ($length) must be greater than 0"}
		}
		var b strings.Builder
		for i := 0; i < len(s); i += int(size) {
			b.WriteString(s[i:min(len(s), i+int(size))])
			b.WriteString(end)
		}
		return b.String(), nil
	})
	rt.RegisterFunc("quotemeta", func(s string) string {
		var b strings.Builder
		for i := 0; i < len(s); i++ {
			if strings.IndexByte(".\\+*?[^]$()", s[i]) >= 0 {
				b.WriteByte('\\')
			}
			b.WriteByte(s[i])
		}
		return b.String()
	})
	rt.RegisterFunc("count_chars", func(s string, rest ...any) (any, error) {
		mode := int64(0)
		if len(rest) > 0 {
			mode = phpval.Int(rest[0])
		}
		var counts [256]int64
		for i := 0; i < len(s); i++ {
			counts[s[i]]++
		}
		switch mode {
		case 0, 1, 2:
			out := model.NewArray()
			for c, n := range counts {
				if mode == 0 || mode == 1 && n > 0 || mode == 2 && n == 0 {
					out.Set(int64(c), n)
				}
			}
			return out, nil
		case 3, 4:
			var b strings.Builder
			for c, n := range counts {
				if mode == 3 && n > 0 || mode == 4 && n == 0 {
					b.WriteByte(byte(c))
				}
			}
			return b.String(), nil
		}
		return nil, &valueError{message: "count_chars(): Argument #2 ($mode) must be between 0 and 4 (inclusive)"}
	})
	rt.RegisterFunc("substr_compare", func(main, s string, offset int64, rest ...any) (int64, error) {
		if offset < 0 {
			offset = max(0, int64(len(main))+offset)
		}
		if offset > int64(len(main)) {
			return 0, &valueError{message: "substr_compare(): Argument #3 ($offset) must be contained in argument #1 ($main_str)"}
		}
		a, b := main[offset:], s
		if len(rest) > 0 && rest[0] != nil {
			n := int(phpval.Int(rest[0]))
			if n < 0 {
				return 0, &valueError{message: "substr_compare(): Argument #4 ($length) must be greater than or equal to 0"}
			}
			a, b = a[:min(len(a), n)], b[:min(len(b), n)]
		}
		if len(rest) > 1 && phpval.Truthy(rest[1]) {
			a, b = strings.ToLower(a), strings.ToLower(b)
		}
		return int64(strings.Compare(a, b)), nil
	})
	rt.RegisterFunc("str_word_count", func(s string, rest ...any) (any, error) {
		format := int64(0)
		if len(rest) > 0 && rest[0] != nil {
			format = phpval.Int(rest[0])
		}
		extra := ""
		if len(rest) > 1 && rest[1] != nil {
			extra = phpval.String(rest[1])
		}
		isWord := func(c byte) bool {
			return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '\'' || c == '-' || strings.IndexByte(extra, c) >= 0
		}
		out := model.NewArray()
		count := int64(0)
		for i := 0; i < len(s); {
			if !isWord(s[i]) || s[i] == '-' && !(strings.IndexByte(extra, '-') >= 0) && false {
				i++
				continue
			}
			j := i
			for j < len(s) && isWord(s[j]) {
				j++
			}
			word := s[i:j]
			// A word does not begin with a hyphen or end with one.
			start := i
			for strings.HasPrefix(word, "-") && !strings.Contains(extra, "-") {
				word, start = word[1:], start+1
			}
			word = strings.TrimRight(word, "-")
			if word != "" {
				count++
				switch format {
				case 1:
					out.Append(word)
				case 2:
					out.Set(int64(start), word)
				}
			}
			i = j
		}
		switch format {
		case 0:
			return count, nil
		case 1, 2:
			return out, nil
		}
		return nil, &valueError{message: "str_word_count(): Argument #2 ($format) must be a valid format value"}
	})
	rt.RegisterFunc("strcoll", func(a, b string) int64 { return int64(strings.Compare(a, b)) })
	rt.RegisterFunc("md5_file", func(name string, rest ...any) (any, error) { return fileDigest(rt, "md5", name, rest) })
	rt.RegisterFunc("sha1_file", func(name string, rest ...any) (any, error) { return fileDigest(rt, "sha1", name, rest) })
	rt.RegisterFunc("fprintf", func(stream any, format string, args ...any) (any, error) {
		return formatTo(rt, stream, "sprintf", format, args)
	})
	rt.RegisterFunc("vfprintf", func(stream any, format string, args any) (any, error) {
		return formatTo(rt, stream, "vsprintf", format, []any{args})
	})
	rt.RegisterFunc("vprintf", func(format string, args any) (any, error) {
		out, err := rt.InvokeNamed("vsprintf", format, args)
		if err != nil {
			return nil, err
		}
		s := fmt.Sprint(out)
		_, err = rt.Output().Write([]byte(s))
		return int64(len(s)), err
	})
	rt.RegisterFunc("soundex", soundex)
	rt.RegisterFunc("quoted_printable_encode", quotedPrintableEncode)
	rt.RegisterFunc("quoted_printable_decode", quotedPrintableDecode)
	rt.RegisterFunc("array_multisort", func(args ...any) (bool, error) { return multisort(args) })
	rt.RegisterFunc("hash_hmac_algos", func() any {
		algos := []string{"md5", "sha1", "sha224", "sha256", "sha384", "sha512", "sha512/224", "sha512/256", "sha3-224", "sha3-256", "sha3-384", "sha3-512", "ripemd160"}
		sort.Strings(algos)
		out := model.NewArray()
		for _, a := range algos {
			out.Append(a)
		}
		return out
	})
}

// fileDigest hashes a file's contents with md5 or sha1.
func fileDigest(rt *runner.Runtime, algo, name string, rest []any) (any, error) {
	content, err := os.ReadFile(name)
	if err != nil {
		get, ok := rt.Callable("file_get_contents")
		if !ok {
			return false, nil
		}
		res, gerr := get(name)
		s, isString := res.(string)
		if gerr != nil || !isString {
			return false, nil
		}
		content = []byte(s)
	}
	sum, err := rt.InvokeNamed(algo, string(content), len(rest) > 0 && phpval.Truthy(rest[0]))
	return sum, err
}

// formatTo renders a format with sprintf or vsprintf and writes it to a stream.
func formatTo(rt *runner.Runtime, stream any, fn, format string, args []any) (any, error) {
	out, err := rt.InvokeNamed(fn, append([]any{format}, args...)...)
	if err != nil {
		return nil, err
	}
	s := fmt.Sprint(out)
	write, ok := rt.Callable("fwrite")
	if !ok {
		return false, nil
	}
	if _, err := write(stream, s); err != nil {
		return nil, err
	}
	return int64(len(s)), nil
}

// soundex is php's soundex: a letter and three digits.
func soundex(s string) string {
	codes := "01230120022455012623010202"
	var out []byte
	var last byte
	for i := 0; i < len(s) && len(out) < 4; i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		if c < 'A' || c > 'Z' {
			continue
		}
		code := codes[c-'A']
		if len(out) == 0 {
			out = append(out, c)
			last = code
			continue
		}
		if code != last {
			if code != '0' {
				out = append(out, code)
			}
			if c != 'H' && c != 'W' {
				last = code
			}
		}
	}
	if len(out) == 0 {
		return ""
	}
	for len(out) < 4 {
		out = append(out, '0')
	}
	return string(out)
}

func quotedPrintableEncode(s string) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	lineLen := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\r' && i+1 < len(s) && s[i+1] == '\n' {
			b.WriteString("\r\n")
			i++
			lineLen = 0
			continue
		}
		var piece string
		if c < 32 && c != '\t' || c == '=' || c > 126 || (c == ' ' || c == '\t') && (i+1 == len(s) || s[i+1] == '\r') {
			piece = "=" + string(hex[c>>4]) + string(hex[c&15])
		} else {
			piece = string(c)
		}
		if lineLen+len(piece) > 75 {
			b.WriteString("=\r\n")
			lineLen = 0
		}
		b.WriteString(piece)
		lineLen += len(piece)
	}
	return b.String()
}

func quotedPrintableDecode(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			if i+2 < len(s)+0 && isHex(s[i+1]) && isHex(s[i+2]) {
				b.WriteByte(unhex(s[i+1])<<4 | unhex(s[i+2]))
				i += 2
				continue
			}
			j := i + 1
			for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
				j++
			}
			if j < len(s) && s[j] == '\r' && j+1 < len(s) && s[j+1] == '\n' {
				i = j + 1
				continue
			}
			if j < len(s) && s[j] == '\n' {
				i = j
				continue
			}
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func isHex(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

func unhex(c byte) byte {
	switch {
	case c >= 'a':
		return c - 'a' + 10
	case c >= 'A':
		return c - 'A' + 10
	}
	return c - '0'
}
