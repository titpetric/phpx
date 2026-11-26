//go:build phpx && !windows

package phpx

import (
	"syscall"

	"github.com/titpetric/phpx/model"
	"github.com/titpetric/phpx/runner"
)

// The rusage selectors php names. RUSAGE_CHILDREN is -1 in php and in the
// kernel's rusage(2); the two others follow it.
const (
	rusageSelf     = 0
	rusageChildren = -1
	rusageThread   = 1
)

func init() {
	runner.RegisterArea(`Runtime\PX\Rusage`, registerPhpxRusage)
}

// registerPhpxRusage installs getrusage(), which phpunit's telemetry reads to
// time a test. The array answers php's keys, so a consumer that reads one of
// them finds the value rather than a missing index.
func registerPhpxRusage(rt *runner.Runtime) {
	rt.RegisterFunc("getrusage", func(mode ...int64) any {
		who := rusageSelf
		if len(mode) > 0 {
			who = int(mode[0])
		}
		return phpxRusage(who)
	})
}

// phpxRusage reads one rusage record, or false for a selector the host lacks.
func phpxRusage(who int) any {
	var ru syscall.Rusage
	if err := syscall.Getrusage(who, &ru); err != nil {
		return false
	}
	out := model.NewArray()
	out.Set("ru_utime.tv_sec", ru.Utime.Sec)
	out.Set("ru_utime.tv_usec", ru.Utime.Usec)
	out.Set("ru_stime.tv_sec", ru.Stime.Sec)
	out.Set("ru_stime.tv_usec", ru.Stime.Usec)
	out.Set("ru_maxrss", ru.Maxrss)
	out.Set("ru_ixrss", ru.Ixrss)
	out.Set("ru_idrss", ru.Idrss)
	out.Set("ru_isrss", ru.Isrss)
	out.Set("ru_minflt", ru.Minflt)
	out.Set("ru_majflt", ru.Majflt)
	out.Set("ru_nswap", ru.Nswap)
	out.Set("ru_inblock", ru.Inblock)
	out.Set("ru_oublock", ru.Oublock)
	out.Set("ru_msgsnd", ru.Msgsnd)
	out.Set("ru_msgrcv", ru.Msgrcv)
	out.Set("ru_nsignals", ru.Nsignals)
	out.Set("ru_nvcsw", ru.Nvcsw)
	out.Set("ru_nivcsw", ru.Nivcsw)
	return out
}
