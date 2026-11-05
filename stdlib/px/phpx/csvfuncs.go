//go:build phpx

package phpx

import (
	"strings"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\CSV`, registerCsvFuncs)
	runner.PhpxRegisterLate(registerCsvFiles)
}

// csvControls reads the delimiter, enclosure and escape arguments, each one
// byte; an empty escape turns escaping off.
func csvArgs(fn string, args []any) (delim, enc byte, esc int, err error) {
	delim, enc, esc = ',', '"', '\\'
	pick := func(i int, name string, single bool) (string, bool, error) {
		if i >= len(args) || args[i] == nil {
			return "", false, nil
		}
		s := phpval.String(args[i])
		if single && len(s) != 1 {
			return "", false, &valueError{message: fn + "(): Argument #" + string(rune('2'+i)) + " ($" + name + ") must be a single character"}
		}
		return s, true, nil
	}
	if s, ok, e := pick(0, "separator", true); e != nil {
		return 0, 0, 0, e
	} else if ok {
		delim = s[0]
	}
	if s, ok, e := pick(1, "enclosure", true); e != nil {
		return 0, 0, 0, e
	} else if ok {
		enc = s[0]
	}
	if len(args) > 2 && args[2] != nil {
		s := phpval.String(args[2])
		switch len(s) {
		case 0:
			esc = -1
		case 1:
			esc = int(s[0])
		default:
			return 0, 0, 0, &valueError{message: fn + "(): Argument #4 ($escape) must be empty or a single character"}
		}
	}
	return delim, enc, esc, nil
}

func csvArray(fields []any) *model.Array {
	out := model.NewArraySize(len(fields))
	for _, f := range fields {
		out.Append(f)
	}
	return out
}

// registerCsvFuncs installs str_getcsv and a fgetcsv that follows php's own
// parser: blank lines are a record of one null, the escape character is kept,
// and a field may run over several lines.
func registerCsvFuncs(rt *runner.Runtime) {
	rt.RegisterFunc("str_getcsv", func(s string, rest ...any) (any, error) {
		delim, enc, esc, err := csvArgs("str_getcsv", rest)
		if err != nil {
			return nil, err
		}
		return csvArray(csvRecord(s, nil, delim, enc, esc)), nil
	})
}

// registerCsvFiles installs the csv functions that read and write a stream, after
// the filesystem functions are rooted.
func registerCsvFiles(rt *runner.Runtime) {
	rt.RegisterFunc("fputcsv", func(stream any, fields any, rest ...any) (any, error) {
		delim, enc, esc, err := csvArgs("fputcsv", rest)
		if err != nil {
			return false, err
		}
		eol := "\n"
		if len(rest) > 3 && rest[3] != nil {
			eol = phpval.String(rest[3])
		}
		write, ok := rt.Callable("fwrite")
		if !ok {
			return false, nil
		}
		line := csvLine(fields, delim, enc, esc, eol)
		if _, err := write(stream, line); err != nil {
			return false, err
		}
		return int64(len(line)), nil
	})
	rt.RegisterFunc("fgetcsv", func(stream any, rest ...any) (any, error) {
		fgets, ok := rt.Callable("fgets")
		if !ok {
			return false, nil
		}
		var tail []any
		if len(rest) > 1 {
			tail = rest[1:]
		}
		delim, enc, esc, err := csvArgs("fgetcsv", tail)
		if err != nil {
			return nil, err
		}
		first, err := fgets(stream)
		line, isString := first.(string)
		if err != nil || !isString {
			return false, err
		}
		more := func() (string, bool) {
			next, err := fgets(stream)
			s, isString := next.(string)
			return s, err == nil && isString
		}
		return csvArray(csvRecord(line, more, delim, enc, esc)), nil
	})
}

// csvLine renders one record the way php_fputcsv does: a field is enclosed when
// it holds the delimiter, the enclosure, the escape character, a line break, a
// tab or a space, and the enclosure inside it is doubled unless escaped.
func csvLine(fields any, delim, enc byte, esc int, eol string) string {
	var out strings.Builder
	first := true
	model.RangeValues(fields, func(_, v any) bool {
		if !first {
			out.WriteByte(delim)
		}
		first = false
		s := phpval.String(v)
		needs := strings.IndexByte(s, delim) >= 0 || strings.IndexByte(s, enc) >= 0 ||
			esc >= 0 && strings.IndexByte(s, byte(esc)) >= 0 || strings.ContainsAny(s, "\n\r\t ")
		if !needs {
			out.WriteString(s)
			return true
		}
		out.WriteByte(enc)
		escaped := false
		for i := 0; i < len(s); i++ {
			c := s[i]
			switch {
			case esc >= 0 && int(c) == esc:
				escaped = true
			case !escaped && c == enc:
				out.WriteByte(enc)
			default:
				escaped = false
			}
			out.WriteByte(c)
		}
		out.WriteByte(enc)
		return true
	})
	out.WriteString(eol)
	return out.String()
}
