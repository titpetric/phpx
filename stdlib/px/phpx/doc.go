// Package phpx holds the bindings the fork adds after stdlib/core has
// installed its own, so a constant or function it registers wins. The files
// carry the phpx build tag; the untagged build compiles only the stubs that
// answer a name the parser lowers in both builds.
package phpx
