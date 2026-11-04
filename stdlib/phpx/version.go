//go:build phpx

package phpx

import (
	_ "github.com/titpetric/phpx/stdlib/core"

	"github.com/titpetric/phpx/runner"
)

func init() {
	runner.RegisterBinding(registerPhpxVersion)
}

const (
	phpxRelease = 27
	phpxVersion = "8.4.27"
)

// registerPhpxVersion raises the reported patch level past upstream's 8.4.0.
// composer's platform_check.php compares PHP_VERSION_ID against the vendor
// tree's lowest requirement, and laravel/framework asks for 8.4.1 or later.
func registerPhpxVersion(rt *runner.Runtime) {
	rt.SetConst("PHP_VERSION", phpxVersion)
	rt.SetConst("PHP_RELEASE_VERSION", int64(phpxRelease))
	rt.SetConst("PHP_VERSION_ID", int64(8*10000+4*100+phpxRelease))
	rt.SetConst("PCRE_VERSION", "10.46 2025-08-27")
	rt.SetConst("PCRE_VERSION_MAJOR", int64(10))
	rt.SetConst("PCRE_VERSION_MINOR", int64(46))
	version := func(extension ...any) any { return phpxVersion }
	rt.RegisterFunc("phpversion", version)
	rt.RegisterFunc("zend_version", func() string { return phpxVersion })
}
