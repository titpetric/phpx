package stdlib

// Blank imports wire the standard binding packages into any binary that uses
// stdlib: each one contributes its runtime installer through runner.
// RegisterBinding in its init(), and Register runs them. A host needing a
// different set constructs its Runtime without this package, or passes extra
// bindings to Register.
import (
	_ "github.com/titpetric/phpx/stdlib/auth"
	_ "github.com/titpetric/phpx/stdlib/compat"
	_ "github.com/titpetric/phpx/stdlib/core"
	_ "github.com/titpetric/phpx/stdlib/crypto"
	_ "github.com/titpetric/phpx/stdlib/database"
	_ "github.com/titpetric/phpx/stdlib/files"
	_ "github.com/titpetric/phpx/stdlib/gd"
	_ "github.com/titpetric/phpx/stdlib/http"
	_ "github.com/titpetric/phpx/stdlib/info"
	_ "github.com/titpetric/phpx/stdlib/internals"
	_ "github.com/titpetric/phpx/stdlib/mail"
	_ "github.com/titpetric/phpx/stdlib/pexec"
	_ "github.com/titpetric/phpx/stdlib/phpx"
	_ "github.com/titpetric/phpx/stdlib/regexp"
	_ "github.com/titpetric/phpx/stdlib/runtime"
	_ "github.com/titpetric/phpx/stdlib/session"
	_ "github.com/titpetric/phpx/stdlib/span"
	_ "github.com/titpetric/phpx/stdlib/time"
)
