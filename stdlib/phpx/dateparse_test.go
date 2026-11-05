//go:build phpx

package phpx

import (
	"testing"
	"time"
)

// base is Friday 2024-03-15 10:30:45 UTC, the day the php outputs below were taken on.
var fridayBase = time.Date(2024, 3, 15, 10, 30, 45, 0, time.UTC)

func TestParseDateCases(t *testing.T) {
	cases := map[string]string{
		"2024-06-01 12:00 America/New_York": "2024-06-01 16:00:00",
		"10 September 2000":                 "2000-09-10 00:00:00",
		"next week":                         "2024-03-18 10:30:45",
		"first day of next month":           "2024-04-01 10:30:45",
		"last friday of this month":         "2024-03-29 00:00:00",
		"+1 week 2 days":                    "2024-03-24 10:30:45",
		"3:30 PM":                           "2024-03-15 15:30:00",
		"2024-01-31 +1 month":               "2024-03-02 00:00:00",
	}
	for spec, want := range cases {
		got, _, ok := phpxStrtotime(spec, fridayBase)
		if !ok {
			t.Errorf("%q did not parse", spec)
			continue
		}
		if s := got.UTC().Format("2006-01-02 15:04:05"); s != want {
			t.Errorf("%q = %s, want %s", spec, s, want)
		}
	}
	if _, _, ok := phpxStrtotime("garbage", fridayBase); ok {
		t.Error("garbage parsed")
	}
}
