//go:build phpx

package phpx

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterArea(`Runtime\PX\Datetime`, registerPhpxDateTime)
}

// DateTimeZone is php's DateTimeZone over a Go location.
type DateTimeZone struct{ loc *time.Location }

// DateTimeImmutable is php's DateTimeImmutable; DateTime shares the type's
// behaviour and mutates in place.
type DateTimeImmutable struct{ t time.Time }

// DateTime is php's DateTime.
type DateTime struct{ DateTimeImmutable }

// DateInterval holds the fields of an ISO 8601 duration.
type DateInterval struct {
	Y, M, D, H, I, S int
	Invert           int
	F                float64
	// Days is the total number of days of an interval diff() made, and false
	// for one built from a spec.
	Days any
}

func registerPhpxDateTime(rt *runner.Runtime) {
	rt.RegisterConstructor("DateTimeZone", func(name string) (*DateTimeZone, error) {
		loc, err := phpxLocation(name)
		if err != nil {
			return nil, fmt.Errorf("DateTimeZone::__construct(): Unknown or bad timezone (%s)", name)
		}
		return &DateTimeZone{loc: loc}, nil
	})
	rt.RegisterConstructor("DateInterval", func(spec string) (*DateInterval, error) { return phpxInterval(spec) })
	rt.RegisterConstructor("DateTimeImmutable", func(args ...any) (*DateTimeImmutable, error) {
		t, err := phpxParseDate(args, zoneOf(rt))
		return &DateTimeImmutable{t: t}, err
	})
	rt.RegisterConstructor("DateTime", func(args ...any) (*DateTime, error) {
		t, err := phpxParseDate(args, zoneOf(rt))
		return &DateTime{DateTimeImmutable{t: t}}, err
	})
}

func phpxLocation(name string) (*time.Location, error) {
	switch strings.ToUpper(name) {
	case "UTC", "Z", "":
		return time.UTC, nil
	}
	if len(name) >= 5 && (name[0] == '+' || name[0] == '-') {
		var h, m int
		if _, err := fmt.Sscanf(strings.ReplaceAll(name[1:], ":", ""), "%02d%02d", &h, &m); err == nil {
			off := h*3600 + m*60
			if name[0] == '-' {
				off = -off
			}
			return time.FixedZone(name, off), nil
		}
	}
	return time.LoadLocation(name)
}

func phpxParseDate(args []any, defaults ...*time.Location) (time.Time, error) {
	loc := time.UTC
	if len(defaults) > 0 {
		loc = defaults[0]
	}
	spec := "now"
	if len(args) > 0 {
		if s, ok := args[0].(string); ok && s != "" {
			spec = s
		}
	}
	if len(args) > 1 {
		if z, ok := args[1].(*DateTimeZone); ok && z != nil {
			loc = z.loc
		}
	}
	return phpxParseSpec(spec, time.Now().In(loc), loc)
}

var phpxLayouts = []string{
	"2006-01-02T15:04:05.999999999Z07:00", "2006-01-02T15:04:05Z07:00", "2006-01-02 15:04:05.999999999", "2006-01-02 15:04:05",
	"2006-01-02T15:04:05.999999999", "2006-01-02T15:04:05", "2006-01-02 15:04", "2006-01-02", "2006/01/02", "Mon, 02 Jan 2006 15:04:05 -0700",
}

// phpxParseSpec reads the formats strtotime-style constructors are given most:
// now, today, tomorrow, @epoch, ISO dates and "+1 day" style relative offsets.
func phpxParseSpec(spec string, now time.Time, loc *time.Location) (time.Time, error) {
	if t, zone, ok := phpxStrtotime(spec, now); ok {
		if zone != nil {
			t = t.In(zone)
		}
		return t, nil
	}
	s := strings.TrimSpace(spec)
	switch strings.ToLower(s) {
	case "now", "":
		return now, nil
	case "today", "midnight":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc), nil
	case "tomorrow":
		return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc), nil
	case "yesterday":
		return time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, loc), nil
	}
	if strings.HasPrefix(s, "@") {
		f, err := strconv.ParseFloat(s[1:], 64)
		if err != nil {
			return now, fmt.Errorf("Failed to parse time string (%s)", spec)
		}
		return time.Unix(int64(f), int64((f-float64(int64(f)))*1e9)).In(time.UTC), nil
	}
	for _, l := range phpxLayouts {
		if t, err := time.ParseInLocation(l, s, loc); err == nil {
			return t, nil
		}
	}
	// A trailing zone name or offset moves the rest into that zone.
	if fields := strings.Fields(s); len(fields) > 1 {
		last := fields[len(fields)-1]
		var zone *time.Location
		switch strings.ToUpper(last) {
		case "UTC", "GMT", "Z":
			zone = time.UTC
		default:
			if z, err := phpxLocation(last); err == nil {
				zone = z
			}
		}
		if zone != nil {
			rest := strings.Join(fields[:len(fields)-1], " ")
			if t, err := phpxParseSpec(rest, now.In(zone), zone); err == nil {
				return t, nil
			}
		}
	}
	if t, ok := phpxRelative(s, now); ok {
		return t, nil
	}
	return now, fmt.Errorf("DateTime::__construct(): Failed to parse time string (%s) at position 0 (%c): The timezone could not be found in the database", spec, s[0])
}

func phpxRelative(s string, base time.Time) (time.Time, bool) {
	fields := strings.Fields(strings.ToLower(s))
	t := base
	ok := false
	for i := 0; i+1 < len(fields); i += 2 {
		n, err := strconv.Atoi(strings.TrimPrefix(fields[i], "+"))
		if err != nil {
			return base, false
		}
		unit := strings.TrimSuffix(fields[i+1], "s")
		switch unit {
		case "sec", "second":
			t = t.Add(time.Duration(n) * time.Second)
		case "min", "minute":
			t = t.Add(time.Duration(n) * time.Minute)
		case "hour":
			t = t.Add(time.Duration(n) * time.Hour)
		case "day":
			t = t.AddDate(0, 0, n)
		case "week":
			t = t.AddDate(0, 0, 7*n)
		case "month":
			t = t.AddDate(0, n, 0)
		case "year":
			t = t.AddDate(n, 0, 0)
		default:
			return base, false
		}
		ok = true
	}
	return t, ok
}

func phpxInterval(spec string) (*DateInterval, error) {
	if !strings.HasPrefix(spec, "P") {
		return nil, fmt.Errorf("DateInterval::__construct(): Unknown or bad format (%s)", spec)
	}
	d := &DateInterval{Days: false}
	inTime, num := false, ""
	for _, r := range spec[1:] {
		switch {
		case r == 'T':
			inTime = true
		case r >= '0' && r <= '9':
			num += string(r)
		default:
			n, _ := strconv.Atoi(num)
			num = ""
			switch {
			case r == 'Y':
				d.Y = n
			case r == 'M' && !inTime:
				d.M = n
			case r == 'W':
				d.D += 7 * n
			case r == 'D':
				d.D += n
			case r == 'H':
				d.H = n
			case r == 'M':
				d.I = n
			case r == 'S':
				d.S = n
			default:
				return nil, fmt.Errorf("DateInterval::__construct(): Unknown or bad format (%s)", spec)
			}
		}
	}
	return d, nil
}

// GetName is the zone's name.
func (z *DateTimeZone) GetName() string { return z.loc.String() }

// GetOffset is the zone's offset from UTC in seconds at the given moment.
func (z *DateTimeZone) GetOffset(at *DateTimeImmutable) int64 {
	_, off := at.t.In(z.loc).Zone()
	return int64(off)
}

// Format writes the moment as php's date() does.
func (d *DateTimeImmutable) Format(layout string) string { return phpxFormat(d.t, layout) }

// GetTimestamp is the Unix time in seconds.
func (d *DateTimeImmutable) GetTimestamp() int64 { return d.t.Unix() }

// GetTimezone is the moment's zone.
func (d *DateTimeImmutable) GetTimezone() *DateTimeZone { return &DateTimeZone{loc: d.t.Location()} }

// GetOffset is the offset from UTC in seconds.
func (d *DateTimeImmutable) GetOffset() int64 { _, o := d.t.Zone(); return int64(o) }

// SetTimezone moves the moment to a zone, returning a new value.
func (d *DateTimeImmutable) SetTimezone(z *DateTimeZone) *DateTimeImmutable {
	return &DateTimeImmutable{t: d.t.In(z.loc)}
}

// SetTimestamp returns the moment at a Unix time.
func (d *DateTimeImmutable) SetTimestamp(ts int64) *DateTimeImmutable {
	return &DateTimeImmutable{t: time.Unix(ts, 0).In(d.t.Location())}
}

// SetDate replaces the date.
func (d *DateTimeImmutable) SetDate(y, m, day int64) *DateTimeImmutable {
	return &DateTimeImmutable{t: time.Date(int(y), time.Month(m), int(day), d.t.Hour(), d.t.Minute(), d.t.Second(), d.t.Nanosecond(), d.t.Location())}
}

// SetTime replaces the time of day.
func (d *DateTimeImmutable) SetTime(h, i int64, rest ...int64) *DateTimeImmutable {
	var s, us int64
	if len(rest) > 0 {
		s = rest[0]
	}
	if len(rest) > 1 {
		us = rest[1]
	}
	return &DateTimeImmutable{t: time.Date(d.t.Year(), d.t.Month(), d.t.Day(), int(h), int(i), int(s), int(us)*1000, d.t.Location())}
}

// Modify applies a relative or absolute change.
func (d *DateTimeImmutable) Modify(spec string) (*DateTimeImmutable, error) {
	t, err := phpxParseSpec(spec, d.t, d.t.Location())
	return &DateTimeImmutable{t: t}, err
}

func (d *DateTimeImmutable) shift(i *DateInterval, sign int) *DateTimeImmutable {
	if i.Invert == 1 {
		sign = -sign
	}
	t := d.t.AddDate(sign*i.Y, sign*i.M, sign*i.D)
	t = t.Add(time.Duration(sign) * (time.Duration(i.H)*time.Hour + time.Duration(i.I)*time.Minute + time.Duration(i.S)*time.Second))
	return &DateTimeImmutable{t: t}
}

// Add returns the moment plus an interval.
func (d *DateTimeImmutable) Add(i *DateInterval) *DateTimeImmutable { return d.shift(i, 1) }

// Sub returns the moment minus an interval.
func (d *DateTimeImmutable) Sub(i *DateInterval) *DateTimeImmutable { return d.shift(i, -1) }

// Format on a DateTime, which mutates in place for the setters below.
func (d *DateTime) Format(layout string) string { return d.DateTimeImmutable.Format(layout) }

// SetTimezone moves this DateTime to a zone and returns it.
func (d *DateTime) SetTimezone(z *DateTimeZone) *DateTime { d.t = d.t.In(z.loc); return d }

// SetTimestamp sets this DateTime to a Unix time and returns it.
func (d *DateTime) SetTimestamp(ts int64) *DateTime {
	d.t = time.Unix(ts, 0).In(d.t.Location())
	return d
}

// Modify changes this DateTime and returns it.
func (d *DateTime) Modify(spec string) (*DateTime, error) {
	t, err := phpxParseSpec(spec, d.t, d.t.Location())
	d.t = t
	return d, err
}

// Add adds an interval in place.
func (d *DateTime) Add(i *DateInterval) *DateTime { d.t = d.shift(i, 1).t; return d }

// Sub subtracts an interval in place.
func (d *DateTime) Sub(i *DateInterval) *DateTime { d.t = d.shift(i, -1).t; return d }

func phpxFormat(t time.Time, layout string) string {
	var b strings.Builder
	for i := 0; i < len(layout); i++ {
		c := layout[i]
		switch c {
		case '\\':
			if i+1 < len(layout) {
				i++
				b.WriteByte(layout[i])
			}
		case 'd':
			b.WriteString(t.Format("02"))
		case 'D':
			b.WriteString(t.Format("Mon"))
		case 'j':
			b.WriteString(strconv.Itoa(t.Day()))
		case 'l':
			b.WriteString(t.Format("Monday"))
		case 'N':
			w := int(t.Weekday())
			if w == 0 {
				w = 7
			}
			b.WriteString(strconv.Itoa(w))
		case 'w':
			b.WriteString(strconv.Itoa(int(t.Weekday())))
		case 'z':
			b.WriteString(strconv.Itoa(t.YearDay() - 1))
		case 'W':
			_, w := t.ISOWeek()
			fmt.Fprintf(&b, "%02d", w)
		case 'F':
			b.WriteString(t.Format("January"))
		case 'm':
			b.WriteString(t.Format("01"))
		case 'M':
			b.WriteString(t.Format("Jan"))
		case 'n':
			b.WriteString(strconv.Itoa(int(t.Month())))
		case 't':
			b.WriteString(strconv.Itoa(time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()))
		case 'L':
			if y := t.Year(); y%4 == 0 && (y%100 != 0 || y%400 == 0) {
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
			b.WriteString(t.Format("06"))
		case 'a':
			b.WriteString(t.Format("pm"))
		case 'A':
			b.WriteString(t.Format("PM"))
		case 'g':
			b.WriteString(t.Format("3"))
		case 'G':
			b.WriteString(strconv.Itoa(t.Hour()))
		case 'h':
			b.WriteString(t.Format("03"))
		case 'H':
			b.WriteString(t.Format("15"))
		case 'i':
			b.WriteString(t.Format("04"))
		case 's':
			b.WriteString(t.Format("05"))
		case 'u':
			fmt.Fprintf(&b, "%06d", t.Nanosecond()/1000)
		case 'v':
			fmt.Fprintf(&b, "%03d", t.Nanosecond()/1000000)
		case 'e':
			b.WriteString(t.Location().String())
		case 'I':
			if t.IsDST() {
				b.WriteByte('1')
			} else {
				b.WriteByte('0')
			}
		case 'O':
			b.WriteString(t.Format("-0700"))
		case 'P':
			b.WriteString(t.Format("-07:00"))
		case 'p':
			if _, off := t.Zone(); off == 0 {
				b.WriteByte('Z')
			} else {
				b.WriteString(t.Format("-07:00"))
			}
		case 'T':
			b.WriteString(t.Format("MST"))
		case 'Z':
			_, off := t.Zone()
			b.WriteString(strconv.Itoa(off))
		case 'c':
			b.WriteString(t.Format("2006-01-02T15:04:05-07:00"))
		case 'r':
			b.WriteString(t.Format("Mon, 02 Jan 2006 15:04:05 -0700"))
		case 'U':
			b.WriteString(strconv.FormatInt(t.Unix(), 10))
		case 'S':
			d := t.Day()
			suf := "th"
			if d%100 < 11 || d%100 > 13 {
				switch d % 10 {
				case 1:
					suf = "st"
				case 2:
					suf = "nd"
				case 3:
					suf = "rd"
				}
			}
			b.WriteString(suf)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}
