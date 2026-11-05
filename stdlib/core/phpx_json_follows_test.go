//go:build phpx

package core_test

// jsonFollowsPHP is true in the fork's build, where stdlib/phpx replaces
// json_encode with php's flags and slash escaping.
const jsonFollowsPHP = true
