//go:build phpx

package phpx

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\DatetimeMore`, registerDateMore)
}

// instantOf reads the moment out of either date class.
func instantOf(v any) (time.Time, bool) {
	switch t := runner.PhpxUnwrapHost(v).(type) {
	case *DateTime:
		return t.t, true
	case *DateTimeImmutable:
		return t.t, true
	}
	return time.Time{}, false
}

// diffIntervals is DateTimeInterface::diff: the calendar difference counted
// from the earlier moment, the way timelib counts it, with the total in days.
func diffIntervals(a, b time.Time, absolute bool) *DateInterval {
	iv := &DateInterval{}
	one, two := a, b
	if one.After(two) {
		one, two = two, one
		iv.Invert = 1
	}
	// Both moments read in the first one's zone, as php does for a pair of zones.
	two = two.In(one.Location())
	y, m, d := two.Year()-one.Year(), int(two.Month())-int(one.Month()), two.Day()-one.Day()
	h, i, s := two.Hour()-one.Hour(), two.Minute()-one.Minute(), two.Second()-one.Second()
	us := (two.Nanosecond() - one.Nanosecond()) / 1000
	if us < 0 {
		us += 1000000
		s--
	}
	if s < 0 {
		s += 60
		i--
	}
	if i < 0 {
		i += 60
		h--
	}
	if h < 0 {
		h += 24
		d--
	}
	// A negative day count borrows days of a month. Counting forward, it is the
	// months before the end date, walking back; counting an inverted pair it is
	// the months from the earlier date, walking forward. php 8.1 does both.
	if iv.Invert == 0 {
		ey, em := two.Year(), int(two.Month())
		for d < 0 {
			em--
			if em < 1 {
				em, ey = 12, ey-1
			}
			d += daysIn(ey, em)
			m--
		}
	} else {
		by, bm := one.Year(), int(one.Month())
		for d < 0 {
			d += daysIn(by, bm)
			m--
			bm++
			if bm > 12 {
				bm, by = 1, by+1
			}
		}
	}
	if m < 0 {
		m += 12
		y--
	}
	iv.Y, iv.M, iv.D, iv.H, iv.I, iv.S = y, m, d, h, i, s
	iv.F = float64(us) / 1e6
	iv.Days = int64(two.Sub(one).Hours() / 24)
	if one.Location() != two.Location() || true {
		iv.Days = int64(daysBetween(one, two))
	}
	if absolute {
		iv.Invert = 0
	}
	return iv
}

func daysIn(y, m int) int { return time.Date(y, time.Month(m)+1, 0, 0, 0, 0, 0, time.UTC).Day() }

// daysBetween counts whole days between two moments by their calendar days and
// the time of day left over.
func daysBetween(one, two time.Time) int {
	a := time.Date(one.Year(), one.Month(), one.Day(), 0, 0, 0, 0, time.UTC)
	b := time.Date(two.Year(), two.Month(), two.Day(), 0, 0, 0, 0, time.UTC)
	days := int(b.Sub(a).Hours() / 24)
	if tod(two) < tod(one) {
		days--
	}
	return days
}

func tod(t time.Time) time.Duration {
	return time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute + time.Duration(t.Second())*time.Second + time.Duration(t.Nanosecond())
}

// Diff returns the interval to another moment.
func (d *DateTimeImmutable) Diff(other any, absolute ...bool) *DateInterval {
	o, _ := instantOf(other)
	return diffIntervals(d.t, o, len(absolute) > 0 && absolute[0])
}

// SetISODate sets the date from an ISO year, week and weekday.
func (d *DateTimeImmutable) SetISODate(year, week int64, dow ...int64) *DateTimeImmutable {
	return &DateTimeImmutable{t: isoDate(d.t, int(year), int(week), dowOf(dow))}
}

func dowOf(dow []int64) int {
	if len(dow) > 0 {
		return int(dow[0])
	}
	return 1
}

func isoDate(t time.Time, year, week, dow int) time.Time {
	jan4 := time.Date(year, 1, 4, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	monday := jan4.AddDate(0, 0, -((int(jan4.Weekday()) + 6) % 7))
	return monday.AddDate(0, 0, (week-1)*7+dow-1)
}

// GetMicrosecond is the microsecond part, php 8.4.
func (d *DateTimeImmutable) GetMicrosecond() int64 { return int64(d.t.Nanosecond() / 1000) }

// SetDate on a DateTime changes it in place.
func (d *DateTime) SetDate(y, m, day int64) *DateTime {
	d.t = d.DateTimeImmutable.SetDate(y, m, day).t
	return d
}

// SetTime on a DateTime changes it in place.
func (d *DateTime) SetTime(h, i int64, rest ...int64) *DateTime {
	d.t = d.DateTimeImmutable.SetTime(h, i, rest...).t
	return d
}

// SetISODate on a DateTime changes it in place.
func (d *DateTime) SetISODate(year, week int64, dow ...int64) *DateTime {
	d.t = isoDate(d.t, int(year), int(week), dowOf(dow))
	return d
}

// SetMicrosecond on a DateTime changes it in place, php 8.4.
func (d *DateTime) SetMicrosecond(us int64) *DateTime {
	d.t = time.Date(d.t.Year(), d.t.Month(), d.t.Day(), d.t.Hour(), d.t.Minute(), d.t.Second(), int(us)*1000, d.t.Location())
	return d
}

// SetMicrosecond returns the moment with the microsecond replaced.
func (d *DateTimeImmutable) SetMicrosecond(us int64) *DateTimeImmutable {
	return &DateTimeImmutable{t: time.Date(d.t.Year(), d.t.Month(), d.t.Day(), d.t.Hour(), d.t.Minute(), d.t.Second(), int(us)*1000, d.t.Location())}
}

// Format on an interval writes the %-directives of DateInterval::format().
func (i *DateInterval) Format(layout string) string {
	var b strings.Builder
	for n := 0; n < len(layout); n++ {
		if layout[n] != '%' || n+1 >= len(layout) {
			b.WriteByte(layout[n])
			continue
		}
		n++
		switch layout[n] {
		case 'Y':
			fmt.Fprintf(&b, "%04d", i.Y)
		case 'y':
			fmt.Fprintf(&b, "%d", i.Y)
		case 'M':
			fmt.Fprintf(&b, "%02d", i.M)
		case 'm':
			fmt.Fprintf(&b, "%d", i.M)
		case 'D':
			fmt.Fprintf(&b, "%02d", i.D)
		case 'd':
			fmt.Fprintf(&b, "%d", i.D)
		case 'H':
			fmt.Fprintf(&b, "%02d", i.H)
		case 'h':
			fmt.Fprintf(&b, "%d", i.H)
		case 'I':
			fmt.Fprintf(&b, "%02d", i.I)
		case 'i':
			fmt.Fprintf(&b, "%d", i.I)
		case 'S':
			fmt.Fprintf(&b, "%02d", i.S)
		case 's':
			fmt.Fprintf(&b, "%d", i.S)
		case 'F':
			fmt.Fprintf(&b, "%06d", int(i.F*1e6))
		case 'f':
			fmt.Fprintf(&b, "%d", int(i.F*1e6))
		case 'a':
			if days, ok := i.Days.(int64); ok {
				fmt.Fprintf(&b, "%d", days)
			} else {
				b.WriteString("(unknown)")
			}
		case 'R':
			if i.Invert == 1 {
				b.WriteByte('-')
			} else {
				b.WriteByte('+')
			}
		case 'r':
			if i.Invert == 1 {
				b.WriteByte('-')
			}
		case '%':
			b.WriteByte('%')
		default:
			b.WriteByte('%')
			b.WriteByte(layout[n])
		}
	}
	return b.String()
}

func errorsArray(e *dateErrors) any {
	if e == nil || (len(e.errors) == 0 && len(e.warnings) == 0) {
		return false
	}
	out := model.NewArray()
	out.Set("warning_count", int64(len(e.warnings)))
	out.Set("warnings", positions(e.warnings))
	out.Set("error_count", int64(len(e.errors)))
	out.Set("errors", positions(e.errors))
	return out
}

func positions(m map[int]string) *model.Array {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	out := model.NewArray()
	for _, k := range keys {
		out.Set(int64(k), m[k])
	}
	return out
}

func registerDateMore(rt *runner.Runtime) {
	var last *dateErrors
	build := func(mutable bool) func(format, input string, tz ...any) any {
		return func(format, input string, tz ...any) any {
			loc := zoneOf(rt)
			if len(tz) > 0 {
				if z, ok := tz[0].(*DateTimeZone); ok && z != nil {
					loc = z.loc
				}
			}
			t, zone, errs := createFromFormat(format, input, loc, time.Now())
			last = errs
			if len(errs.errors) > 0 {
				return false
			}
			if zone != nil {
				t = t.In(zone)
			}
			if mutable {
				return &DateTime{DateTimeImmutable{t: t}}
			}
			return &DateTimeImmutable{t: t}
		}
	}
	rt.RegisterFunc("DateTime::createFromFormat", build(true))
	rt.RegisterFunc("DateTimeImmutable::createFromFormat", build(false))
	rt.RegisterFunc("date_create_from_format", build(true))
	rt.RegisterFunc("date_create_immutable_from_format", build(false))
	lastErrors := func() any { return errorsArray(last) }
	rt.RegisterFunc("DateTime::getLastErrors", lastErrors)
	rt.RegisterFunc("DateTimeImmutable::getLastErrors", lastErrors)
	rt.RegisterFunc("date_get_last_errors", lastErrors)
	fromInterface := func(mutable bool) func(any) any {
		return func(v any) any {
			t, ok := instantOf(v)
			if !ok {
				return false
			}
			if mutable {
				return &DateTime{DateTimeImmutable{t: t}}
			}
			return &DateTimeImmutable{t: t}
		}
	}
	rt.RegisterFunc("DateTime::createFromInterface", fromInterface(true))
	rt.RegisterFunc("DateTime::createFromImmutable", fromInterface(true))
	rt.RegisterFunc("DateTimeImmutable::createFromInterface", fromInterface(false))
	rt.RegisterFunc("DateTimeImmutable::createFromMutable", fromInterface(false))
	rt.RegisterFunc("DateInterval::createFromDateString", func(spec string) any {
		p, ok := parseDate(spec)
		if !ok {
			return false
		}
		return &DateInterval{Y: p.ry, M: p.rm, D: p.rd, H: p.rh, I: p.ri, S: p.rs, F: float64(p.rus) / 1e6}
	})
	rt.RegisterFunc("date_interval_create_from_date_string", func(spec string) any {
		p, ok := parseDate(spec)
		if !ok {
			return false
		}
		return &DateInterval{Y: p.ry, M: p.rm, D: p.rd, H: p.rh, I: p.ri, S: p.rs, F: float64(p.rus) / 1e6}
	})
	rt.RegisterFunc("date_interval_format", func(i *DateInterval, layout string) string { return i.Format(layout) })
	rt.RegisterFunc("DateTimeZone::listIdentifiers", func(rest ...any) any { return zoneIdentifiers() })
	rt.RegisterFunc("timezone_identifiers_list", func(rest ...any) any { return zoneIdentifiers() })
	rt.RegisterFunc("timezone_open", func(name string) any {
		loc, err := phpxLocation(name)
		if err != nil {
			return false
		}
		return &DateTimeZone{loc: loc}
	})
	rt.RegisterFunc("timezone_name_get", func(z *DateTimeZone) string { return z.GetName() })
	rt.RegisterFunc("timezone_offset_get", func(z *DateTimeZone, at any) int64 {
		t, _ := instantOf(at)
		_, off := t.In(z.loc).Zone()
		return int64(off)
	})
	rt.RegisterFunc("date_create", func(rest ...any) any {
		t, err := phpxParseDate(rest, zoneOf(rt))
		if err != nil {
			return false
		}
		return &DateTime{DateTimeImmutable{t: t}}
	})
	rt.RegisterFunc("date_create_immutable", func(rest ...any) any {
		t, err := phpxParseDate(rest, zoneOf(rt))
		if err != nil {
			return false
		}
		return &DateTimeImmutable{t: t}
	})
	rt.RegisterFunc("date_format", func(d any, layout string) string {
		t, _ := instantOf(d)
		return phpxFormat(t, layout)
	})
	rt.RegisterFunc("date_timestamp_get", func(d any) int64 { t, _ := instantOf(d); return t.Unix() })
	rt.RegisterFunc("date_offset_get", func(d any) int64 { t, _ := instantOf(d); _, off := t.Zone(); return int64(off) })
	rt.RegisterFunc("date_timezone_get", func(d any) any {
		t, ok := instantOf(d)
		if !ok {
			return false
		}
		return &DateTimeZone{loc: t.Location()}
	})
	rt.RegisterFunc("date_diff", func(a, b any, absolute ...bool) any {
		ta, _ := instantOf(a)
		tb, _ := instantOf(b)
		return diffIntervals(ta, tb, len(absolute) > 0 && absolute[0])
	})
	rt.RegisterFunc("date_parse", func(spec string) any { return dateParseArray(spec) })
}

// zoneIdentifiers lists the zone names the host knows, region by region.
func zoneIdentifiers() any {
	out := model.NewArray()
	var names []string
	for _, root := range []string{"/usr/share/zoneinfo", "/usr/lib/zoneinfo", "/usr/share/lib/zoneinfo"} {
		for _, region := range []string{"Africa", "America", "Antarctica", "Arctic", "Asia", "Atlantic", "Australia", "Europe", "Indian", "Pacific"} {
			_ = filepath.WalkDir(filepath.Join(root, region), func(p string, d os.DirEntry, err error) error {
				if err == nil && !d.IsDir() {
					if rel, rerr := filepath.Rel(root, p); rerr == nil {
						names = append(names, rel)
					}
				}
				return nil
			})
		}
		if len(names) > 0 {
			break
		}
	}
	names = append(names, "UTC")
	sort.Strings(names)
	for _, n := range names {
		out.Append(n)
	}
	return out
}

// dateParseArray is date_parse(): the parts the string names, false for the
// ones it does not.
func dateParseArray(spec string) any {
	out := model.NewArray()
	p, ok := parseDate(spec)
	field := func(name string, set bool, v int) {
		if set {
			out.Set(name, int64(v))
		} else {
			out.Set(name, false)
		}
	}
	if !ok {
		p = &parsedDate{weekday: -1}
	}
	field("year", p.haveDate && p.y != 0, p.y)
	field("month", p.haveDate && p.m != 0, p.m)
	field("day", p.haveDate && p.d != 0, p.d)
	field("hour", p.haveTime, p.h)
	field("minute", p.haveTime, p.i)
	field("second", p.haveTime, p.s)
	if p.haveTime {
		out.Set("fraction", float64(p.us)/1e6)
	} else {
		out.Set("fraction", false)
	}
	errCount := 0
	if !ok {
		errCount = 1
	}
	out.Set("warning_count", int64(0))
	out.Set("warnings", model.NewArray())
	out.Set("error_count", int64(errCount))
	errs := model.NewArray()
	if !ok {
		errs.Set(int64(0), "The timezone could not be found in the database")
	}
	out.Set("errors", errs)
	out.Set("is_localtime", p.haveZone)
	return out
}
