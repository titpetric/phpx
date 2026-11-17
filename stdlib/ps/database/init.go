package database

import (
	"github.com/titpetric/phpx/runner"
)

// init contributes the Database and Database\Migrate bindings.
func init() {
	runner.RegisterArea(`Runtime\PS\Database`, Register)
}
