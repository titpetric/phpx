//go:build phpx

package phpx

import (
	"fmt"
	"os"
	"time"

	"github.com/titpetric/phpx/internal/phpval"
	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxClock)
}

// registerPhpxClock installs the clock readers. The runtime keeps one zone,
// host zone, so gmdate() moves the stamp by the zone's offset before date()
// formats it, and the zone setter accepts only the name the host reads.
func registerPhpxClock(rt *runner.Runtime) {
	rt.RegisterFunc("gmdate", func(format string, timestamp ...any) (any, error) {
		date, ok := rt.Callable("date")
		if !ok {
			return false, nil
		}
		now := time.Now().Unix()
		if len(timestamp) > 0 {
			now = phpval.Int(timestamp[0])
		}
		// date() reads the host zone; moving the stamp by its offset prints
		// the UTC wall clock. The zone characters (e, T, O, P) still name the host.
		_, offset := time.Unix(now, 0).Zone()
		return date(format, now-int64(offset))
	})
	rt.RegisterFunc("date_default_timezone_set", func(zone string) bool { return zone == os.Getenv("TZ") || zone == "UTC" && os.Getenv("TZ") == "" })
	rt.RegisterFunc("hrtime", func(asNumber ...bool) any {
		now := time.Now().UnixNano()
		if len(asNumber) > 0 && asNumber[0] {
			return now
		}
		out := model.NewArray()
		out.Append(now / int64(time.Second))
		out.Append(now % int64(time.Second))
		return out
	})
	rt.RegisterFunc("uniqid", func(prefix ...string) string {
		now := time.Now()
		p := ""
		if len(prefix) > 0 {
			p = prefix[0]
		}
		return fmt.Sprintf("%s%08x%05x", p, now.Unix(), now.Nanosecond()/1000)
	})
}
