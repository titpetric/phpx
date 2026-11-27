//go:build !phpx

package phpx

import "github.com/titpetric/phpx/runner"

// The parser lowers `print x` to a call of __phpx_print in both builds, so the
// name has to answer where the phpx runtime is absent too. The phpx build
// registers it beside match; this is the same binding over the untagged string
// conversion.
func init() {
	runner.RegisterArea(`Runtime\PX\Logical\Print`, registerPrintStub)
}

// registerPrintStub installs print as echo that answers 1.
func registerPrintStub(rt *runner.Runtime) {
	rt.RegisterFunc("__phpx_print", func(v any) (int64, error) {
		if err := rt.WriteEcho(v); err != nil {
			return 0, err
		}
		return 1, nil
	})
}
