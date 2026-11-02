# phpx

The [phpx](https://github.com/titpetric/phpx) project is an automated fork of [phpscript](https://github.com/titpetric/phpscript). To hold fork fidelity, phpx keeps no hand-maintained diff of upstream's files: a follow takes phpscript's tree verbatim and `migrations/` runs again over it, renaming the module path and writing the files the fork owns.

The intent of the fork is to reverse the decisions that make phpscript incompatible with code written for PHP. The decisions this fork reverses:

- phpx implements global values, `global` keyword to use
- phpx implements inheritance, `extends`
- phpx implements traits
- phpx implements `abstract class`
- phpx implements `stdClass`

This is the complete goal of the fork. It aims to increase compatibility with code written for PHP. For this purpose, the agent should research more compatibility gaps against PHP, but keep to OOP issues.

Agent notes:

- [ACCEPTANCE.md](ACCEPTANCE.md) holds the command that settles each claim below. This file holds the status; that file holds the proof.
- `origin` is `git@github.com:titpetric/phpx.git` and `upstream` is `titpetric/phpscript`. `phpx/scripts/follow.sh` adds the second remote when a checkout has none.
- `model.ClassDecl` carries `Parent` for `extends` and an `Abstract` flag, and the parser fills both (`model/ast.go:181`). Of the five reversals, only traits needed parser work.
- `model.Class`, the runtime form, drops `Parent`; `Abstract` survives as a flag the runtime enforces. `runner.TestNoInheritanceAtRuntime` (`runner/runner_control_test.go:128`) fails the build when `model.Class` grows a field named `Parent`, `Extends`, `Parents`, `Super` or `Base`. Flattening the parent's fields, methods, statics and constants into the child's `model.Class` leaves that test passing.
- `rt.hoist` (`runner/runner.go:150`) turns a `ClassDecl` into a `Class`. One seam there carries `extends`, traits and `abstract`.
- `extends` works. `runner/phpx_inheritance.go` flattens the parent's members at hoist, keeps the chain in `model.Class.Ancestors` for `instanceof`, `parent::` and a catch clause, and records the declaring class in `model.Class.MethodOwner` so `self::` inside an inherited body resolves to the class that wrote it. `parent::__construct()` chains a constructor. The engine declines a program that declares `extends`, so the whole of it runs on the interpreter. Two limits stay: a parent that is a host class (`extends Exception`) contributes its name and no member, so `parent::__construct()` on one fails, and a static property keeps one bag per class rather than php's shared storage.
- `global` works. `runner/phpx.go` binds each name to the global frame through `Scope.bindStatic`, the mechanism `static $x` already used, and `phpx/testdata/patches/` carries the dispatcher case, the bytecode engine's refusal, the linter rule it drops and the two upstream fixtures that pinned the no-op.
- `stdClass` is `model.StdClass()` and `model.NewStdClass()` (`model/stdclass.go`). Measure what a framework misses before writing a replacement.
- traits work. The parser reads `trait Name { ... }` into `model.TraitDecl` and the `use A, B;` line inside a class or trait body into `model.ClassDecl.Traits`, the formatter prints both back, and `runner/phpx_traits.go` folds a trait's properties, statics, constants and methods into every class that names it. A class's own method wins over the trait's; a class's own property or constant is an ordinary definition that must agree with the trait's, as php checks it; two traits declaring one name compose only when their definitions agree, and php's fatal with php's message otherwise. A trait composes another trait through its own `use` line. A trait contributes no ancestor, so `instanceof TraitName` is false. The engine declines a program that declares or uses a trait. Three limits stay: the adaptation block (`A::f insteadof B;`, `A::f as g;`) is a parse error, a trait declared in one included file is not seen by a class in another, and `__CLASS__` inside a trait method compiles to the trait's name where php answers the using class.
- `abstract` works. `runner/phpx_abstract.go` refuses `new` on a class that declared the modifier and a call to a method with no body, and hoist refuses a concrete class that leaves a method to a child, with php's message and php's order: the class's own declarations, then what the parent left, then what a trait contributed. A child that implements every method instantiates; an abstract class with no abstract method instantiates through a concrete child; a trait's abstract method is satisfied by the class that uses the trait or by a concrete method it inherits. A class may not turn a method its parent implements back into a declaration without a body, and a class's own abstract declaration is reported before traits bind, which is php's compile-time order; a private method without a body is refused, while a trait may declare one because the using class supplies the body. The parser already recorded both spellings and the formatter printed them back, so no overlay syntax half was needed. The engine declines a program that declares an abstract class or an abstract method, so the whole of it runs on the interpreter.
- Of the five reversals, traits alone needed parser work, so its syntax half travels in the overlay: `parser/parser.go`, `model/ast.go`, `formatter/print_decl.go`, `formatter/print_stmt.go`, `formatter/printer.go` and `lint/names.go` sit under `phpx/testdata/patches/`. The untagged build parses a trait and takes nothing from it, the same one-direction rule `extends` follows.

## The components

Every part the fork adds, with a box that ticks when the part works. A worker ticks a box in the commit that earns it, and adds a line for a component this list does not name yet. A box stays open while the component exists and fails.

The harness a managed fork needs before a reversal can land:

- [X] `README.phpx.md`, the standing prompt and this list
- [X] `ACCEPTANCE.md`, one box per claim and the command that settles it
- [X] `CHANGELOG.md`, the scratchpad, one entry per follow
- [X] `atkins.phpx.yml`, a counterpart for all 24 jobs in `atkins.yml`
- [X] `migrations/run.sh`, every step in order
- [X] `migrations/001-module-path.sh`, the module rename and the import pass
- [X] `migrations/002-readme.sh`, `README.md` from `README.phpx.md`
- [X] `migrations/003-overlay.sh`, `phpx/testdata/patches/` copied over the tree
- [X] `phpx/scripts/overlay.sh`, the read-back and the report behind that copy
- [X] `phpx/scripts/splint-fix.sh`, `splint fix` with upstream's markdown kept
- [X] `migrations/README.md`, the contract a step holds to
- [X] `phpx/scripts/follow.sh`, upstream's tree verbatim, the steps over it, one commit
- [X] `phpx/scripts/lint-delta.sh`, splint measured against upstream
- [X] `lefthook.yml`, the `README.md` guard on commit and on push
- [X] `phpx/docker/Dockerfile`, the image. Docker Hub serves `titpetric/phpx:latest`, which carries the git stamp and answers `version` as `phpx` and as `phpscript`.
- [X] `atkins -f atkins.phpx.yml` green from end to end, eleven legs including the 294-fixture matrix on all three runtimes

The overlay a reversal lands in:

- [X] `runner/phpx.go`, `//go:build phpx`, the overlay's answer in the package that needs it
- [X] `runner/phpx_stub.go`, `//go:build !phpx`, upstream's answer, which keeps an untagged build reviewable
- [X] `phpx/testdata/patches/`, the fork's version of an upstream file
- [X] `phpx/testdata/`, one folder per reversal, documented in its own README.md
- [ ] `stdlib/phpx/`, for a reversal that needs a binding
- [ ] `phpx/testdata/report.md`, the blockers and the standard library each snippet reached for

The five reversals, each a folder under `phpx/testdata/` and a fixture that php settles first:

- [X] `global` values and the `global` keyword that reads them
- [X] inheritance, `extends`
- [X] traits, `trait` and `use`
- [X] `abstract class`
- [ ] `stdClass`

The frameworks, one snippet tree each under `phpx/testdata/frameworks/`:

- [ ] Slim router
- [ ] Latte template engine
- [ ] Filament
- [ ] Laravel

What the follow already holds, with no box left to tick:

- [X] the module path reads `github.com/titpetric/phpx`, over 446 files
- [X] `README.md` is the generated copy, and a hand edit cannot land
- [X] `origin` is the fork and `upstream` is phpscript
- [X] a follow takes upstream's tree, runs the steps and commits without a hand on it

## The fork

The solution is a fork of phpscript. As such:

1. the fork follows new changes in one direction, phpscript to phpx, and never the other way,
2. prefer an addition over a fix, and treat an edit to an existing file as the last option

Agent notes:

- `atkins -f atkins.phpx.yml` is the whole gate. Every job in `atkins.yml` has a counterpart there, the binary is `bin/phpx` with `bin/phpscript` linked beside it for the shebang fixture, and `default` ends on `fmt` as upstream's does: `test:introspection` regenerates two tracked documents, the generator wraps its prose where mdox does not, and a run without that leg ends on a diff. `fmt` runs `splint fix` through `phpx/scripts/splint-fix.sh`, because splint rewrites markdown as well as Go and the fork keeps upstream's prose. `gen` stays out, and a follow refuses a dirty tree.
- `phpx/scripts/lint-delta.sh` stands in for the bare `splint ./...`. Upstream main reports 236 `godoc/outweighs` findings against the splint on this machine, and the fork's rule is to add none of its own. No checkout of this tree reaches the zero upstream claims.
- `atkins -f atkins.phpx.yml phpx:follow` is that follow, and `phpx/scripts/follow.sh` holds the logic: an `upstream` remote when the checkout has none, the removal of every path upstream stopped carrying, upstream's tree checked out verbatim, `migrations/run.sh` over the result, and one commit.
- The follow neither pulls nor merges. A pull asks git to reconcile two edits of one file, and the fork's edit of an upstream file is a migration over the whole tree: git would report a conflict on every file the rename touched, each one resolved by running the step again. The step runs instead, and no file has two authors.
- Three things come out of that. The history is append-only, and every push is a fast-forward. A conflict cannot happen. An edit the fork wants inside an upstream file belongs in a migration step, because upstream's version of that file lands verbatim and a committed edit to it disappears.
- `follow.sh` carries the list of files the fork owns, and the policy above reads back as code there: `README.phpx.md`, `ACCEPTANCE.md`, `CHANGELOG.md`, `atkins.phpx.yml`, `lefthook.yml`, everything under `phpx/`, `migrations/` and `stdlib/phpx/`, and any file named `phpx.go` or `phpx_*.go` in any package. A new fork-owned path goes on that list in the commit that creates it. [migrations/README.md](migrations/README.md) carries the contract a step holds to.
- `README.md` in the fork is the copy of this file that `migrations/002-readme.sh` writes. A follow takes upstream's version and the step writes over it, and the one file both projects edit produces no conflict. `lefthook.yml` refuses a commit that stages `README.md` by hand and a push where the copy drifted, and the pipeline passes `PHPX_README=1` to write it.
- `CHANGELOG.md` is the fork's scratchpad. Every follow records the upstream commit it landed on, and an agent adds what the overlay had to adjust under that entry.
- Migrations are for a change that reaches the whole tree. The overlay itself is hand-written code in a fork-owned file, and a seam of one line inside an upstream file is a step, because the follow overwrites that file.
- atkins opens its run database at `.atkins/atkins.db`, beside the tracked skills it reads, and SQLite leaves `-shm` and `-wal` files there while a run holds one. `migrations/004-atkins-state.sh` ignores the three in `.gitignore`, because a pipeline run otherwise leaves a tree `follow.sh` refuses as dirty.

## Local changes over upstream files

`phpx/testdata/patches/` mirrors the source tree: `phpx/testdata/patches/runner/runner.go` is the fork's `runner/runner.go`. `migrations/003-overlay.sh` copies each one over the tree after a follow has landed upstream's version, and that is how a local change to an upstream file survives.

Agent notes:

- The overlay carries a whole file. A diff fails on context drift and leaves the tree half-migrated, and a copy always applies. The cost is that an upstream edit to an overlaid file goes unseen, and `atkins -f atkins.phpx.yml phpx:overlay:report` is what reads it back: it diffs each overlay against `upstream/main` and names the lines the fork drops. Read it after a follow.
- The overlay sits under `testdata` because the go tool ignores every path with a `testdata` element. Without that, `go build ./...` compiles the copies as packages of their own and splint reads them as source: the first run of this mechanism reported a 1218-line file and three missing test files, every one of them a copy. `test:phpx:fixtures` skips the folder for the same reason, since the fixture runner would otherwise run an overlaid fixture twice.
- `phpx/testdata/patches/REMOVED` lists the upstream paths the fork drops. The removal runs before the copy, and a path there cannot also be overlaid.
- An overlay with no file under it in the tree is an addition, and that is the fork's route to a new file inside an upstream directory. `follow.sh` drops a tracked path upstream does not carry, and the step writes it again.
- `phpx:overlay:capture` reads every overlaid file back out of the tree, and `default` ends on it. A job in the run edits an overlaid file, gofmt regrouping an import or mdox rewrapping a paragraph, and without the read-back the overlay goes stale and the next follow installs the old version over the fixed one.
- Prefer a fork-owned file to an overlay. A new file named `phpx.go` or `phpx_*.go` in the package that needs it survives a follow by name and carries no review burden. The overlay is for the line that has to sit inside an upstream file: a `case` in a dispatcher, an expected output in a fixture, a rule the linter stops reporting.

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
