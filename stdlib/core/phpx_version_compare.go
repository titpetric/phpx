//go:build phpx

package core

import (
	"strconv"
	"strings"

	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxVersionCompare)
}

// registerPhpxVersionCompare installs version_compare(), which framework
// bootstrap code uses to gate on the php and package versions.
func registerPhpxVersionCompare(rt *runner.Runtime) {
	rt.RegisterFunc("version_compare", func(a, b string, op ...string) any {
		c := phpxVersionCmp(a, b)
		if len(op) == 0 || op[0] == "" {
			return c
		}
		switch op[0] {
		case "<", "lt":
			return c < 0
		case "<=", "le":
			return c <= 0
		case ">", "gt":
			return c > 0
		case ">=", "ge":
			return c >= 0
		case "==", "eq":
			return c == 0
		case "!=", "<>", "ne":
			return c != 0
		}
		return nil
	})
}

// phpxVersionCanon puts a dot between a digit run and a letter run and turns
// the separators - _ + into dots, as php's canonicalisation does.
func phpxVersionCanon(v string) []string {
	var parts []string
	var cur []byte
	flush := func() {
		if len(cur) > 0 {
			parts = append(parts, string(cur))
			cur = cur[:0]
		}
	}
	isDigit := func(c byte) bool { return c >= '0' && c <= '9' }
	for i := 0; i < len(v); i++ {
		c := v[i]
		switch {
		case c == '-' || c == '_' || c == '+' || c == '.':
			flush()
		case len(cur) > 0 && isDigit(cur[len(cur)-1]) != isDigit(c):
			flush()
			cur = append(cur, c)
		default:
			cur = append(cur, c)
		}
	}
	flush()
	return parts
}

func phpxVersionOrder(s string) int {
	for _, f := range []struct {
		name string
		n    int
	}{{"dev", 0}, {"alpha", 1}, {"a", 1}, {"beta", 2}, {"b", 2}, {"RC", 3}, {"rc", 3}, {"#", 4}, {"pl", 5}, {"p", 5}} {
		if strings.HasPrefix(s, f.name) {
			return f.n
		}
	}
	return -6
}

func phpxVersionIsDigit(s string) bool { return s != "" && s[0] >= '0' && s[0] <= '9' }

func phpxVersionCmp(a, b string) int {
	if a == "" && b == "" {
		return 0
	}
	if a == "" {
		return -1
	}
	if b == "" {
		return 1
	}
	pa, pb := phpxVersionCanon(a), phpxVersionCanon(b)
	cmp := func(x, y int) int {
		switch {
		case x < y:
			return -1
		case x > y:
			return 1
		}
		return 0
	}
	i := 0
	for ; i < len(pa) && i < len(pb); i++ {
		x, y := pa[i], pb[i]
		var c int
		switch {
		case phpxVersionIsDigit(x) && phpxVersionIsDigit(y):
			nx, _ := strconv.ParseInt(x, 10, 64)
			ny, _ := strconv.ParseInt(y, 10, 64)
			c = cmp(int(nx), int(ny))
		case !phpxVersionIsDigit(x) && !phpxVersionIsDigit(y):
			c = cmp(phpxVersionOrder(x), phpxVersionOrder(y))
		case phpxVersionIsDigit(x):
			c = cmp(phpxVersionOrder("#"), phpxVersionOrder(y))
		default:
			c = cmp(phpxVersionOrder(x), phpxVersionOrder("#"))
		}
		if c != 0 {
			return c
		}
	}
	if i < len(pa) {
		if phpxVersionIsDigit(pa[i]) {
			return 1
		}
		return phpxVersionCmp(strings.Join(pa[i:], "."), "#N#")
	}
	if i < len(pb) {
		if phpxVersionIsDigit(pb[i]) {
			return -1
		}
		return phpxVersionCmp("#N#", strings.Join(pb[i:], "."))
	}
	return 0
}
