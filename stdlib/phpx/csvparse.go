//go:build phpx

package phpx

import (
	"strings"
	"unicode"
)

// csvLineEnd splits the line ending off the end of buf: php's
// php_fgetcsv_lookup_trailing_spaces, which strips a "\n", "\r\n" or "\r".
func csvLineEnd(buf string) (body, eol string) {
	switch {
	case strings.HasSuffix(buf, "\r\n"):
		return buf[:len(buf)-2], "\r\n"
	case strings.HasSuffix(buf, "\n"):
		return buf[:len(buf)-1], "\n"
	case strings.HasSuffix(buf, "\r"):
		return buf[:len(buf)-1], "\r"
	}
	return buf, ""
}

// csvRecord parses one record the way php_fgetcsv does. buf is the first line
// with its line ending; more supplies the next line of the stream when an
// enclosure is still open at the end of a line, or is nil for a plain string.
// An empty line is a record of one null. esc is -1 when there is no escape.
func csvRecord(buf string, more func() (string, bool), delim, enc byte, esc int) []any {
	body, eol := csvLineEnd(buf)
	if body == "" && eol != "" || body == "" && buf == "" {
		return []any{nil}
	}
	var fields []any
	pos := 0
	for {
		// A field's leading white space goes only when an enclosure follows it.
		start := pos
		for start < len(body) && body[start] != delim && unicode.IsSpace(rune(body[start])) {
			start++
		}
		var field strings.Builder
		if start < len(body) && body[start] == enc {
			pos = start + 1
			state := 0
			hunk := pos
		scan:
			for {
				if pos >= len(body) {
					switch state {
					case 2:
						field.WriteString(body[hunk : pos-1])
						hunk = pos
						break scan
					case 1:
						field.WriteString(body[hunk:pos])
						hunk = pos
						fallthrough
					case 0:
						field.WriteString(body[hunk:pos])
						field.WriteString(eol)
						if more == nil {
							hunk = pos
							break scan
						}
						next, ok := more()
						if !ok {
							hunk = pos
							break scan
						}
						body, eol = csvLineEnd(next)
						pos, hunk, state = 0, 0, 0
						continue
					}
				}
				c := body[pos]
				switch state {
				case 2:
					if c != enc {
						field.WriteString(body[hunk : pos-1])
						hunk = pos
						break scan
					}
					field.WriteString(body[hunk:pos])
					pos++
					hunk = pos
					state = 0
				case 1:
					pos++
					state = 0
				default:
					if c == enc {
						state = 2
					} else if esc >= 0 && int(c) == esc {
						state = 1
					}
					pos++
				}
			}
			// What follows the closing enclosure, up to the delimiter, is kept.
			for pos < len(body) && body[pos] != delim {
				pos++
			}
			field.WriteString(body[hunk:pos])
		} else {
			hunk := pos
			for pos < len(body) && body[pos] != delim {
				pos++
			}
			field.WriteString(body[hunk:pos])
		}
		fields = append(fields, field.String())
		if pos >= len(body) {
			break
		}
		pos++ // the delimiter
		if pos >= len(body) {
			fields = append(fields, "")
			break
		}
	}
	return fields
}
