# phpx

The [phpx](https://github.com/titpetric/phpx) project is an automated fork of [phpscript](https://github.com/titpetric/phpscript). To hold fork fidelity, phpx keeps no hand-maintained diff: a pull rebases this tree onto upstream phpscript, and `migrations/` runs again to rename the module path and write the files the fork owns.

The intent of the fork is to reverse the decisions that make phpscript incompatible with code written for PHP. The decisions this fork reverses:

- phpx implements global values, `global` keyword to use
- phpx implements inheritance, `extends`
- phpx implements traits
- phpx implements `abstract class`
- phpx implements `stdClass`

This is the complete goal of the fork. It aims to increase compatibility with code written for PHP. For this purpose, the agent should research more compatibility gaps against PHP, but keep to OOP issues.

Agent notes:

- [ACCEPTANCE.md](ACCEPTANCE.md) holds the command that settles each claim below. This file holds the status; that file holds the proof.
- `origin` is `git@github.com:titpetric/phpx.git` and `upstream` is `titpetric/phpscript`. `phpx/scripts/pull.sh` adds the second remote when a checkout has none.
- `model.ClassDecl` carries `Parent` for `extends` and a `Abstract` flag, and the parser fills both (`model/ast.go:181`). Two of the five reversals need no parser change.
- `model.Class`, the runtime form, drops them both. `runner.TestNoInheritanceAtRuntime` (`runner/runner_control_test.go:128`) fails the build when `model.Class` grows a field named `Parent`, `Extends`, `Parents`, `Super` or `Base`. Flattening the parent's fields, methods, statics and constants into the child's `model.Class` leaves that test passing.
- `rt.hoist` (`runner/runner.go:150`) turns a `ClassDecl` into a `Class`. One seam there carries `extends`, traits and `abstract`.
- `global` parses to `model.Global` (`parser/parser.go:500`). `runner/runner.go:358` is a documented no-op and `lint/walk.go:81` reports the statement. The overlay binds the names and the AST needs no change.
- `stdClass` is `model.StdClass()` and `model.NewStdClass()` (`model/stdclass.go`). Measure what a framework misses before writing a replacement.
- Traits do not parse at all (`parser/parser.go:9` lists them among the omissions). Of the five, traits alone need parser work.

## The components

Every part the fork adds, with a box that ticks when the part works. A worker ticks a box in the commit that earns it, and adds a line for a component this list does not name yet. A box stays open while the component exists and fails.

The harness a managed fork needs before a reversal can land:

- [X] `README.phpx.md`, the standing prompt and this list
- [X] `ACCEPTANCE.md`, one box per claim and the command that settles it
- [X] `CHANGELOG.md`, the scratchpad, one entry per pull
- [X] `atkins.phpx.yml`, a counterpart for all 24 jobs in `atkins.yml`
- [X] `migrations/run.sh`, every step in order
- [X] `migrations/001-module-path.sh`, the module rename and the import pass
- [X] `migrations/002-readme.sh`, `README.md` from `README.phpx.md`
- [X] `migrations/README.md`, the contract a step holds to
- [X] `phpx/scripts/pull.sh`, the rebase, the steps, and the two commits it writes
- [X] `phpx/scripts/lint-delta.sh`, splint measured against upstream
- [X] `lefthook.yml`, the `README.md` guard on commit and on push
- [ ] `phpx/docker/Dockerfile`, the `titpetric/phpx` image. The file builds and the published tag is still missing.
- [ ] `atkins -f atkins.phpx.yml` green from end to end. Nine legs of ten pass; the image leg failed on a binary linked against glibc and the pipeline now clears `CGO_ENABLED`.

The overlay a reversal lands in:

- [ ] `runner/phpx.go`, the hook variable the stock path calls
- [ ] `runner/phpx_overlay.go`, `//go:build phpx`, the implementation it assigns
- [ ] `phpx/`, the package the overlay keeps its own code in
- [ ] `stdlib/phpx/`, for a reversal that needs a binding
- [ ] `phpx/testdata/`, one folder per reversal, documented in its own README.md
- [ ] `phpx/testdata/report.md`, the blockers and the standard library each snippet reached for

The five reversals, each a folder under `phpx/testdata/` and a fixture that php settles first:

- [ ] `global` values and the `global` keyword that reads them
- [ ] inheritance, `extends`
- [ ] traits
- [ ] `abstract class`
- [ ] `stdClass`

The frameworks, one snippet tree each under `phpx/testdata/frameworks/`:

- [ ] Slim router
- [ ] Latte template engine
- [ ] Filament
- [ ] Laravel

What the pull already holds, with no box left to tick:

- [X] the module path reads `github.com/titpetric/phpx`, over 446 files
- [X] `README.md` is the generated copy, and a hand edit cannot land
- [X] `origin` is the fork and `upstream` is phpscript
- [X] a pull rebases, runs the steps and commits without a hand on it

## The fork

The solution is a fork of phpscript. As such:

1. the fork pulls new changes in, `git pull --rebase` on the fork,
2. prefer an addition over a fix, and treat an edit to an existing file as the last option

Agent notes:

- `atkins -f atkins.phpx.yml` is the whole gate. Every job in `atkins.yml` has a counterpart there, the binary is `bin/phpx` with `bin/phpscript` linked beside it for the shebang fixture, and no job in `default` writes a tracked file, because the next pull refuses a dirty tree. `fmt` and `gen` are jobs a person calls.
- `phpx/scripts/lint-delta.sh` stands in for the bare `splint ./...`. Upstream main reports 236 `godoc/outweighs` findings against the splint on this machine, and the fork's rule is to add none of its own. No checkout of this tree reaches the zero upstream claims.
- `atkins -f atkins.phpx.yml phpx:pull` is that pull. `phpx/scripts/pull.sh` holds the logic: an `upstream` remote when the checkout has none, the generated commit dropped, the rebase, `migrations/run.sh`, and a commit of what the steps wrote.
- The fork carries two kinds of commit. An addition under a path upstream never writes replays through a rebase untouched. A rewrite of upstream's own files comes out of `migrations/`, and the next pull drops that commit and writes it again. [migrations/README.md](migrations/README.md) argues the split.
- `README.md` in the fork is the copy of this file that `migrations/002-readme.sh` writes. A pull throws away whatever upstream wrote there, and the one file both projects edit stops producing a conflict. `lefthook.yml` refuses a commit that stages `README.md` by hand and a push where the copy drifted, and the pipeline passes `PHPX_README=1` to write it.
- `CHANGELOG.md` is the fork's scratchpad. Every pull records the upstream range it landed on, and an agent adds what the overlay had to adjust under that entry.
- A rebase conflict stops the script. An addition does not conflict. A conflict means an upstream commit reached a path the fork owns, and that resolution is a decision for an agent.

## The module path

`migrations/001-module-path.sh` renames the module to `github.com/titpetric/phpx` and rewrites every import of it. `PHPX_MODULE` names a different target.

Agent notes:

- `go mod edit -module` is the whole of what the go toolchain carries for this. It writes the go.mod module line and leaves every import of the old path alone, and the go documentation names a text rewrite over the sources as the other half. `gomvpkg` and `gorename` under `golang.org/x/tools` move a package and rename an identifier, predate modules, and sit on no path here.
- The step rewrites one string, `github.com/titpetric/phpscript`. The word phpscript stays in the help text, in the fixtures and in every document. Rewriting those would edit the behaviour the fork inherits.
- `splint fix ./...` runs after the rewrite. The new path sorts into a different import group and the lint gate reads the grouping.
- After the rename, `go install .` writes a binary called `phpx`. A binary takes the last element of the module path.

## Expectations for the additions

- `phpx/testdata/{globals,inheritance,traits,...}` and assisting subfolders with code
- extend with the `/phpx` package, or prefer file additions in existing scopes for go code
- smallest change for the integration, enforce with `-tags phpx` in build steps
- can git clone phpscript into a gitignored subfolder or a go workspace

The agent should choose the methods of implementation based on what causes the least friction from phpscript imports. The phpscript project may adjust structure or add hooks that can make this integration smaller. Any change outside of `/phpx` and `/stdlib/phpx` has to carry phpx attribution, the preferred way is to avoid changes in other places, but if required then keep the changes to `phpx.go` in the relevant packages. It may be hard to integrate phpx. The result embeds phpscript and produces an overlay that implements these runtime compatibilities with the official PHP runtime.

Agent notes:

- A go build tag applies to a file in the module the build compiles. `-tags phpx` and an import of phpscript cannot both hold for one file. The choice is per package: a tagged file in this tree, or an exported hook upstream.
- An imported `runner` cannot see the AST of an included file. `runner/runner_load.go:47` and `runner/precompiler.go:124` call `parser.ParseFile` themselves, and a program the overlay never sees runs phpscript's rules. A transform hook upstream is the smallest change that enables it, and the first `phpx: ...` PR to write.
- The shape that holds each seam to one line: `runner/phpx.go` declares a hook variable the stock path calls, and `runner/phpx_overlay.go` carries `//go:build phpx` and assigns the implementation. An untagged build answers as phpscript does.

## Outputs

The intended outputs of the workflow are:

- A published docker image on `titpetric/phpx`
- Commits and pushes to the main branch of `phpx`

Agent notes:

- `phpx:image` builds `phpx/docker/Dockerfile`, which links `phpscript` to the `phpx` binary inside the image. A compose file or a script written against phpscript runs there unchanged.
- `docker/Dockerfile` upstream reads `./bin/phpscript-linux-amd64` while `go:build` writes `bin/phpx-linux-amd64`, which fails upstream's `atkins build` in this checkout. That file takes a `NAME` build argument and ignores it. Reading that argument is a one-file fix and a second `phpx: ...` PR.
- atkins reads the image name off the checkout directory (`$HOME/.atkins/skills/docker.yml`), and `titpetric/phpx` resolves without configuration.
- `phpx:push` and `phpx:image` sit outside the default pipeline. A run that publishes is a run somebody asked to publish.

## The testdata tree

As the fork gets updated with repeated runs of this workflow, phpx should create a `phpx/testdata` tree, documented in its own README.md. On creation, it should perform research over what php frameworks are common, with this starting list:

- Slim router
- Latte template engine
- Filament
- Laravel

The agent should write test snippets against all the frameworks, and note any blockers or incompatibilities that remain. Anything the reverted decisions leave unstated, but that forms a foundational need of the runtime, stays the same. Adding extensions to the standard library is not the intention here, and `phpx/testdata/report.md` records what those extensions would be.

Agent notes:

- A fixture is a `.phpt` document, or a `<name>_test.php` body with its `<name>_test.txt` output beside it. The expected output comes from running the source through `php` first. A fixture's own folder is its include root.
- `phpscript test ./...` runs a tree. `phpscript test .` matches the fixtures sitting directly in one directory and passes over every folder below it.
- Filament is a Laravel package. Its blockers are Laravel's blockers plus a Livewire front end, and the report should separate the two.

## Contributing upstream

The phpscript contribution policy allows you to open PRs against phpscript, with the label of `phpx: ...`. The ideal solution would have phpx integrate against the phpscript `runner` package and wrap the `phpscript` cmd toolchain. The output binary of `phpscript` and `phpx` have to work the same way in use, hence the need for a maintained fork. This prompt lives in `README.phpx.md` to avoid git conflicts in the fork.

Agent notes:

- The first draft of this file named a `runtime/` package. The package is `runner/`, and the toolchain is `cmd/phpscript/`.
- This file answers to the same prose gate as every other document here. `scripts/vale-gate.sh` reads 39 rules at error level against a baseline of zero, and an edit to this file runs `atkins -f atkins.phpx.yml lint`.
