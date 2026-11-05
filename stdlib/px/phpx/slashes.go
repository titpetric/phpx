//go:build phpx

package phpx

import (
	"strings"

	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Slashes`, registerSlashes)
}

func registerSlashes(rt *runner.Runtime) {
	rt.RegisterFunc("addslashes", addSlashes)
	rt.RegisterFunc("stripslashes", stripSlashes)
}

// addSlashes backslash-escapes the quotes, the backslash and NUL.
func addSlashes(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		switch c := s[i]; c {
		case '\'', '"', '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
		case 0:
			b.WriteString("\\0")
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// stripSlashes undoes addSlashes: `\0` is NUL and `\x` is x.
func stripSlashes(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] != '\\' {
			b.WriteByte(s[i])
			continue
		}
		if i+1 >= len(s) {
			break
		}
		i++
		if s[i] == '0' {
			b.WriteByte(0)
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
