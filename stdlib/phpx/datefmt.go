//go:build phpx

package phpx

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxDate)
}

// registerPhpxDate installs date() and gmdate() with every format character php
// defines, in the host zone and in UTC.
func registerPhpxDate(rt *runner.Runtime) {
	stamp := func(ts []any) time.Time {
		if len(ts) > 0 && ts[0] != nil {
			return time.Unix(phpval.Int(ts[0]), 0).In(zoneOf(rt))
		}
		return time.Now().In(zoneOf(rt))
	}
	rt.RegisterFunc("date", func(format string, ts ...any) string {
		return phpxDateFormat(format, stamp(ts), false)
	})
	rt.RegisterFunc("gmdate", func(format string, ts ...any) string {
		return phpxDateFormat(format, stamp(ts).UTC(), true)
	})
}

// phpxDateFormat renders t as php's date() renders it with format.
func phpxDateFormat(format string, t time.Time, gmt ...bool) string {
	var b strings.Builder
	pad2 := func(n int) string { return fmt.Sprintf("%02d", n) }
	for i := 0; i < len(format); i++ {
		c := format[i]
		switch c {
		case '\\':
			if i+1 < len(format) {
				i++
				b.WriteByte(format[i])
			}
		case 'd':
			b.WriteString(pad2(t.Day()))
		case 'D':
			b.WriteString(t.Weekday().String()[:3])
		case 'j':
			b.WriteString(strconv.Itoa(t.Day()))
		case 'l':
			b.WriteString(t.Weekday().String())
		case 'N':
			w := int(t.Weekday())
			if w == 0 {
				w = 7
			}
			b.WriteString(strconv.Itoa(w))
		case 'S':
			d := t.Day()
			switch {
			case d%10 == 1 && d != 11:
				b.WriteString("st")
			case d%10 == 2 && d != 12:
				b.WriteString("nd")
			case d%10 == 3 && d != 13:
				b.WriteString("rd")
			default:
				b.WriteString("th")
			}
		case 'w':
			b.WriteString(strconv.Itoa(int(t.Weekday())))
		case 'z':
			b.WriteString(strconv.Itoa(t.YearDay() - 1))
		case 'W':
			_, w := t.ISOWeek()
			b.WriteString(pad2(w))
		case 'F':
			b.WriteString(t.Month().String())
		case 'm':
			b.WriteString(pad2(int(t.Month())))
		case 'M':
			b.WriteString(t.Month().String()[:3])
		case 'n':
			b.WriteString(strconv.Itoa(int(t.Month())))
		case 't':
			b.WriteString(strconv.Itoa(time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()))
		case 'L':
			y := t.Year()
			if y%4 == 0 && (y%100 != 0 || y%400 == 0) {
				b.WriteByte('1')
			} else {
				b.WriteByte('0')
			}
		case 'o':
			y, _ := t.ISOWeek()
			b.WriteString(strconv.Itoa(y))
		case 'Y':
			b.WriteString(strconv.Itoa(t.Year()))
		case 'y':
			b.WriteString(pad2(t.Year() % 100))
		case 'a':
			if t.Hour() < 12 {
				b.WriteString("am")
			} else {
				b.WriteString("pm")
			}
		case 'A':
			if t.Hour() < 12 {
				b.WriteString("AM")
			} else {
				b.WriteString("PM")
			}
		case 'B':
			u := t.UTC()
			beat := ((u.Hour()*3600 + u.Minute()*60 + u.Second() + 3600) % 86400) * 1000 / 86400
			b.WriteString(fmt.Sprintf("%03d", beat))
		case 'g':
			h := t.Hour() % 12
			if h == 0 {
				h = 12
			}
			b.WriteString(strconv.Itoa(h))
		case 'G':
			b.WriteString(strconv.Itoa(t.Hour()))
		case 'h':
			h := t.Hour() % 12
			if h == 0 {
				h = 12
			}
			b.WriteString(pad2(h))
		case 'H':
			b.WriteString(pad2(t.Hour()))
		case 'i':
			b.WriteString(pad2(t.Minute()))
		case 's':
			b.WriteString(pad2(t.Second()))
		case 'u':
			b.WriteString(fmt.Sprintf("%06d", t.Nanosecond()/1000))
		case 'v':
			b.WriteString(fmt.Sprintf("%03d", t.Nanosecond()/1000000))
		case 'e':
			name := t.Location().String()
			if name == "Local" {
				name = "UTC"
			}
			b.WriteString(name)
		case 'I':
			if t.IsDST() {
				b.WriteByte('1')
			} else {
				b.WriteByte('0')
			}
		case 'O':
			b.WriteString(phpxOffset(t, false))
		case 'P':
			b.WriteString(phpxOffset(t, true))
		case 'p':
			if _, off := t.Zone(); off == 0 {
				b.WriteByte('Z')
			} else {
				b.WriteString(phpxOffset(t, true))
			}
		case 'T':
			name, off := t.Zone()
			if len(gmt) > 0 && gmt[0] {
				name = "GMT"
			}
			if name == "" || strings.HasPrefix(name, "+") || strings.HasPrefix(name, "-") {
				name = phpxOffset(t, true)
			}
			if name == "UTC" && off == 0 && t.Location() == time.UTC {
				name = "UTC"
			}
			b.WriteString(name)
		case 'Z':
			_, off := t.Zone()
			b.WriteString(strconv.Itoa(off))
		case 'c':
			b.WriteString(phpxDateFormat("Y-m-d\\TH:i:sP", t, gmt...))
		case 'r':
			b.WriteString(phpxDateFormat("D, d M Y H:i:s O", t, gmt...))
		case 'U':
			b.WriteString(strconv.FormatInt(t.Unix(), 10))
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

// phpxOffset is the UTC offset of t as +0100, or +01:00 when colon is set.
func phpxOffset(t time.Time, colon bool) string {
	_, off := t.Zone()
	sign := '+'
	if off < 0 {
		sign = '-'
		off = -off
	}
	if colon {
		return fmt.Sprintf("%c%02d:%02d", sign, off/3600, off%3600/60)
	}
	return fmt.Sprintf("%c%02d%02d", sign, off/3600, off%3600/60)
}
