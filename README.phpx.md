# phpx

The [phpx](https://github.com/phpscript/phpx) project is an automated fork of [phpscript](https://github.com/titpetric/phpscript). To hold fork fidelity, the decision is that phpx imports phpscript where it can.

The intent of the fork is to reverse the decisions that make phpscript incompatible with code written for PHP. The decisions this fork reverses:

- phpx implements global values, `global` keyword to use
- phpx implements inheritance, `extends`
- phpx implements traits
- phpx implements `abstract class`
- phpx implements `stdClass`

This is the complete goal of the fork. It aims to increase compatibility with code written for PHP. For this purpose, the agent should research more compatibility gaps against PHP, but keep to OOP issues.

Agent notes:

- The git remote is `git@github.com:phpscript/phpx.git`. The first draft of this file linked `titpetric/phpx`. One of the two is wrong and the repository owner decides which.
- `model.ClassDecl` carries `Parent` for `extends` and a `Abstract` flag, and the parser fills both (`model/ast.go:181`). Two of the five reversals need no parser change.
- `model.Class`, the runtime form, drops them both. `runner.TestNoInheritanceAtRuntime` (`runner/runner_control_test.go:128`) fails the build when `model.Class` grows a field named `Parent`, `Extends`, `Parents`, `Super` or `Base`. Flattening the parent's fields, methods, statics and constants into the child's `model.Class` leaves that test passing.
- `rt.hoist` (`runner/runner.go:150`) turns a `ClassDecl` into a `Class`. One seam there carries `extends`, traits and `abstract`.
- `global` parses to `model.Global` (`parser/parser.go:500`). `runner/runner.go:358` is a documented no-op and `lint/walk.go:81` reports the statement. The overlay binds the names and the AST needs no change.
- `stdClass` is `model.StdClass()` and `model.NewStdClass()` (`model/stdclass.go`). Measure what a framework misses before writing a replacement.
- Traits do not parse at all (`parser/parser.go:9` lists them among the omissions). Of the five, traits alone need parser work.

## The fork

The solution is a fork of phpscript. As such:

1. the fork pulls new changes in, `git pull --rebase` on the fork,
2. prefer an addition over a fix, and treat an edit to an existing file as the last option

Agent notes:

- `atkins -f atkins.phpx.yml phpx:pull` is that pull. `phpx/scripts/pull.sh` holds the logic: an `upstream` remote when the checkout has none, the generated commit dropped, the rebase, `migrations/run.sh`, and a commit of what the steps wrote.
- The fork carries two kinds of commit. An addition under a path upstream never writes replays through a rebase untouched. A rewrite of upstream's own files comes out of `migrations/`, and the next pull drops that commit and writes it again. [migrations/README.md](migrations/README.md) argues the split.
- `README.md` in the fork is the copy of this file that `migrations/002-readme.sh` writes. A pull throws away whatever upstream wrote there, and the one file both projects edit stops producing a conflict. `lefthook.yml` refuses a commit that stages `README.md` by hand and a push where the copy drifted, and the pipeline passes `PHPX_README=1` to write it.
- `CHANGELOG.md` is the fork's scratchpad. Every pull records the upstream range it landed on, and an agent adds what the overlay had to adjust under that entry.
- A rebase conflict stops the script. An addition does not conflict. A conflict means an upstream commit reached a path the fork owns, and that resolution is a decision for an agent.

## The module path

`migrations/001-module-path.sh` renames the module to `github.com/phpscript/phpx` and rewrites every import of it. `PHPX_MODULE` names a different target.

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
- This file answers to the same gates as every other document here. `atkins lint` runs `scripts/vale-gate.sh` over 39 rules at error level against a baseline of zero, and an edit here runs it.
