//go:build phpx

package phpx

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// A date string in the grammar php's strtotime and the DateTime constructors
// read: ISO and American and European dates, English month and weekday names,
// clock times, time zones, and relative phrases. The parts are read in any
// order, the way timelib reads them, and applied to a base time together.

type parsedDate struct {
	y, m, d       int
	haveDate      bool
	h, i, s, us   int
	haveTime      bool
	loc           *time.Location
	haveZone      bool
	stamp         *int64
	ry, rm, rd    int
	rh, ri, rs    int
	rus           int
	weekday       int // 0..6 or -1
	weekdayRel    int // the amount of a "next"/"last" weekday: 0 this/bare, +1 next, -1 last
	weekdayWeek   bool
	weekMonday    bool
	firstLast     int // 1 first day of, 2 last day of
	nthWeekday    int // "second tuesday of": the ordinal, 0 none, -1 last
	nthDay        int // the weekday of that phrase
	weekdays      int // +n weekdays
	timeReset     bool
	haveRelative  bool
	haveMonthOnly bool
}

var monthNames = map[string]int{
	"jan": 1, "january": 1, "feb": 2, "february": 2, "mar": 3, "march": 3, "apr": 4, "april": 4, "may": 5,
	"jun": 6, "june": 6, "jul": 7, "july": 7, "aug": 8, "august": 8, "sep": 9, "sept": 9, "september": 9,
	"oct": 10, "october": 10, "nov": 11, "november": 11, "dec": 12, "december": 12,
}

var weekdayNames = map[string]int{
	"sun": 0, "sunday": 0, "mon": 1, "monday": 1, "tue": 2, "tues": 2, "tuesday": 2, "wed": 3, "wednesday": 3,
	"thu": 4, "thur": 4, "thurs": 4, "thursday": 4, "fri": 5, "friday": 5, "sat": 6, "saturday": 6,
}

var ordinalWords = map[string]int{
	"first": 1, "second": 2, "third": 3, "fourth": 4, "fifth": 5, "sixth": 6, "seventh": 7, "eighth": 8,
	"ninth": 9, "tenth": 10, "eleventh": 11, "twelfth": 12, "next": 1, "last": -1, "previous": -1, "this": 0,
}

// zoneOffsets are the abbreviations php reads, in seconds east of UTC.
var zoneOffsets = map[string]int{
	"utc": 0, "gmt": 0, "z": 0, "est": -5 * 3600, "edt": -4 * 3600, "cst": -6 * 3600, "cdt": -5 * 3600,
	"mst": -7 * 3600, "mdt": -6 * 3600, "pst": -8 * 3600, "pdt": -7 * 3600, "cet": 3600, "cest": 7200,
	"eet": 7200, "eest": 10800, "wet": 0, "west": 3600, "bst": 3600, "ist": 19800, "jst": 9 * 3600,
	"kst": 9 * 3600, "hst": -10 * 3600, "akst": -9 * 3600, "akdt": -8 * 3600, "aest": 10 * 3600, "aedt": 11 * 3600,
}

var (
	reStamp        = regexp.MustCompile(`^@(-?\d+)(?:\.(\d{1,6}))?`)
	reISODate      = regexp.MustCompile(`^(\d{4})-(\d{1,2})-(\d{1,2})`)
	reSlashYMD     = regexp.MustCompile(`^(\d{4})/(\d{1,2})/(\d{1,2})`)
	reCompact      = regexp.MustCompile(`^(\d{4})(\d{2})(\d{2})(t?)(?:(\d{2}):?(\d{2}):?(\d{2})?)?`)
	reYearMonth    = regexp.MustCompile(`^(\d{4})-(\d{2})(?:[^\d-]|$)`)
	reAmerican     = regexp.MustCompile(`^(\d{1,2})/(\d{1,2})(?:/(\d{4}|\d{2}))?`)
	reDMY          = regexp.MustCompile(`^(\d{1,2})[-.](\d{1,2})[-.](\d{4})`)
	reDMY2         = regexp.MustCompile(`^(\d{1,2})-(\d{1,2})-(\d{2})(?:[^\d]|$)`)
	reClock        = regexp.MustCompile(`^t?(\d{1,2}):(\d{1,2})(?::(\d{2})(?:[.,:](\d+))?)?`)
	reMeridian     = regexp.MustCompile(`^\s*([ap])\.?m\.?`)
	reDotClock     = regexp.MustCompile(`^t?(\d{1,2})\.(\d{1,2})(?:\.(\d{2}))?(?:\s*([ap])\.?m\.?)?`)
	reHour12       = regexp.MustCompile(`^(\d{1,2})\s*([ap])\.?m\.?`)
	reZoneNum      = regexp.MustCompile(`^(?:gmt|utc)?([+-])(\d{1,2}):?(\d{2})?`)
	reZoneName     = regexp.MustCompile(`^[a-z]+/[a-z_/+-]+`)
	reWord         = regexp.MustCompile(`^[a-z]+`)
	reRelNum       = regexp.MustCompile(`^([+-]?)\s*(\d+)\s*([a-z]+)`)
	reRelWord      = regexp.MustCompile(`^(next|last|previous|this|first|second|third|fourth|fifth|sixth|seventh|eighth|ninth|tenth|eleventh|twelfth)\s+([a-z]+)`)
	reDayMonth     = regexp.MustCompile(`^(\d{1,2})(?:st|nd|rd|th)?[\s.,-]*([a-z]+)`)
	reDashMonthDMY = regexp.MustCompile(`^(\d{1,2})-([a-z]{3,9})-(\d{4}|\d{2})`)
	reDashMonthMDY = regexp.MustCompile(`^([a-z]{3,9})-(\d{1,2})-(\d{4})`)
	reDashYMonthD  = regexp.MustCompile(`^(\d{4})-([a-z]{3,9})-(\d{1,2})`)
	reDayOnly      = regexp.MustCompile(`^(\d{1,2})`)
	reDayNum       = regexp.MustCompile(`^(\d{1,2})(?:st|nd|rd|th)?(?:[\s.,-]|$)`)
	reYear4        = regexp.MustCompile(`^(\d{4})(?:[^\d]|$)`)
)

// unitOf names the field a unit word stands for: "sec", "min", "hour", "day",
// "week", "fortnight", "month", "year", "weekday"; "" for anything else.
func unitOf(w string) string {
	w = strings.TrimSuffix(w, "s")
	switch w {
	case "sec", "second":
		return "sec"
	case "min", "minute":
		return "min"
	case "hour":
		return "hour"
	case "day":
		return "day"
	case "week":
		return "week"
	case "fortnight", "forthnight":
		return "fortnight"
	case "month":
		return "month"
	case "year":
		return "year"
	case "weekday":
		return "weekday"
	case "msec", "millisecond", "ms":
		return "msec"
	case "usec", "microsecond", "µs":
		return "usec"
	}
	return ""
}

func (p *parsedDate) addRel(n int, unit string) {
	p.haveRelative = true
	switch unit {
	case "sec":
		p.rs += n
	case "min":
		p.ri += n
	case "hour":
		p.rh += n
	case "day":
		p.rd += n
	case "week":
		p.rd += 7 * n
	case "fortnight":
		p.rd += 14 * n
	case "month":
		p.rm += n
	case "year":
		p.ry += n
	case "weekday":
		p.weekdays += n
	case "msec":
		p.rus += n * 1000
	case "usec":
		p.rus += n
	}
}

func twoDigitYear(y int) int {
	if y < 70 {
		return 2000 + y
	}
	if y < 100 {
		return 1900 + y
	}
	return y
}

// parseDate reads s; ok is false when some part of it is not in the grammar.
func parseDate(s string) (*parsedDate, bool) {
	p := &parsedDate{weekday: -1}
	src := strings.ToLower(strings.TrimSpace(s))
	if src == "" {
		return p, true
	}
	pos := 0
	rest := func() string { return src[pos:] }
	for {
		for pos < len(src) && strings.IndexByte(" \t,\n\r", src[pos]) >= 0 {
			pos++
		}
		if pos >= len(src) {
			break
		}
		r := rest()
		if m := reStamp.FindStringSubmatch(r); m != nil {
			n, _ := strconv.ParseInt(m[1], 10, 64)
			p.stamp = &n
			if m[2] != "" {
				us, _ := strconv.Atoi((m[2] + "000000")[:6])
				p.us, p.haveTime = us, true
			}
			p.loc, p.haveZone = time.UTC, true
			pos += len(m[0])
			continue
		}
		if m := reCompact.FindStringSubmatch(r); m != nil && (m[4] == "t" || m[5] == "") {
			p.y, _ = strconv.Atoi(m[1])
			p.m, _ = strconv.Atoi(m[2])
			p.d, _ = strconv.Atoi(m[3])
			p.haveDate = true
			if m[5] != "" {
				p.h, _ = strconv.Atoi(m[5])
				p.i, _ = strconv.Atoi(m[6])
				if m[7] != "" {
					p.s, _ = strconv.Atoi(m[7])
				}
				p.haveTime = true
			}
			pos += len(m[0])
			continue
		}
		if m := reDashMonthDMY.FindStringSubmatch(r); m != nil {
			if mon, ok := monthNames[m[2]]; ok {
				p.d, p.m = mustAtoi(m[1]), mon
				y := mustAtoi(m[3])
				if len(m[3]) == 2 {
					y = twoDigitYear(y)
				}
				p.y, p.haveDate = y, true
				pos += len(m[0])
				continue
			}
		}
		if m := reDashMonthMDY.FindStringSubmatch(r); m != nil {
			if mon, ok := monthNames[m[1]]; ok {
				p.m, p.d, p.y, p.haveDate = mon, mustAtoi(m[2]), mustAtoi(m[3]), true
				pos += len(m[0])
				continue
			}
		}
		if m := reDashYMonthD.FindStringSubmatch(r); m != nil {
			if mon, ok := monthNames[m[2]]; ok {
				p.y, p.m, p.d, p.haveDate = mustAtoi(m[1]), mon, mustAtoi(m[3]), true
				pos += len(m[0])
				continue
			}
		}
		if m := reISODate.FindStringSubmatch(r); m != nil {
			if mo, d := mustAtoi(m[2]), mustAtoi(m[3]); mo < 1 || mo > 12 || d < 1 || d > 31 {
				return nil, false
			}
			p.y, _ = strconv.Atoi(m[1])
			p.m, _ = strconv.Atoi(m[2])
			p.d, _ = strconv.Atoi(m[3])
			p.haveDate = true
			pos += len(m[0])
			if pos < len(src) && src[pos] == 't' && pos+1 < len(src) && src[pos+1] >= '0' && src[pos+1] <= '9' {
				pos++
			}
			continue
		}
		if m := reSlashYMD.FindStringSubmatch(r); m != nil {
			p.y, _ = strconv.Atoi(m[1])
			p.m, _ = strconv.Atoi(m[2])
			p.d, _ = strconv.Atoi(m[3])
			p.haveDate = true
			pos += len(m[0])
			continue
		}
		if m := reYearMonth.FindStringSubmatch(r); m != nil {
			p.y, _ = strconv.Atoi(m[1])
			p.m, _ = strconv.Atoi(m[2])
			p.d, p.haveDate = 1, true
			pos += 7
			continue
		}
		if m := reAmerican.FindStringSubmatch(r); m != nil {
			p.m, _ = strconv.Atoi(m[1])
			p.d, _ = strconv.Atoi(m[2])
			if m[3] != "" {
				y, _ := strconv.Atoi(m[3])
				if len(m[3]) == 2 {
					y = twoDigitYear(y)
				}
				p.y = y
			}
			p.haveDate = true
			pos += len(m[0])
			continue
		}
		if m := reDMY.FindStringSubmatch(r); m != nil {
			p.d, _ = strconv.Atoi(m[1])
			p.m, _ = strconv.Atoi(m[2])
			p.y, _ = strconv.Atoi(m[3])
			p.haveDate = true
			pos += len(m[0])
			continue
		}
		if m := reDMY2.FindStringSubmatch(r); m != nil {
			p.d, _ = strconv.Atoi(m[1])
			p.m, _ = strconv.Atoi(m[2])
			y, _ := strconv.Atoi(m[3])
			p.y = twoDigitYear(y)
			p.haveDate = true
			pos += len(m[1]) + len(m[2]) + len(m[3]) + 2
			continue
		}
		if m := reClock.FindStringSubmatch(r); m != nil {
			p.h, _ = strconv.Atoi(m[1])
			p.i, _ = strconv.Atoi(m[2])
			if m[3] != "" {
				p.s, _ = strconv.Atoi(m[3])
			}
			if m[4] != "" {
				p.us, _ = strconv.Atoi((m[4] + "000000")[:6])
			}
			consumed := len(m[0])
			if mer := reMeridian.FindStringSubmatch(r[consumed:]); mer != nil && !startsLetter(r[consumed+len(mer[0]):]) {
				applyMeridian(p, mer[1])
				consumed += len(mer[0])
			}
			p.haveTime = true
			pos += consumed
			continue
		}
		if m := reHour12.FindStringSubmatch(r); m != nil && !startsLetter(r[len(m[0]):]) {
			p.h, _ = strconv.Atoi(m[1])
			p.i, p.s = 0, 0
			applyMeridian(p, m[2])
			p.haveTime = true
			pos += len(m[0])
			continue
		}
		if m := reDotClock.FindStringSubmatch(r); m != nil && !startsDigit(r[len(m[0]):]) {
			p.h, _ = strconv.Atoi(m[1])
			p.i, _ = strconv.Atoi(m[2])
			if m[3] != "" {
				p.s, _ = strconv.Atoi(m[3])
			}
			applyMeridian(p, m[4])
			p.haveTime = true
			pos += len(m[0])
			continue
		}
		if m := reRelNum.FindStringSubmatch(r); m != nil {
			if day, ok := weekdayNames[strings.TrimSuffix(m[3], "s")]; ok {
				n, _ := strconv.Atoi(m[2])
				if m[1] == "-" {
					n = -n
				}
				p.weekday, p.weekdayRel, p.timeReset = day, n, true
				pos += len(m[0])
				continue
			}
			if unit := unitOf(m[3]); unit != "" {
				n, _ := strconv.Atoi(m[2])
				if m[1] == "-" {
					n = -n
				}
				p.addRel(n, unit)
				pos += len(m[0])
				continue
			}
		}
		if m := reZoneNum.FindStringSubmatch(r); m != nil && (p.haveTime || p.haveDate || p.haveZone) {
			h, _ := strconv.Atoi(m[2])
			mi := 0
			if m[3] != "" {
				mi, _ = strconv.Atoi(m[3])
			}
			off := h*3600 + mi*60
			if m[1] == "-" {
				off = -off
			}
			p.loc, p.haveZone = time.FixedZone("", off), true
			pos += len(m[0])
			continue
		}
		if m := reZoneName.FindString(r); m != "" {
			name := titleZone(s, pos, len(m))
			if loc, err := time.LoadLocation(name); err == nil {
				p.loc, p.haveZone = loc, true
				pos += len(m)
				continue
			}
		}
		if m := reRelWord.FindStringSubmatch(r); m != nil {
			n := ordinalWords[m[1]]
			word := m[2]
			if word == "day" && len(r) > len(m[0]) && strings.HasPrefix(strings.TrimSpace(r[len(m[0]):]), "of") && (m[1] == "first" || m[1] == "last") {
				p.firstLast = 1
				if m[1] == "last" {
					p.firstLast = 2
				}
				pos += len(m[0])
				pos += len(r[len(m[0]):]) - len(strings.TrimLeft(r[len(m[0]):], " "))
				pos += 2
				p.haveRelative = true
				continue
			}
			if day, ok := weekdayNames[word]; ok {
				after := strings.TrimSpace(r[len(m[0]):])
				if strings.HasPrefix(after, "of") && (after == "of" || after[2] == ' ') {
					p.nthWeekday, p.nthDay = n, day
					if m[1] == "last" {
						p.nthWeekday = -1
					}
					p.timeReset, p.haveRelative = true, true
					pos += len(m[0]) + len(r[len(m[0]):]) - len(strings.TrimLeft(r[len(m[0]):], " ")) + 2
					continue
				}
				p.weekday, p.weekdayRel, p.timeReset = day, n, true
				if m[1] == "last" || m[1] == "previous" {
					p.weekdayRel = -1
				}
				if m[1] == "this" {
					p.weekdayRel = 0
				}
				pos += len(m[0])
				continue
			}
			if unit := unitOf(word); unit != "" {
				if m[1] == "this" {
					n = 0
				}
				if unit == "week" {
					// "next week", "last week", "this week": the Monday of that week.
					p.addRel(n, unit)
					p.weekMonday = true
				} else if n != 0 {
					p.addRel(n, unit)
				}
				pos += len(m[0])
				continue
			}
		}
		if m := reWord.FindString(r); m != "" {
			consumed := len(m)
			switch {
			case m == "now":
			case m == "today" || m == "midnight":
				p.timeReset, p.h, p.i, p.s = true, 0, 0, 0
				p.haveTime = true
			case m == "noon":
				p.h, p.i, p.s, p.haveTime = 12, 0, 0, true
			case m == "tomorrow":
				p.rd, p.haveRelative = 1, true
				p.timeReset, p.h, p.i, p.s, p.haveTime = true, 0, 0, 0, true
			case m == "yesterday":
				p.rd, p.haveRelative = -1, true
				p.timeReset, p.h, p.i, p.s, p.haveTime = true, 0, 0, 0, true
			case m == "ago":
				p.weekdayRel = -p.weekdayRel
				p.ry, p.rm, p.rd, p.rh, p.ri, p.rs, p.rus, p.weekdays = -p.ry, -p.rm, -p.rd, -p.rh, -p.ri, -p.rs, -p.rus, -p.weekdays
			case m == "of" || m == "at" || m == "on" || m == "and" || m == "the":
			case m == "weekdays" || m == "weekday":
				p.weekdays++
				p.haveRelative = true
			default:
				if mon, ok := monthNames[m]; ok {
					p.m = mon
					p.haveDate = true
					p.haveMonthOnly = p.d == 0
					// "Jan 1st 2000": a day of one or two digits after the month.
					after := r[consumed:]
					trim := len(after) - len(strings.TrimLeft(after, " .-"))
					if dm := reDayOnly.FindStringSubmatch(after[trim:]); dm != nil && !startsDigit(after[trim+len(dm[1]):]) && !strings.HasPrefix(after[trim+len(dm[1]):], ":") {
						p.d, _ = strconv.Atoi(dm[1])
						p.haveMonthOnly = false
						consumed += trim + len(dm[1])
						for _, suf := range []string{"st", "nd", "rd", "th"} {
							if strings.HasPrefix(r[consumed:], suf) {
								consumed += 2
								break
							}
						}
					}
				} else if day, ok := weekdayNames[m]; ok {
					p.weekday, p.timeReset = day, true
					if p.weekdayRel == 0 {
						p.weekdayRel = 0
					}
				} else if off, ok := zoneOffsets[m]; ok {
					p.loc, p.haveZone = time.FixedZone(strings.ToUpper(m), off), true
				} else if m == "t" {
				} else {
					return nil, false
				}
			}
			pos += consumed
			continue
		}
		// A bare number: a day before a month name ("1 Jan"), a year after one,
		// or a four digit clock ("2024" is HH:MM only when nothing else sets it).
		if m := reDayMonth.FindStringSubmatch(r); m != nil {
			if mon, ok := monthNames[m[2]]; ok {
				p.d, p.m, p.haveDate = mustAtoi(m[1]), mon, true
				pos += len(m[0])
				continue
			}
		}
		if m := reYear4.FindStringSubmatch(r); m != nil {
			p.y, _ = strconv.Atoi(m[1])
			p.haveDate = true
			if p.haveMonthOnly {
				p.d, p.haveMonthOnly = 1, false
			}
			pos += 4
			continue
		}
		return nil, false
	}
	return p, true
}

func mustAtoi(s string) int { n, _ := strconv.Atoi(s); return n }

// applyMeridian turns a 12 hour clock reading into 24 hours.
func applyMeridian(p *parsedDate, mer string) {
	switch mer {
	case "a":
		if p.h == 12 {
			p.h = 0
		}
	case "p":
		if p.h < 12 {
			p.h += 12
		}
	}
}

func startsLetter(s string) bool { return s != "" && (s[0] >= 'a' && s[0] <= 'z') }
func startsDigit(s string) bool  { return s != "" && s[0] >= '0' && s[0] <= '9' }

// titleZone gives a zone name its php spelling (America/New_York) from the
// original text of the string, which was lower-cased for matching.
func titleZone(orig string, pos, n int) string {
	trimmed := strings.TrimSpace(orig)
	// pos indexes the lower-cased trimmed string, of the same byte length.
	if pos+n > len(trimmed) {
		return trimmed[pos:]
	}
	return trimmed[pos : pos+n]
}

// resolve applies the parsed parts to a base time. base carries the zone a
// string without one is read in.
func (p *parsedDate) resolve(base time.Time) time.Time {
	if p.stamp != nil {
		t := time.Unix(*p.stamp, int64(p.us)*1000).In(time.UTC)
		return p.applyRelative(t)
	}
	loc := base.Location()
	if p.haveZone && p.loc != nil {
		loc = p.loc
		base = base.In(loc)
	}
	y, mo, d := base.Year(), int(base.Month()), base.Day()
	h, mi, s, ns := base.Hour(), base.Minute(), base.Second(), base.Nanosecond()
	if p.haveDate {
		if p.y != 0 {
			y = p.y
		}
		if p.m != 0 {
			mo = p.m
		}
		if p.d != 0 {
			d = p.d
		}
		if !p.haveTime {
			h, mi, s, ns = 0, 0, 0, 0
		}
	}
	if p.haveTime {
		h, mi, s, ns = p.h, p.i, p.s, p.us*1000
	}
	if p.timeReset && !p.haveTime {
		h, mi, s, ns = 0, 0, 0, 0
	}
	t := time.Date(y, time.Month(mo), d, h, mi, s, ns, loc)
	return p.applyRelative(t)
}

func (p *parsedDate) applyRelative(t time.Time) time.Time {
	loc := t.Location()
	y, mo, d := t.Year()+p.ry, int(t.Month())+p.rm, t.Day()
	switch p.firstLast {
	case 1:
		d = 1
	case 2:
		d = 1
	}
	t = time.Date(y, time.Month(mo), d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
	if p.firstLast == 2 {
		t = t.AddDate(0, 1, -1)
	}
	t = t.AddDate(0, 0, p.rd)
	t = t.Add(time.Duration(p.rh)*time.Hour + time.Duration(p.ri)*time.Minute + time.Duration(p.rs)*time.Second + time.Duration(p.rus)*time.Microsecond)
	if p.weekMonday && p.weekday < 0 {
		t = t.AddDate(0, 0, -((int(t.Weekday()) + 6) % 7))
	}
	if p.nthWeekday != 0 {
		first := time.Date(t.Year(), t.Month(), 1, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
		if p.timeReset {
			first = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc)
		}
		if p.nthWeekday > 0 {
			ahead := (p.nthDay - int(first.Weekday()) + 7) % 7
			t = first.AddDate(0, 0, ahead+7*(p.nthWeekday-1))
		} else {
			last := first.AddDate(0, 1, -1)
			back := (int(last.Weekday()) - p.nthDay + 7) % 7
			t = last.AddDate(0, 0, -back)
		}
	}
	if p.weekday >= 0 {
		cur := int(t.Weekday())
		switch {
		case p.weekMonday || p.weekdayWeek:
			// "monday next week": the weekday of the week the relative lands in,
			// weeks counted from Monday.
			monday := t.AddDate(0, 0, -((cur + 6) % 7))
			t = monday.AddDate(0, 0, (p.weekday+6)%7)
		case p.weekdayRel > 0:
			ahead := (p.weekday - cur + 7) % 7
			if ahead == 0 {
				ahead = 7
			}
			t = t.AddDate(0, 0, ahead+7*(p.weekdayRel-1))
		case p.weekdayRel < 0:
			back := (cur - p.weekday + 7) % 7
			if back == 0 {
				back = 7
			}
			t = t.AddDate(0, 0, -back)
		default:
			t = t.AddDate(0, 0, (p.weekday-cur+7)%7)
		}
		if p.timeReset && !p.haveTime {
			t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
		}
	}
	if p.weekdays != 0 {
		n := p.weekdays
		step := 1
		if n < 0 {
			step, n = -1, -n
		}
		for n > 0 {
			t = t.AddDate(0, 0, step)
			if wd := t.Weekday(); wd != time.Saturday && wd != time.Sunday {
				n--
			}
		}
	}
	return t
}

// phpxStrtotime reads a php date string against base.
func phpxStrtotime(spec string, base time.Time) (time.Time, *time.Location, bool) {
	p, ok := parseDate(spec)
	if !ok {
		return base, nil, false
	}
	t := p.resolve(base)
	if t.IsZero() && math.IsNaN(0) {
		return base, nil, false
	}
	var zone *time.Location
	if p.haveZone {
		zone = p.loc
	}
	return t, zone, true
}
