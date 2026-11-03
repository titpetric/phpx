//go:build phpx

package phpx

import "github.com/titpetric/phpx/runner"

func init() {
	runner.RegisterBinding(registerPhpxCtype)
}

// registerPhpxCtype installs the ctype family, each over one string and one
// byte class. A non-string answers false, which is php 8.1's answer for an
// int outside the ASCII range and the only one a string-reading caller meets.
func registerPhpxCtype(rt *runner.Runtime) {
	classes := map[string]func(byte) bool{
		"ctype_digit": func(c byte) bool { return c >= '0' && c <= '9' },
		"ctype_alpha": func(c byte) bool { return c|0x20 >= 'a' && c|0x20 <= 'z' },
		"ctype_alnum": func(c byte) bool {
			return c >= '0' && c <= '9' || c|0x20 >= 'a' && c|0x20 <= 'z'
		},
		"ctype_upper": func(c byte) bool { return c >= 'A' && c <= 'Z' },
		"ctype_lower": func(c byte) bool { return c >= 'a' && c <= 'z' },
		"ctype_space": func(c byte) bool { return c == ' ' || c >= '\t' && c <= '\r' },
	}
	for name, in := range classes {
		rt.RegisterFunc(name, func(value any) bool {
			s, ok := value.(string)
			if !ok || s == "" {
				return false
			}
			for i := 0; i < len(s); i++ {
				if !in(s[i]) {
					return false
				}
			}
			return true
		})
	}
}
