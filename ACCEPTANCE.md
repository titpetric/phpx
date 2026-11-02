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

- [ ] `global` binds a name in the global frame, and the `global` keyword reads it. `phpx test --matrix ./phpx/testdata/globals/...`. The statement parses to `model.Global` today and `runner/runner.go` answers it with a no-op. The fixture covers a write in one function and a read in another.
- [ ] `extends` carries fields, methods, statics and constants to the child. `phpx test --matrix ./phpx/testdata/inheritance/...`, plus `go test -tags phpx ./runner/ -run TestNoInheritanceAtRuntime`, which stays green: the parent's members flatten into the child's `model.Class` and no `Parent` field appears on it.
- [ ] A trait's members land in every class that uses one. `phpx test --matrix ./phpx/testdata/traits/...`. Traits reach no parser today. This box needs `parser` work and a fixture for the conflict rules php applies when two traits declare one name.
- [ ] `abstract class` refuses instantiation and leaves its abstract methods to the child. `phpx test --matrix ./phpx/testdata/abstract/...`, with a fixture for the fatal error php raises on `new` of an abstract class.
- [ ] `stdClass` takes a dynamic property, from `new stdClass` and from a cast. `phpx test --matrix ./phpx/testdata/stdclass/...`. `model.StdClass()` exists, and the fixture measures what a framework reaches for before any code changes.

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

## Outside the scope

These stay as phpscript decided them, and a box that needs one of them is a box that needs an issue first.

- No extension to the standard library. The snippets record what they wanted; `phpx/testdata/report.md` is where that goes.
- The "Won't build" table in `docs/design.md`: `eval`, `goto`, generators, `curl_*`, PDO, `&` outside `foreach`, and the rest.
- One type per variable name stays a lint convention. The runtime accepts a reassignment and the linter reports it as advice.
