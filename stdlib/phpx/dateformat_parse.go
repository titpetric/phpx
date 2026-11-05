//go:build phpx

package phpx

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// dateErrors is what DateTime::getLastErrors() reports: warnings and errors by
// the position in the input they were found at.
type dateErrors struct {
	warnings map[int]string
	errors   map[int]string
}

func (e *dateErrors) fail(pos int, msg string) {
	if e.errors == nil {
		e.errors = map[int]string{}
	}
	if _, set := e.errors[pos]; !set {
		e.errors[pos] = msg
	}
}

// createFromFormat reads input by a DateTime::createFromFormat format. now
// supplies the fields the format leaves out; ok is false when the input does
// not match, with the reasons in errs.
func createFromFormat(format, input string, loc *time.Location, now time.Time) (t time.Time, zone *time.Location, errs *dateErrors) {
	errs = &dateErrors{}
	type fields struct {
		y, m, d, h, i, s, us   int
		ys, ms, ds, hs, is, ss bool
		us_                    bool
		pm, haveAMPM           bool
		doy                    int
		haveDOY                bool
		stamp                  *int64
	}
	var f fields
	reset := false
	pos := 0
	num := func(max int) (int, bool) {
		start := pos
		for pos < len(input) && pos-start < max && input[pos] >= '0' && input[pos] <= '9' {
			pos++
		}
		if pos == start {
			return 0, false
		}
		n, _ := strconv.Atoi(input[start:pos])
		return n, true
	}
	word := func() string {
		start := pos
		for pos < len(input) && (input[pos] >= 'a' && input[pos] <= 'z' || input[pos] >= 'A' && input[pos] <= 'Z') {
			pos++
		}
		return strings.ToLower(input[start:pos])
	}
	for fi := 0; fi < len(format); fi++ {
		c := format[fi]
		switch c {
		case 'd', 'j':
			n, ok := num(2)
			if !ok {
				errs.fail(pos, "A two digit day could not be found")
				return
			}
			f.d, f.ds = n, true
		case 'S':
			if pos+2 <= len(input) {
				switch strings.ToLower(input[pos : pos+2]) {
				case "st", "nd", "rd", "th":
					pos += 2
					continue
				}
			}
			errs.fail(pos, "A two digit day could not be found")
			return
		case 'z':
			n, ok := num(3)
			if !ok {
				errs.fail(pos, "A three digit day-of-year could not be found")
				return
			}
			f.doy, f.haveDOY = n, true
		case 'D', 'l':
			if _, ok := weekdayNames[word()]; !ok {
				errs.fail(pos, "A textual day could not be found")
				return
			}
		case 'm', 'n':
			n, ok := num(2)
			if !ok {
				errs.fail(pos, "A two digit month could not be found")
				return
			}
			f.m, f.ms = n, true
		case 'M', 'F':
			mon, ok := monthNames[word()]
			if !ok {
				errs.fail(pos, "A textual month could not be found")
				return
			}
			f.m, f.ms = mon, true
		case 'y':
			n, ok := num(2)
			if !ok {
				errs.fail(pos, "A two digit year could not be found")
				return
			}
			f.y, f.ys = twoDigitYear(n), true
		case 'Y':
			n, ok := num(4)
			if !ok {
				errs.fail(pos, "A four digit year could not be found")
				return
			}
			f.y, f.ys = n, true
		case 'a', 'A':
			w := word()
			if w != "am" && w != "pm" {
				errs.fail(pos, "A meridian could not be found")
				return
			}
			f.pm, f.haveAMPM = w == "pm", true
		case 'g', 'h', 'G', 'H':
			n, ok := num(2)
			if !ok {
				errs.fail(pos, "A two digit hour could not be found")
				return
			}
			f.h, f.hs = n, true
		case 'i':
			n, ok := num(2)
			if !ok {
				errs.fail(pos, "A two digit minute could not be found")
				return
			}
			f.i, f.is = n, true
		case 's':
			n, ok := num(2)
			if !ok {
				errs.fail(pos, "A two digit second could not be found")
				return
			}
			f.s, f.ss = n, true
		case 'u', 'v':
			start := pos
			max := 6
			if c == 'v' {
				max = 3
			}
			for pos < len(input) && pos-start < max && input[pos] >= '0' && input[pos] <= '9' {
				pos++
			}
			if pos == start {
				errs.fail(pos, "A six digit microsecond could not be found")
				return
			}
			digits := (input[start:pos] + "000000")[:6]
			f.us, _ = strconv.Atoi(digits)
			f.us_ = true
		case 'U':
			start := pos
			if pos < len(input) && input[pos] == '-' {
				pos++
			}
			for pos < len(input) && input[pos] >= '0' && input[pos] <= '9' {
				pos++
			}
			n, err := strconv.ParseInt(input[start:pos], 10, 64)
			if err != nil {
				errs.fail(pos, "A unix timestamp could not be found")
				return
			}
			f.stamp = &n
		case 'e', 'T', 'O', 'P', 'p':
			rest := input[pos:]
			end := 0
			for end < len(rest) && rest[end] != ' ' && rest[end] != ')' {
				end++
			}
			name := rest[:end]
			if name == "" {
				errs.fail(pos, "The timezone could not be found in the database")
				return
			}
			if z, ok := parseZoneToken(name); ok {
				zone = z
				pos += end
			} else {
				errs.fail(pos, "The timezone could not be found in the database")
				return
			}
		case ' ':
			if pos < len(input) && (input[pos] == ' ' || input[pos] == '\t') {
				pos++
			} else {
				errs.fail(pos, "The separation symbol could not be found")
				return
			}
		case '#':
			if pos < len(input) && strings.IndexByte(";:/.,-()", input[pos]) >= 0 {
				pos++
			} else {
				errs.fail(pos, "The separation symbol ([;:/.,-]) could not be found")
				return
			}
		case ';', ':', '/', '.', ',', '-', '(', ')':
			if pos < len(input) && input[pos] == c {
				pos++
			} else {
				errs.fail(pos, "The separation symbol could not be found")
				return
			}
		case '?':
			if pos < len(input) {
				pos++
			} else {
				errs.fail(pos, "Unexpected data found.")
				return
			}
		case '*':
			for pos < len(input) && strings.IndexByte(" ,;:/.-()", input[pos]) < 0 {
				pos++
			}
		case '!':
			reset = true
			f = fields{}
		case '|':
			reset = true
		case '+':
			pos = len(input)
		case '\\':
			fi++
			if fi < len(format) && pos < len(input) && input[pos] == format[fi] {
				pos++
			} else {
				errs.fail(pos, "The escaped character could not be found")
				return
			}
		default:
			if pos < len(input) && input[pos] == c {
				pos++
			} else {
				errs.fail(pos, "The format separator does not match")
				return
			}
		}
	}
	if pos < len(input) {
		errs.fail(pos, "Trailing data")
		return
	}
	if f.stamp != nil {
		return time.Unix(*f.stamp, int64(f.us)*1000).In(time.UTC), time.UTC, errs
	}
	where := loc
	if zone != nil {
		where = zone
	}
	base := now.In(where)
	if reset {
		base = time.Date(1970, 1, 1, 0, 0, 0, 0, where)
	}
	y, mo, d := base.Year(), int(base.Month()), base.Day()
	h, mi, s, ns := base.Hour(), base.Minute(), base.Second(), base.Nanosecond()
	if f.ys {
		y = f.y
	}
	if f.ms {
		mo = f.m
	}
	if f.ds {
		d = f.d
	}
	if f.hs || f.is || f.ss || f.us_ {
		// A time field in the format zeroes the time fields it leaves out.
		h, mi, s, ns = f.h, f.i, f.s, f.us*1000
	}
	if f.haveAMPM {
		switch {
		case f.pm && h < 12:
			h += 12
		case !f.pm && h == 12:
			h = 0
		}
	}
	if f.haveDOY {
		mo, d = 1, f.doy+1
	}
	return time.Date(y, time.Month(mo), d, h, mi, s, ns, where), zone, errs
}

// parseZoneToken reads a zone written as an abbreviation, an offset or a name.
func parseZoneToken(s string) (*time.Location, bool) {
	if off, ok := zoneOffsets[strings.ToLower(s)]; ok {
		return time.FixedZone(strings.ToUpper(s), off), true
	}
	if m := reZoneNum.FindStringSubmatch(strings.ToLower(s)); m != nil && m[0] == strings.ToLower(s) {
		h, _ := strconv.Atoi(m[2])
		mi := 0
		if m[3] != "" {
			mi, _ = strconv.Atoi(m[3])
		}
		off := h*3600 + mi*60
		if m[1] == "-" {
			off = -off
		}
		return time.FixedZone(s, off), true
	}
	if loc, err := time.LoadLocation(s); err == nil {
		return loc, true
	}
	return nil, false
}

// errorsArray reports the positions and messages in php's shape.
func (e *dateErrors) message() string {
	for pos, msg := range e.errors {
		return fmt.Sprintf("%s (at position %d)", msg, pos)
	}
	return ""
}
