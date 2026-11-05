//go:build phpx

package phpx

import (
	"os"
	"sync"
	"time"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerDateZone)
}

// The default time zone of each runtime: date_default_timezone_set() moves it,
// and date(), strtotime(), mktime() and the DateTime constructors read it.
var dateZones sync.Map // *runner.Runtime -> *zoneBox

type zoneBox struct {
	mu  sync.Mutex
	loc *time.Location
}

func zoneOf(rt *runner.Runtime) *time.Location {
	v, _ := dateZones.LoadOrStore(rt, &zoneBox{loc: hostZone()})
	box := v.(*zoneBox)
	box.mu.Lock()
	defer box.mu.Unlock()
	return box.loc
}

func setZone(rt *runner.Runtime, loc *time.Location) {
	v, _ := dateZones.LoadOrStore(rt, &zoneBox{loc: hostZone()})
	box := v.(*zoneBox)
	box.mu.Lock()
	box.loc = loc
	box.mu.Unlock()
}

// hostZone is the zone php reads when none is set: TZ when it names one, and
// UTC otherwise.
func hostZone() *time.Location {
	if name := os.Getenv("TZ"); name != "" {
		if loc, err := time.LoadLocation(name); err == nil {
			return loc
		}
	}
	return time.UTC
}

func registerDateZone(rt *runner.Runtime) {
	rt.RegisterFunc("date_default_timezone_set", func(name string) bool {
		loc, err := phpxLocation(name)
		if err != nil {
			return false
		}
		setZone(rt, loc)
		return true
	})
	rt.RegisterFunc("date_default_timezone_get", func() string { return zoneOf(rt).String() })
	rt.RegisterFunc("strtotime", func(spec string, base ...any) any {
		now := time.Now().In(zoneOf(rt))
		if len(base) > 0 && base[0] != nil {
			now = time.Unix(phpval.Int(base[0]), 0).In(zoneOf(rt))
		}
		if spec == "" {
			return false
		}
		t, _, ok := phpxStrtotime(spec, now)
		if !ok {
			return false
		}
		return t.Unix()
	})
	rt.RegisterFunc("mktime", func(parts ...any) any { return mktime(zoneOf(rt), parts) })
	rt.RegisterFunc("gmmktime", func(parts ...any) any { return mktime(time.UTC, parts) })
	rt.RegisterFunc("checkdate", func(m, d, y int64) bool {
		if m < 1 || m > 12 || d < 1 || y < 1 || y > 32767 {
			return false
		}
		return d <= int64(time.Date(int(y), time.Month(m)+1, 0, 0, 0, 0, 0, time.UTC).Day())
	})
	rt.RegisterFunc("idate", func(format string, ts ...any) (any, error) {
		t := time.Now().In(zoneOf(rt))
		if len(ts) > 0 && ts[0] != nil {
			t = time.Unix(phpval.Int(ts[0]), 0).In(zoneOf(rt))
		}
		if len(format) != 1 {
			return nil, &valueError{message: "idate(): Argument #1 ($format) must be one character"}
		}
		out := phpxDateFormat(format, t)
		return phpval.Int(out), nil
	})
	rt.RegisterFunc("getdate", func(ts ...any) any {
		t := time.Now().In(zoneOf(rt))
		if len(ts) > 0 && ts[0] != nil {
			t = time.Unix(phpval.Int(ts[0]), 0).In(zoneOf(rt))
		}
		out := model.NewArray()
		out.Set("seconds", int64(t.Second()))
		out.Set("minutes", int64(t.Minute()))
		out.Set("hours", int64(t.Hour()))
		out.Set("mday", int64(t.Day()))
		out.Set("wday", int64(t.Weekday()))
		out.Set("mon", int64(t.Month()))
		out.Set("year", int64(t.Year()))
		out.Set("yday", int64(t.YearDay()-1))
		out.Set("weekday", t.Weekday().String())
		out.Set("month", t.Month().String())
		out.Set(int64(0), t.Unix())
		return out
	})
}

// mktime builds a timestamp from hour, minute, second, month, day and year,
// each defaulting to now; the fields may overflow and normalise.
func mktime(loc *time.Location, parts []any) any {
	now := time.Now().In(loc)
	v := []int{now.Hour(), now.Minute(), now.Second(), int(now.Month()), now.Day(), now.Year()}
	for i := 0; i < len(parts) && i < len(v); i++ {
		if parts[i] != nil {
			v[i] = int(phpval.Int(parts[i]))
		}
	}
	if len(parts) > 5 && parts[5] != nil {
		switch y := v[5]; {
		case y >= 0 && y < 70:
			v[5] = 2000 + y
		case y >= 70 && y <= 100:
			v[5] = 1900 + y
		}
	}
	return time.Date(v[5], time.Month(v[3]), v[4], v[0], v[1], v[2], 0, loc).Unix()
}
