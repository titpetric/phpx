//go:build phpx

package phpx

import (
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Mbstring`, registerPhpxMbstring)
}

// mbEncodings maps a lower-cased alias to the spelling mbstring reports and
// the codec that reads and writes it. UTF-8 and ASCII carry no codec: they
// are handled by hand.
var mbEncodings = func() map[string]mbEncoding {
	table := map[string]mbEncoding{}
	add := func(canonical string, codec encoding.Encoding, aliases ...string) {
		e := mbEncoding{name: canonical, codec: codec}
		table[strings.ToLower(canonical)] = e
		for _, a := range aliases {
			table[strings.ToLower(a)] = e
		}
	}
	add("UTF-8", nil, "utf8")
	add("ASCII", nil, "us-ascii", "ansi_x3.4-1968", "iso646-us")
	add("ISO-8859-1", charmap.ISO8859_1, "latin1", "iso8859-1", "iso_8859-1", "l1")
	isoMaps := map[string]encoding.Encoding{
		"ISO-8859-2": charmap.ISO8859_2, "ISO-8859-3": charmap.ISO8859_3, "ISO-8859-4": charmap.ISO8859_4,
		"ISO-8859-5": charmap.ISO8859_5, "ISO-8859-6": charmap.ISO8859_6, "ISO-8859-7": charmap.ISO8859_7,
		"ISO-8859-8": charmap.ISO8859_8, "ISO-8859-9": charmap.ISO8859_9, "ISO-8859-10": charmap.ISO8859_10,
		"ISO-8859-13": charmap.ISO8859_13, "ISO-8859-14": charmap.ISO8859_14, "ISO-8859-15": charmap.ISO8859_15,
		"ISO-8859-16": charmap.ISO8859_16,
	}
	for name, codec := range isoMaps {
		add(name, codec, strings.ReplaceAll(name, "ISO-8859", "iso8859"))
	}
	windows := map[string]encoding.Encoding{
		"1250": charmap.Windows1250, "1251": charmap.Windows1251, "1252": charmap.Windows1252,
		"1253": charmap.Windows1253, "1254": charmap.Windows1254, "1255": charmap.Windows1255,
		"1256": charmap.Windows1256, "1257": charmap.Windows1257, "1258": charmap.Windows1258,
	}
	for n, codec := range windows {
		add("Windows-"+n, codec, "cp"+n)
	}
	add("KOI8-R", charmap.KOI8R, "koi8r")
	add("KOI8-U", charmap.KOI8U, "koi8u")
	add("UTF-16", unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), "ucs-2", "ucs-2be")
	add("UTF-16BE", unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM))
	add("UTF-16LE", unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), "ucs-2le")
	add("UTF-32", utf32.UTF32(utf32.BigEndian, utf32.IgnoreBOM), "ucs-4", "ucs-4be", "utf-32be")
	add("UTF-32LE", utf32.UTF32(utf32.LittleEndian, utf32.IgnoreBOM), "ucs-4le")
	add("SJIS", japanese.ShiftJIS, "shift_jis", "sjis-win", "cp932", "ms_kanji")
	add("EUC-JP", japanese.EUCJP, "eucjp", "eucjp-win")
	add("JIS", japanese.ISO2022JP, "iso-2022-jp")
	add("EUC-KR", korean.EUCKR, "euckr", "uhc", "cp949")
	add("GB18030", simplifiedchinese.GB18030)
	add("CP936", simplifiedchinese.GBK, "gbk", "gb2312", "euc-cn")
	add("BIG-5", traditionalchinese.Big5, "big5", "cp950")
	add("8bit", nil, "binary", "7bit")
	return table
}()

type mbEncoding struct {
	name  string
	codec encoding.Encoding
}

// phpxEncodingName spells an encoding the way mbstring reports it.
func phpxEncodingName(name string) string {
	key := strings.ToLower(strings.TrimSpace(name))
	if e, ok := mbEncodings[key]; ok {
		return e.name
	}
	return strings.ToUpper(strings.TrimSpace(name))
}

// phpxEncodingList reads the encodings argument: a comma list or an array.
func phpxEncodingList(v any) []string {
	var out []string
	if t, ok := v.(string); ok {
		for _, part := range strings.Split(t, ",") {
			if part = strings.TrimSpace(part); part != "" {
				out = append(out, phpxEncodingName(part))
			}
		}
		return out
	}
	model.RangeValues(v, func(_, item any) bool {
		if s, ok := item.(string); ok {
			out = append(out, phpxEncodingName(s))
		}
		return true
	})
	return out
}

// mbDefaultOrder is what mb_detect_encoding tries for 'auto' or no list.
var mbDefaultOrder = []string{"ASCII", "UTF-8"}

// phpxValidIn reports whether s is a well-formed string of the encoding.
func phpxValidIn(s, name string) bool {
	return mbInvalid(s, name) == 0
}

// mbInvalid counts the sequences of s that are not valid in the encoding.
func mbInvalid(s, name string) int {
	e, known := mbEncodings[strings.ToLower(name)]
	switch {
	case !known:
		return 0
	case e.name == "ASCII":
		n := 0
		for i := 0; i < len(s); i++ {
			if s[i] >= 0x80 {
				n++
			}
		}
		return n
	case e.name == "UTF-8":
		n := 0
		for i := 0; i < len(s); {
			r, size := utf8.DecodeRuneInString(s[i:])
			if r == utf8.RuneError && size <= 1 {
				n++
			}
			i += max(size, 1)
		}
		return n
	case e.codec == nil || strings.HasPrefix(e.name, "ISO-8859") || strings.HasPrefix(e.name, "Windows-") || strings.HasPrefix(e.name, "KOI8"):
		return 0
	}
	decoded, err := e.codec.NewDecoder().String(s)
	if err != nil {
		return 1
	}
	return strings.Count(decoded, "�") - strings.Count(s, "�")
}

// phpxConvertEncoding converts s from one encoding to another. A byte the source
// cannot decode, and a character the target cannot hold, become '?'.
func phpxConvertEncoding(s, to, from string) string {
	if from == "" {
		from = "UTF-8"
	}
	target, ok := mbEncodings[strings.ToLower(to)]
	if !ok {
		return s
	}
	source, ok := mbEncodings[strings.ToLower(from)]
	if !ok {
		return s
	}
	if source.name == target.name {
		if target.name == "UTF-8" {
			return strings.ToValidUTF8(s, "?")
		}
		return s
	}
	text := s
	switch {
	case source.name == "UTF-8":
		text = strings.ToValidUTF8(s, "?")
	case source.name == "ASCII":
		text = strings.Map(func(r rune) rune {
			if r >= 0x80 {
				return '?'
			}
			return r
		}, s)
	case source.codec != nil:
		if decoded, err := source.codec.NewDecoder().String(s); err == nil {
			text = decoded
		}
	}
	switch {
	case target.name == "UTF-8":
		return text
	case target.name == "ASCII":
		return strings.Map(func(r rune) rune {
			if r >= 0x80 {
				return '?'
			}
			return r
		}, text)
	case target.codec != nil:
		var b strings.Builder
		enc := target.codec.NewEncoder()
		for _, r := range text {
			out, err := enc.String(string(r))
			if err != nil || out == "" {
				b.WriteByte('?')
				continue
			}
			b.WriteString(out)
		}
		return b.String()
	}
	return text
}

// registerPhpxMbstring installs the encoding detection and conversion.
func registerPhpxMbstring(rt *runner.Runtime) {
	rt.RegisterFunc("mb_detect_encoding", func(s string, rest ...any) any {
		candidates := mbDefaultOrder
		if len(rest) > 0 && rest[0] != nil {
			if t, ok := rest[0].(string); !ok || !strings.EqualFold(strings.TrimSpace(t), "auto") {
				if list := phpxEncodingList(rest[0]); len(list) > 0 {
					candidates = list
				}
			}
		}
		strict := len(rest) > 1 && phpval.Truthy(rest[1])
		best, fewest := "", -1
		for _, name := range candidates {
			n := mbInvalid(s, name)
			if n == 0 {
				return name
			}
			if fewest < 0 || n < fewest {
				best, fewest = name, n
			}
		}
		if strict {
			return false
		}
		return best
	})
	var convert func(v any, to, from string) any
	convert = func(v any, to, from string) any {
		switch t := v.(type) {
		case string:
			return phpxConvertEncoding(t, to, from)
		}
		if model.IsCollection(v) {
			out := model.NewArray()
			model.RangeValues(v, func(k, item any) bool {
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
	rt.RegisterFunc("mb_check_encoding", func(value any, rest ...any) bool {
		name := "UTF-8"
		if len(rest) > 0 && rest[0] != nil {
			name = phpxEncodingName(phpval.String(rest[0]))
		}
		ok := true
		check := func(item any) {
			if s, isString := item.(string); isString && !phpxValidIn(s, name) {
				ok = false
			}
		}
		if model.IsCollection(value) {
			model.RangeValues(value, func(k, item any) bool { check(k); check(item); return ok })
		} else {
			check(value)
		}
		return ok
	})
	rt.RegisterFunc("utf8_encode", func(s string) string { return phpxConvertEncoding(s, "UTF-8", "ISO-8859-1") })
	rt.RegisterFunc("utf8_decode", func(s string) string { return phpxConvertEncoding(s, "ISO-8859-1", "UTF-8") })
	iconv := func(from, to, s string) any {
		to = strings.SplitN(to, "//", 2)[0]
		from = strings.SplitN(from, "//", 2)[0]
		if _, ok := mbEncodings[strings.ToLower(to)]; !ok {
			return false
		}
		if _, ok := mbEncodings[strings.ToLower(from)]; !ok {
			return false
		}
		return phpxConvertEncoding(s, phpxEncodingName(to), phpxEncodingName(from))
	}
	rt.RegisterFunc("iconv", iconv)
	rt.RegisterFunc("iconv_strlen", func(s string, _ ...any) any {
		if !utf8.ValidString(s) {
			return false
		}
		return int64(utf8.RuneCountInString(s))
	})
}

// mbEncodingNames lists the encodings the runtime reads and writes, once each,
// in the spelling mbstring reports.
func mbEncodingNames() []string {
	seen := map[string]bool{}
	var out []string
	for _, e := range mbEncodings {
		if !seen[e.name] {
			seen[e.name] = true
			out = append(out, e.name)
		}
	}
	sort.Strings(out)
	return out
}
