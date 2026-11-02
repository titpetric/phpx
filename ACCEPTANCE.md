# ACCEPTANCE

What phpx owes, one box per claim, with the command that settles it. The claims come from [README.phpx.md](README.phpx.md); nothing here adds a goal that file does not name.

Two rules hold over every box. `atkins -f atkins.phpx.yml` passes before a box counts, because a reversal that breaks the runtime it reverses is no reversal. A box that names php behaviour ticks on the `--matrix` run and not on phpx alone: the expected output comes from running the source through `php` first, and output copied from phpx locks in its bugs.

`phpx` below is the binary `atkins -f atkins.phpx.yml build` installs. A tree with no GOBIN reads `bin/phpx` for the same command.

## The harness

- [X] The fork publishes to `titpetric/phpx`, with upstream on `titpetric/phpscript`. `git remote get-url origin` and `git remote get-url upstream`.
- [X] The module path is `github.com/titpetric/phpx`. `head -1 go.mod`, and `go build ./...` resolves every import of it.
- [X] `README.md` is the copy of `README.phpx.md`. `cmp README.md README.phpx.md`.
- [X] A `README.md` edited by hand cannot land. `git add README.md && git commit` exits non-zero through `lefthook.yml`, and `git push` with a drifted copy does the same.
- [X] Every job in `atkins.yml` has a counterpart in `atkins.phpx.yml`. `atkins -l -f atkins.phpx.yml` lists all 24, with `test:phpscript:matrix` and `test:phpscript:run` reachable under their upstream names.
- [X] The fork adds no splint finding of its own. `bash phpx/scripts/lint-delta.sh` reports no rule grew over `upstream/main`.
- [X] A follow from upstream lands without a hand on it. `PHPX_FORCE=1 atkins -f atkins.phpx.yml phpx:follow` takes upstream's tree, runs `migrations/`, and commits. The integration runs one way and no file has two authors. The run reports no conflict.
- [X] The image builds and runs under both names. `atkins -f atkins.phpx.yml phpx:image`, then `docker run --rm titpetric/phpx:latest version` prints `github.com/titpetric/phpx`, and `docker run --rm --entrypoint phpscript titpetric/phpx:latest version` prints the same through the linked name.
- [X] The image publishes to `titpetric/phpx`. `docker push titpetric/phpx:latest`, then `docker rmi` it and `docker pull` it back: the pulled image answers `version` with the git stamp under both names. The semver tag builds when HEAD is that tag and at no other time, because this checkout inherited phpscript's tags.
- [X] The fork's main branch carries the generated tree. `git log origin/main -1` names the commit, and `git merge-base --is-ancestor origin/main main` holds on every push.
- [X] `atkins -f atkins.phpx.yml` passes end to end. Eleven legs: the follow, the three lint gates, the go suite under the tag, the binary and the image, the database pair, the four fixture and lifecycle legs, and the closing fmt that leaves the tree clean.

## The five reversals

One folder per reversal under `phpx/testdata/`, documented in the README beside them. `phpx test ./phpx/testdata/...` runs the lot; each box names the folder that settles it on all three runtimes.

- [X] `global` binds a name in the global frame, and the `global` keyword reads it. `phpx test --matrix ./phpx/testdata/globals/...` passes the php column over nine php behaviours in one fixture, and `go test -tags phpx ./runner/ -run TestPhpxGlobal` covers the binding, the null a missing global arrives as, `unset`, and the local value the statement pushes out of reach. `go test ./runner/ -run TestPhpxGlobalBindsNothing` pins that an untagged build still answers upstream's no-op. The Flat stack column skips: the bytecode engine declines the program and the whole of it runs on the interpreter.
- [X] `extends` carries fields, methods, statics and constants to the child, `parent::` runs the parent's body with `$this` bound, `self::` in an inherited body resolves to the declaring class, an instance answers `instanceof` for every ancestor, and a catch clause naming an ancestor takes a thrown child. `phpx test --matrix ./phpx/testdata/inheritance/...`, plus `go test -tags phpx ./runner/ -run 'TestPhpxClass|TestNoInheritanceAtRuntime'`, which stays green: the parent's members flatten into the child's `model.Class` and no `Parent` field appears on it. The fixture's Flat stack column skips because the engine declines a program that declares `extends`.
- [X] A trait's members land in every class that uses one. `phpx test --matrix ./phpx/testdata/traits/...` passes the php column over the members a trait contributes, a class method overriding one, a nested trait, a class property composing with an identical trait property, and the fatals php raises for a method collision, an incompatible property and a missing trait. `go test -tags phpx ./runner/ -run TestPhpxClassAppliesTraits` covers the merge and `TestPhpxClassRejectsTraitCollisions` the three messages. The fixture's Flat stack column skips because the engine declines a program that declares or uses a trait. Limits: the adaptation block (`insteadof`, `as`) is a parse error, a trait declared in another included file is not seen, `__CLASS__` in a trait method answers the trait's name, and an abstract trait method the class leaves unimplemented is accepted until the abstract reversal lands.
- [X] `abstract class` refuses instantiation and leaves its abstract methods to the child. `phpx test --matrix ./phpx/testdata/abstract/...` passes the php column over a child that implements every method, an abstract class with no abstract method instantiated through a concrete child, and a trait's abstract method satisfied by the using class; six `error:` fixtures pin php's fatals: `Cannot instantiate abstract class A`, `Class B contains 1 abstract method and must therefore be declared abstract or implement the remaining methods (A::g)`, `Class C contains 1 abstract method ... (C::z)` for a class's own declaration, `Cannot call abstract method A::f()`, `Cannot make non abstract method P::f() abstract in class C`, and `Abstract function A::f() cannot be declared private`. `go test -tags phpx ./runner/ -run TestPhpxClass` covers the merge, the order the message lists methods in, and both refusals; the untagged build keeps upstream's empty body. The fixture's Flat stack column skips because the engine declines a program that declares an abstract class or method. Limits: a class's own abstract declaration is reported before a trait's, which is php's compile-time order but not a case the fixtures pin.
- [X] `stdClass` is the class a decoded JSON object names, at any depth, and a decoded instance is an ordinary property bag. `phpx test --matrix ./phpx/testdata/stdclass/...` passes the php column over the default shape, `$associative = false` and `null`, the array shape `true` asks for, a nested object, a list of objects, the document's order surviving a round trip, and the property bag's `instanceof`, `foreach`, `unset`, dynamic property and `(array)` cast. `go test -tags phpx ./stdlib/core/ -run TestPhpxJSONDecode` covers the same shape from the binding, and the untagged build keeps upstream's decoder. `stdlib/core/phpx_json.go` re-registers `json_decode` after `json.go`: the fork changes the PHP name alone, and the extension `JSON\Decoder` still answers arrays. Limits: invalid input raises where php answers null, `$depth` and `$flags` are ignored, and a script may still declare its own `stdClass` where php refuses the name (see the discoveries below).

## The frameworks

Each box is a snippet under `phpx/testdata/frameworks/<name>/` that exercises the framework's own entry path, with the blockers it hits recorded in `phpx/testdata/report.md`. A framework that cannot run yet ticks nothing and earns a report entry naming the first wall it meets.

- [ ] Slim router. `phpx test ./phpx/testdata/frameworks/slim/...`.
- [ ] Latte template engine. `phpx test ./phpx/testdata/frameworks/latte/...`.
- [ ] Filament. `phpx test ./phpx/testdata/frameworks/filament/...`. Filament is a Laravel package over Livewire, and the report separates the two.
- [ ] Laravel. `phpx test ./phpx/testdata/frameworks/laravel/...`.
- [ ] `phpx/testdata/report.md` lists every blocker and every standard library extension a snippet reached for. `test -s phpx/testdata/report.md`, and a reader finds one entry per framework.

## Upstream

Each box is a pull request against phpscript, labelled `phpx: ...`. A merged one shrinks this fork.

- [ ] A program transform hook in `runner`. `runner/runner_load.go` and `runner/precompiler.go` call `parser.ParseFile` themselves, and an overlay outside the module never sees an included file's AST. Validated by the merged PR and by `migrations/` losing the step the hook replaces.
- [ ] `docker/Dockerfile` reads the `NAME` build argument it already receives. Validated by `atkins build` passing on this checkout without `phpx/docker/Dockerfile`.
- [ ] `splint ./...` at zero again, or a documented baseline. The lint gate here measures a delta because upstream main reports 236 `godoc/outweighs` findings against the current splint. Validated by `lint-delta.sh` reading a zero baseline.

## Discovered, not implemented

These came up while settling the five reversals and none is one of them. Each is a decision waiting for an issue; a box lands here with the command that shows the difference.

- `is_a()` is not registered, and `is_subclass_of()` is. `php -r 'var_dump(is_a(new stdClass, "stdClass"));'` answers true, and `phpx` raises `call to undefined function is_a()`. Framework code guards an object with `is_a($o, "stdClass")` before reading it, so a port meets this on the same path as the decoded body. A standard library extension, so it waits on the issue the scope section requires.
- `$o->{'name'}` is a parse error (`expected "(" after a braced member name`), and PHP allows a braced expression for a property name, which is how a decoded key with punctuation (`"a-b"`) or a computed name is read. `php -r '$o = new stdClass; $o->{"a-b"} = 1; echo $o->{"a-b"};'` prints `1`; `phpx` refuses the file. General OOP syntax, not stdClass-only.
- A second declaration of one class name is accepted, and so is a class, interface or trait named `stdClass`; PHP fatals `Cannot redeclare class A` and `Cannot redeclare class stdClass`. `php -r 'class A {} class A {}'` fails; `phpx` builds the last one. The fork's hoist already indexes declarations by name, so a refusal is a small change at that seam; it waits because upstream deliberately lets a script shadow `stdClass` (`model/stdClass.go`), and reversing that is its own decision.

## Outside the scope

These stay as phpscript decided them, and a box that needs one of them is a box that needs an issue first.

- No extension to the standard library. The snippets record what they wanted; `phpx/testdata/report.md` is where that goes.
- The "Won't build" table in `docs/design.md`: `eval`, `goto`, generators, `curl_*`, PDO, `&` outside `foreach`, and the rest.
- One type per variable name stays a lint convention. The runtime accepts a reassignment and the linter reports it as advice.
