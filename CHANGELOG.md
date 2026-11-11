# Changelog

## 2026-11-12 - follow upstream fc26e38, relaticle reaches a response

- follow to upstream `fc26e38` (4 commits: prose rules and generated docs, one test fix); `follow.sh` clean, 83 overlays copied, `go test -tags phpx ./...` clean after `test:introspection` regenerated `godoc.json` and two reference pages
- an untyped child redeclaration with no default (`protected $aliases;`) no longer nulls the ancestor's private array default (`phpxMergeFields`). This is an approximation of php's two slots: child code reads `[]` where php reads null. Laravel's console `Command` boots through it
- `new readonly class(...) {}` parses (`parser/expr.go`, overlaid) and sets `ClassDecl.Readonly`
- a call `config()` in a file with `use Pkg\Config;` named the class: `qualify` retries an alias case-insensitively, right for a class and wrong for a function. `parser/phpx_qualify.go` `qualifyCall` reads function aliases in exact spelling; the call sites in `parser/expr.go` (overlaid) use it
- apps (`boot.sh`): relaticle reaches stage 9 (HTTP 301, 519 of 3,431 `app/` statements) where php 8.4 itself fatals on a trait property composition; erpsaas 500 vs 302; urlhub and project-management stop at PDO; akaunting at `stream_filter_register`
- matrix `tests/fixtures/...` plus `phpx/testdata/classes/...` 376 of 376 on all three columns; not run: the atkins pipeline, `lint-delta`
- next: erpsaas's 500, then PDO (`PDO::FETCH_*` and a sqlite driver), then two real slots for a shadowed private property

## 2026-11-11 - upstream fc26e38

- follow `fc26e38`, 54 commits ahead of the last one
- `migrations/run.sh` wrote the generated tree again, 1 files dropped

## 2026-11-11 - spread of a Go collection, relaticle's real wall named

- no upstream movement: `upstream/main` is still `747a625`
- `[...$x]` and `f(...$x)` accept the Go slice or string-keyed map a binding returns (`runner/expr/phpx_spread_go.go`, `goPairs`; a map spreads in key order, since Go keeps none). Unit test `TestGoPairs`
- relaticle's `Only arrays and Traversables can be unpacked` was not a Traversable: it is `...$this->aliases` in `symfony/console` `Command::__construct` reading `null`. `Illuminate\Console\Command` redeclares `protected $aliases;` (null) over the parent's `private array $aliases = []`, and the fork keeps one slot per property name, so the child's null replaces the parent's default. php keeps two. This is the "private property of an ancestor shares the slot" limit, and Laravel's every console command reaches it. Repro: `class P { private array $a = []; function f() { return [...$this->a]; } } class C extends P { protected $a; }`
- next: two slots for a shadowed private property (mangle the ancestor's key by declaring class in the read, write, unset and iteration seams of `runner/phpx_properties.go` and the overlaid `helpers.go`, `runner_assign.go`), then PDO, then erpsaas's 500

## 2026-11-10 - DatePeriod, date ordering, erpsaas reaches a response

- no upstream movement: `upstream/main` is still `747a625`
- `DatePeriod` is a PHP class in the `runner/phpx_prelude/spl.php` prelude (end date, recurrences, `EXCLUDE_START_DATE`, `INCLUDE_END_DATE`, `getStartDate`/`getEndDate`/`getDateInterval`/`getRecurrences`), registered in `phpx_prelude.go`. `phpx/testdata/classes/dateperiod_test.php` is settled against php. Carbon's `DatePeriodBase extends DatePeriod` now resolves
- date ordering was a bug: `$a < $b` on two `DateTime`/`DateTimeImmutable` compared the empty strings of two host values and answered false for `<`, `>` and true for `==`. `internal/phpval/compare.go` (overlay) asks a host value to order itself (`hostComparer`, `phpx_compare.go`); `DateTimeImmutable.PhpxCompare` compares the moments
- `PHP_MAXPATHLEN` is bound
- apps run (`boot.sh`, prod trees): erpsaas passed boot and reaches `GET /` with a 500 where php answers 302 (was: stops at `DatePeriod`); urlhub is next stopped by PDO (`PDO::FETCH_OBJ`), as project-management is; akaunting by `stream_filter_register`; relaticle by argument unpacking of a non-array Traversable (`Only arrays and Traversables can be unpacked`)
- matrix `tests/fixtures/...` 309 of 309; `phpx test ./phpx/testdata/classes` 64 of 64; not run: the atkins pipeline, `lint-delta`. Next: PDO (`PDO::FETCH_*` and a sqlite driver), then the spread of a Traversable, then erpsaas's 500

## 2026-11-09 - follow upstream 747a625, `model` splits into `model` and `modelutil`

- follow to upstream `747a625` (3 commits: `feof`, a recursive list sweep that leaves `testdata` out, and the vale pass that moved the helpers of `model` into a new `modelutil` package and added `model.Valid`)
- `follow.sh` stopped at the build: `model.ToArray`, `CopyValue`, `UnwrapParenthesized`, `ByRefArg`, `IsCollection`, `CheckInterfaceContracts`, the `Go*` helpers and `GoNamespace` now live in `modelutil`. 51 overlays merged clean with `git merge-file` (base `2432830`), nine took a hand: `parser/{expr,parser,parser_test}.go`, `lint/lint.go`, `runner/{runner,runner_assign}.go`, `runner/expr/direct.go`, `flatstack/engine/program_compiler.go` and `tests/bindings_test.go` kept the fork's side of each conflict
- `model/{copy,byref,copy_test}.go` overlays moved to `phpx/testdata/patches/modelutil/`. `model/phpx_interfaces.go` moved to `modelutil/phpx_interfaces.go`, because `contractOf` and `interfaceIndex` went with the package. Every fork-owned file that read a moved name was rewritten to `modelutil.`
- `go test -tags phpx ./...` clean after `test:introspection` regenerated `godoc.json`; matrix `tests/fixtures/...` 309 of 309 on flatstack, the interpreter and php; phpunit `compare.sh` same as php on the six files
- not run: the atkins pipeline, `lint-delta`

## 2026-11-08 - follow upstream 2432830, php://memory is upstream's

- follow to upstream `2432830` (3 commits: `php://memory` and `php://temp` as a `memoryStream` wrapper, modes, `stream_get_contents` on a pipe)
- `follow.sh` stopped at the build: `Wrapper` gained a `mode` argument and upstream registers `memory` and `temp` itself. `stdlib/ps/files/phpx_wrappers.go` keeps only `stdout` and `stderr` and drops its scratch-file memory and its `phpxStreamContents`; `rememberStream` also tracks `*memoryStream`, so `stream_get_meta_data` still answers for it
- `stdlib/ps/files/{stream,read,write,csv_test}.go` overlays rebuilt on upstream (`git merge-file`, base `6904f62`); `stream.go` takes upstream's `stream_get_contents` and re-adds `pathinfo`, `stream_get_meta_data`, the `data:` URI, the fopen warning. `file_put_contents("php://...")` is upstream's again
- `go test -tags phpx ./...` clean after `test:introspection` regenerated `godoc.json`; matrix `tests/fixtures/...` 308 of 308; phpunit `compare.sh` same as php
- not run: the atkins pipeline, `lint-delta`, the apps run. Next: `phpx/testdata/apps/run.sh` against the new tip, then `DatePeriod` and PDO

## 2026-11-07 - follow upstream 6904f62, the operator PR landed

- follow to upstream `6904f62` (14 commits: phpscript#211 merged as `model.Coalesce` and `<=>`/`and`/`or`/`xor` binary ops, Go-package imports (`import "net/http"`), division and modulo by zero throw, `compareObjects` and array-vs-bool comparison in `phpval.Compare`)
- `follow.sh` stopped at the build: eleven whole-file overlays were stale. Rebased with `git merge-file` (base `dce927b`): ten merged clean, `flatstack/engine/vm.go`, `stdlib/imports.go` and `parser/lexer.go` took upstream's with the fork's line re-added (`stdlib/px/phpx` import, `?->` token); `runner/runner_assign.go` was rebuilt from upstream's new `assignValue`/`assignIndex`/`assignProp` split with the fork's seams (ref assign, copy-on-read, hooks, magic set, property access, indirect write, string offset, ArrayAccess) re-applied; `runner/expr/direct.go` took upstream's `markCall` fallback for non-variable assignment and kept the slot sync
- retired as the changelog below promised: `parser/phpx_coalesce.go`, `phpx_spaceship.go` (+ tests), `stdlib/px/phpx/spaceship.go`, the `xor` registration, `phpxParseLogical`, and the `parser/expr.go` seams. `phpxTemps` moved to `phpx_nullsafe.go`, its remaining user besides destructuring
- `internal/phpval/compare.go` is now upstream's plus one seam: an instance with `__toString` compares as its string against a string (`tostring.phpt` pinned it; upstream returns 1)
- `go test -tags phpx ./...` clean; matrix `tests/fixtures/...` 307 of 307; `phpx/testdata/...` all pass; phpunit `compare.sh` identical to php on the six files
- not run: the atkins pipeline, `lint-delta`, the apps run. Next: the fork's `phpx/testdata/{functions,syntax}` coalesce/spaceship/word fixtures overlap upstream's three; they stay while they cover `fn`, `print` and `global`

## 2026-11-07 - upstream PR for `??`, `??=`, `<=>`, `and`, `or`, `xor`

- opened [phpscript#211](https://github.com/titpetric/phpscript/pull/211), `phpx: null coalescing, spaceship and word operators`: the fork's parser lowering (`parser/phpx_coalesce.go`, `phpx_spaceship.go`, `phpx_logical.go`) ported as nodes: `model.Coalesce`, and `<=>`, `and`, `or`, `xor` as `model.Binary` ops, with `runner/expr`, the bytecode engine (`opJumpNotNil`), the formatter and the linter reading them. Three fixtures settled against php sit in `tests/fixtures/syntax/` (`null_coalescing`, `spaceship_operator`, `word_operators`), taken from `phpx/testdata/functions/{null_coalesce,spaceship}_test.php` and `phpx/testdata/syntax/word_operators_test.php` minus `global`, `fn` and `print`, which upstream lacks
- review pass on the PR found two parse bugs (`??=` folded onto an operand, `$a ?? $b ??= 3`), a bytecode-engine cliff and a parse-cost regression; all four fixed on the branch, the rest argued in the PR thread
- not ported: `?->` (needs a chain-aware node, not a flag), `print` as an expression, `fn`. `??=` refuses a call in its target where the fork's lowering accepted it with a temporary
- nothing in phpx changes until the PR merges: the next follow brings the nodes, and that follow's commit deletes `parser/phpx_coalesce.go`, `phpx_spaceship.go`, `phpx_logical.go` (the `or`/`xor`/`and` part), `stdlib/px/phpx/spaceship.go` and the `parser/expr.go` seams in the `expr.go` overlay, then drops the duplicated fixtures. Do not delete them before: the untagged and tagged builds both need an answer

## 2026-11-07 - follow upstream dce927b, areas.go overlay rebuilt

- follow to upstream `dce927b` (2 commits: area-qualified class names, `f47c42e` successor); `follow.sh` ran clean (`93c8ee0`)
- the follow left `stdlib` red: `phpx/testdata/patches/stdlib/areas.go` is a whole-file overlay of the old upstream file and shadowed upstream's `ErrorsArea`, `SetAreaClasses` and `areaClasses`. The overlay is rebuilt from upstream's new file plus the one block the fork owns (the `NamespaceLate` area). `overlay.sh report` after a follow is what names this; it ran before the tests this time, not after
- `godoc.json` and the two generated reference documents regenerated with `test:introspection`
- `go test -tags phpx ./...` clean; matrix `tests/fixtures/...` 299 of 299; `phpx/testdata/...` 193 of 194 (the known `patches/` copy)
- `go get -u ./...` would move `pflag` to 1.1.0 only; `go.mod` is upstream's file and a follow takes it verbatim, so the bump is not committed here. It belongs upstream
- not run: the atkins pipeline, `lint-delta`, the apps run

## 2026-11-06 - upstream dce927b

- follow `dce927b`, 22 commits ahead of the last one
- `migrations/run.sh` wrote the generated tree again, 1 files dropped

## 2026-11-06 - follow upstream (21 commits), godoc index rename

- follow to upstream `de399bb`..`f47c42e`'s successor (21 commits, `chore(phpx): follow upstream` `28882bc`); `follow.sh` ran clean
- the follow left four packages red (`list`, `internal/apidoc`, `scripts/list-apis`, `cmd/phpscript/list`): `internal/apidoc/godoc.json` is keyed by Go import path and migration 001 rewrote `*.go` only, so every `stdlib/ps/...` key kept the phpscript module and the embedded index documented nothing. `001-module-path.sh` now rewrites `godoc.json` too; the index and the two generated reference documents were regenerated with `test:introspection`
- `go test -tags phpx ./...` clean; matrix `tests/fixtures/...` 298 of 298; `phpx/testdata/...` 193 of 194 (the known `patches/` copy). Not run: the atkins pipeline, `lint-delta`, the apps run

## 2026-11-06 - ReflectionProperty raw accessors

- `ReflectionProperty::getRawValue()` and `setRawValue()` (php 8.4) in `runner/phpx_prelude/reflection.php`: they read and write the backing slot and skip a property hook. `testdata/hooks/rawvalue_test.php`, php's output beside it. This was the wall urlhub stopped at; the app run was not repeated, so the next wall is not measured
- upstream still `f47c42e`; nothing to follow. `go test -tags phpx ./runner` and the hooks fixtures pass. Not run: the matrix, the pipeline, `go get ./...`

## 2026-11-06 - property hooks, the first wall of the real apps

- `public int $n { get => ...; set { ... } }` parses and runs (php 8.4). `parser/phpx_hooks.go` lowers each hook to a method, `__phpx_get_NAME` and `__phpx_set_NAME`, on the class; `runner/phpx_hooks.go` calls the set hook from `$obj->name = v` (`runner/runner_assign.go`) and the get hook from a read (`runner/helpers.go`). The hook's own access to the property reaches the backing value, through the guard `__get` already used. `set => expr` stores expr, `set(T $v)` names its parameter, `public private(set)` parses (the `(set)` is read, not enforced). `testdata/hooks/hooks_test.php`, php's output beside it
- effect: urlhub no longer stops in `symfony/http-foundation`; it stops at `ReflectionProperty::setRawValue()`, which `symfony/var-exporter` calls. relaticle class load 678 of 679 php loads (was 675)
- limits: a compound write (`$o->n += 1`) and `unset` skip the hook; a hook in an interface, `final` and attributes on a hook are not read; the formatter prints a hook back as two methods
- `go test ./parser ./runner` clean, untagged and with `-tags phpx`. Not run: the matrix, the atkins pipeline, `lint-delta`, `go get ./...`

## 2026-11-05 - five real apps, and four bugs they found

- `phpx/testdata/apps/` measures five open-source Laravel and Filament apps (relaticle, erpsaas, project-management, urlhub, akaunting) with five tests: parse, class load, boot, first request, and the statements of `app/` that ran. `run.sh` prints every table. `phpx/TESTING.md` is rewritten around them: a verdict, what a pass means, one table per project, the blockers ranked by apps stopped, and what the numbers mean to an integrator. The 2026-11-04 report is its appendix
- result: the apps parse at 99.8% to 100%, four of five load 98.7% of the classes php loads, none answers `GET /`. The walls: PHP 8.4 property hooks and `public(set)` in `symfony/http-foundation` (two apps), PDO, `DatePeriod`, `stream_filter_register`
- `\true`, `\false` and `\null` are the constants (`parser/phpx_fqconst.go`, hook in the `parser/expr.go` overlay); `testdata/syntax/qualified_constants.phpt`
- a trait reached by two paths composes once instead of raising a collision (`runner/phpx_traits.go`); `testdata/traits/shared_trait.phpt`. Livewire's `WithFileUploads` hit it in 22 classes
- `IteratorIterator` forwards an unknown method to its inner iterator (`runner/phpx_prelude/spl.php`); `testdata/classes/iterator_iterator_forwards.phpt`. symfony/finder's `ExcludeDirectoryFilterIterator` hit it
- a call to an inaccessible method goes to `__call` / `__callStatic` (`runner/helpers.go` and `runner/static.go` overlays); `testdata/classes/magic_call_inaccessible.phpt`. Laravel's `Route::prefix()` hit it. This lifts the "`__call` does not replace a refused call" limit of the method visibility box
- `go test ./...` and `go test -tags phpx ./...` clean; `phpx test --matrix tests/fixtures/...` 298 of 298; `phpx test --matrix phpx/testdata/...` 191 of 192, the one failure an overlay copy under `patches/` that cannot find its sibling directory (not touched by this run)
- not run: the atkins pipeline, `lint-delta`, `go get ./...`

## 2026-11-05 - upstream f47c42e: the standard library nests by runtime

- follow `f47c42e` (#209): `stdlib/<pkg>` is now `stdlib/ps/<pkg>`, the go bindings sit in `stdlib/go/<pkg>`, and `stdlib/px` is where a host overlay goes. The fork's `stdlib/phpx/` moved to `stdlib/px/phpx/`, its `phpx_*.go` files in `stdlib/{core,compat,crypto,files,pexec}/` to `stdlib/ps/<pkg>/`, and the overlays under `phpx/testdata/patches/stdlib/` the same way. That move is its own commit, made before the follow because a follow leaves a fork-owned file where it lies. `follow.sh` owns `stdlib/px/*` besides `stdlib/phpx/*`
- the overlays of `stdlib/ps/core/{reflect.go,json_test.go}`, `stdlib/ps/internals/internals_test.go`, `list/stdlib_test.go`, `runner/runtime.go` and `internal/phpval/phpval.go` were three-way merged onto the new upstream (`git merge-file`, base the old upstream, theirs the new one)
- earlier runs had edited upstream files in place and no follow carried the edit across, so the follow dropped them: `runner/invoke.go` (host class arguments), `runner/runner_load.go` (autoload re-entry), `stdlib/ps/compat/compat.go` (`phpxRegisterBuffers`), `stdlib/files.go` (`PhpxRunLate` after `RegisterFS`; `csv` and `file_get_contents` overrides were lost with it), `stdlib/areas.go` (the late area in the inventory), `tests/fixture.go` and `tests/fixture_load.go` (the `phpx:` fixture key), two parser and apidoc tests, and seven fixtures with php's answers. All are overlays now. `runner.NamespaceLate` is a fork-owned constant in `runner/phpx_namespace.go`
- `internal/apidoc/godoc.json` and `docs/reference/extensions/implemented-apis.md` regenerated with `-tags phpx`
- `go test ./...` and `go test -tags phpx ./...` clean; `phpx test ./tests/fixtures/...` 298 of 298; `phpx test --matrix tests/fixtures/...` 298 of 298 on flatstack, runtime and php; `phpunit/compare.sh` prints `same` for each file
- not run: the atkins pipeline, `lint-delta`, `go get ./...`

## 2026-11-05 - phpunit.yml and a glob runner

- upstream is still `de399bb`; nothing to follow
- `phpx/testdata/phpunit/phpunit.yml` lists the test globs and `phpx/scripts/phpunit.sh` runs each matched file through `launcher.php`; its totals match `compare.sh` for all nine files. The `phpx phpunit` subcommand is still open (needs a seam in `cmd/phpscript`)
- no Go tests, matrix or pipeline run this time

## 2026-11-05 - the walls an app run met: namespace\\f(), typed and intersection constants, mb_ucfirst

- upstream is still `de399bb`; nothing to follow. A previous run left this work uncommitted; this run verified it (`go test -tags phpx ./...` clean, 181 of 181 fork fixtures, each new fixture matches php) and landed it
- `namespace\f()` and `namespace\CONST` name a symbol of the current namespace, `use const A as B` aliases a constant, a typed class constant (`const int X = 1;`) and an intersection or parenthesised DNF type parse, `$obj::class` on a value, `[$a, $b] = $arrayAccess` reads through `offsetGet`, `mb_ucfirst`/`mb_lcfirst`, `strstr` with `$before_needle`, `stream_get_filters`
- fixtures: `functions/app_walls`, `functions/mb_first`, `syntax/namespace_relative`, `syntax/typed_const_intersection`, `classes/static_on_value`

## 2026-11-05 - the scope check settles on one file set

- upstream is still `de399bb`; nothing to follow
- the table's Slim and Latte columns skipped phpunit (a dev dependency in each `vendor/`) while the tally script did not, which was the whole of the mismatch. `extension-tally.php` now skips `vendor/phpunit`, `sebastian`, `phar-io` and `theseer` by default (`PHPX_TALLY_SKIP=` empty skips none), and its output matches the table's mbstring, dom, libxml, filter and pcntl rows. Laravel and Filament libxml read 17 and 23 calls, not 0, and are corrected. The scope-check box is ticked; the remaining rows are unchecked

## 2026-11-05 - the framework boxes tick

- upstream is still `de399bb`; nothing to follow
- Slim (5 tests), Latte (5), Laravel (4) and Filament (2) print the same `OK (...)` under php and phpx at this head, so their boxes in `README.phpx.md` and `ACCEPTANCE.md` are ticked, with the launcher command as the proof (`phpx test` over a folder finds no `.phpt` and never settled them)

## 2026-11-05 - bindings in the Runtime\PS, Runtime\PX and Runtime\Go namespaces; alternative syntax; WordPress boots to its database check

- every binding installer registers under an area (`runner.RegisterArea`): `stdlib/<folder>` is `Runtime\PS\<Folder>`, a file of `stdlib/core` is `Runtime\PS\Core\<File>`, `stdlib/phpx` is `Runtime\PX\<File>`. Every `Runtime\PS` area installs before every `Runtime\PX` area, so the overlay no longer depends on which Go package initialised first
- `runner.areas` in the configuration (`Options.Areas`) names the namespaces the default runtime installs; empty installs all. A function answers through its area: `\Runtime\PS\Core\Strings\strlen($s)` is `strlen`
- `phpscript list --areas` lists the areas with what each defines and replaces, `--bindings` has an Area column, and `implemented-apis.md` groups by area. `docs/reference/extensions/runtime-areas.md` describes it
- opt-in `Runtime\Go` areas map go's import paths under `stdlib/gostd`: `Runtime\Go\Strings` and `Runtime\Go\Path`, imported with `use`
- the alternative control syntax (`if (): ... endif;`, `while`, `for`, `foreach`, `switch`), `<?= expr ?>`, `$cls::$prop = v`, `ob_clean`, `ob_flush`, `ob_get_length`, `ob_list_handlers`, `header_remove`, `headers_list`, `flush`; wordpress 6 parses whole and stops at the missing `mysqli`
- a fixture marked `phpx: true` loads in a phpx build only, which fixes the untagged `go test ./tests` run that failed on fixtures the fork had rewritten to php's answers

## 2026-11-05 - carbon passes; the date classes are real classes

- `nesbot/carbon` 4/4: `class Carbon extends DateTime` works. A script class that extends a host class inherits its constructor, clones its host value, reads and writes its public properties (`DateInterval::$y`, `$f`, `$days`), passes as the host type into a binding, answers `static::createFromFormat()` with an instance of itself, runs `parent::format()` on the value behind `$this`, and gets `DateTimeInterface`'s constants; a method the host lacks falls to the class's own `__call`
- `DateTime::diff` counts the calendar as php 8.1 does in both directions, `createFromFormat`, `setISODate`, `DateInterval::format`, `createFromDateString`, `date_parse`
- an output parameter (`preg_match_all(..., $parts)`) holds an array `array_shift` changes in place

## 2026-11-05 - the date grammar, resources, the zone

- `strtotime` and the DateTime constructors read php's grammar (`stdlib/phpx/dateparse.go`): ISO, American and European dates, month and weekday names, clock times with a meridian, zones by abbreviation, offset and name, `+1 week 2 days ago`, `next friday`, `monday next week`, `first day of next month`, `last friday of this month`, `@stamp`. A hundred and seven specs differed from php; six do now, all odd ones (`back of 7pm`, `dec 25 noon`)
- `date_default_timezone_set` works and `date()`, `mktime()`, `strtotime()` and the constructors read it; `mktime`, `gmmktime`, `checkdate`, `idate`, `getdate`
- streams are resources with ids and a closed state: `(string)` gives `Resource id #5`, `gettype` and `get_debug_type` follow `fclose`, `tmpfile`, `umask`
- `isset($o->method)` is false; `ReflectionClass` over a trait; `ReflectionProperty` over a property the object grew

## 2026-11-05 - mustache's suite and what it found

- a class declared inside a block (`if (!class_exists('X')) { class X {} }`) is declared when the statement runs; an autoloader registered as `'Class::method'` for a class that has to be loaded no longer recurses until the Go stack overflows (php answers "not found" for a class already loading); `__halt_compiler()` stops the lexer and `__COMPILER_HALT_OFFSET__` is the offset; `file_get_contents` takes its offset and length; `trait_exists`, `class_uses` and `ReflectionClass` over a trait
- the missing string functions: `strtok`, `similar_text`, `str_increment`/`str_decrement`, `stristr`, `strrchr`, `chunk_split`, `quotemeta`, `count_chars`, `substr_compare`, `str_word_count`, `str_ireplace`, `sscanf`, `soundex`, quoted-printable, `htmlspecialchars` and its decode and entity relatives with their flags, `fprintf`/`vprintf`/`vfprintf`, `md5_file`/`sha1_file`, `array_multisort`

## 2026-11-05 - emailvalidator's suite, encodings, the throwable tree

- `egulias/EmailValidator`'s own suite: 358 tests, 472 of 477 assertions, the rest an intl `Spoofchecker`. What it found: `/u` makes `\w \d \s \b` Unicode and refuses invalid UTF-8 (`preg_last_error()` 4), `preg_grep` takes `PREG_GREP_INVERT`, DNS functions over Go's resolver, `mb_detect_encoding('auto')`, and `mb_convert_encoding` between Windows-1252, ISO-8859-x, UTF-16/32, SJIS, EUC and the rest (x/text), `utf8_encode`/`utf8_decode`, `iconv`
- a catch clause takes a child of a built-in throwable (`catch (RuntimeException)` takes `UnexpectedValueException`), `is_subclass_of` and `get_parent_class` agree, and `AssertionError`, `CompileError`, `ParseError` exist. Two upstream tests that pinned the old rule (`TestMatchCatchTypeUsesTheClassName`, `tests/fixtures/exceptions/throwable_classes.phpt`) are overlaid with php's answers

## 2026-11-05 - upstream suites, preg_replace over lists

- `phpx/testdata/upstream/run.sh <name> <git url> <tests>` clones a library, installs it with phpunit and runs its own suite under php and phpx: pimple 68 of 68 and doctrine/inflector 1215 of 1215 match php
- `preg_replace` takes lists of patterns, replacements and subjects, a limit and the count by reference; `preg_filter`, `preg_replace_callback_array`; the replacement follows php's `\n`, `$n`, `${n}` rules
- spread (`f(...$x)`, `[...$x]`) unpacks generators, iterators and aggregates, with string keys as names; this fixes a regression from the generator signature change that took the inflector coverage to 18%
- `ReflectionClass` and `class_exists` know `Closure`, `Generator` and the core interfaces, and report the engine's own classes as internal

## 2026-11-05 - csv, spl files, the array callbacks

- `SplFileObject` and `SplTempFileObject` in the prelude, with php's flags and line numbering; `str_getcsv`, `fgetcsv` and `fputcsv` ported from php's own parser (blank line is `[null]`, multi-line fields, the escape character, the enclosure rule), installed after the filesystem is rooted by the new late-binding phase (`runner.PhpxRegisterLate`)
- `IteratorIterator` and `FilterIterator` cache the inner element as php does and answer `valid()` from the cache, which a stream-backed inner iterator needs
- the array_u* family, `array_walk` with a by-reference value, `natsort`, `array_merge_recursive`, `array_pad`, `array_find` and friends, `mb_convert_case` and the mb search functions, trig and log functions, the ctype classes, `yield from` over a Go slice, a reference to an array element, `$a[$k] = v` on an `ArrayAccess` inside an expression
- php's Mersenne Twister: `mt_srand(n)` gives the sequence php gives, `shuffle`, `str_shuffle` and `array_rand` draw from it
- 700 constants generated from php (`phpx/scripts/gen-consts.php`); `file()` honours its flags; a Go error reaches only a `catch` of a built-in throwable class

## 2026-11-05 - php-parser, mustache, uuid; carbon blocked

- `phpparser` 3/3, `mustache` 4/4, `uuid` 3/3 on phpx; `carbon` stays 0/4 behind `class Carbon extends DateTime` (a user class on a Go-backed builtin)
- `pack()` and the machine-order `unpack` codes; variable variables (`$$name`, `${expr}`, read and write); `CI::CONST` through `interface CI extends A, B`; `clone` of a value a binding returned; a promoted by-reference constructor parameter shares the caller's variable; `class_alias` over an interface, and `instanceof` against either name

## 2026-11-05 - twig, commonmark and yaml shims

- three more shims, all 14 tests green on php and on phpx: `twig` 5, `commonmark` 4, `yaml` 5
- what they needed: `new ($expr)(...)`; `ReflectionFunction` over a builtin or a named closure, with `getClosureThis` and friends; `method_exists('Class', 'm')` autoloads the class; `settype()`; `mb_check_encoding($s, $enc)`; a by-value `foreach` walks a snapshot when the body can reach the array (an unset, a call), as php does; reference wrappers are stripped for `$f(...)` over a builtin and for calls into a Go method; `Throwable` and the other engine interfaces no longer ask the script's autoloaders

## 2026-11-05 - references, and the four frameworks pass

- `phpx/testdata/frameworks`: slim 5/5, latte 5/5, laravel 4/4, filament 2/2 on phpx, matching php
- references: a by-reference parameter binds the caller's variable (`function inc(&$x)`), `$this->p = &$param` shares a property with it, `$a = &$b` and `use (&$x)` share a cell. Array-element references and `function &f()` returns stay value copies. `tests/fixtures/runtime/byref_noop.phpt` pinned the old by-value behaviour and now pins the alias
- latte's generated LALR parser needed three quiet fixes: `goto` to a label inside a nested `if`, an array literal that copies the arrays it reads from storage (`[$this->stack]`), and identifiers with bytes above 0x7F (`$ʟ_args`). Also `addslashes`/`stripslashes`, `extension_loaded` true for the extensions the runtime carries, a nested or non-array destructuring pattern

## 2026-11-05 - phpunit shims for the four frameworks

- `phpx/testdata/frameworks/{slim,latte,laravel,filament}` each hold a phpunit shim: `composer install --ignore-platform-reqs`, `launcher.php`, a `tests/` file, and `summary.sh` for the table in the README. 16 tests, all green on php 8.4.27, none yet on phpx; the first wall of each is in the README. The walls that were parse errors are gone: a method that returns by reference (`function &getIterator()`), `Class::{"name"}()`, a float written `.1`, `new static::$cls()`. What remains: `goto` (laravel's `helpers.php`, latte's lexer), PCRE `(*SKIP)(*F)` and recursion (FastRoute), `hash('xxh128')`
- `static::$prop` reads the called class's redeclared static, as php does (slim's PSR-17 factories pick their classes this way); `key_exists`, `sizeof`, `chop` and the other aliases

## 2026-11-05 - curl, and guzzle passes

- curl in Go over net/http, no cgo: `curl_init`, `curl_setopt(_array)`, `curl_exec`, `curl_getinfo`, `curl_errno`/`curl_error`, `curl_reset`, the multi interface (a transfer runs to its end inside `curl_multi_exec`), a share handle that shares nothing, `curl_version`, and the 679 constants generated from php. Header, write, read and progress callbacks, redirects, timeouts, proxies, TLS options, gzip, `file://`. `CurlHandle`, `CurlMultiHandle` and `CurlShareHandle` are the instance-of names. Guzzle's `CurlHandler` and `CurlMultiHandler` run against a `php -S` child: 147 tests, guzzle at 87.6%, eight of nine libraries over the 80% gate
- what it needed: `proc_open` (+ `proc_close`, `proc_terminate`, `proc_get_status`), `fsockopen`, magic properties (`__get`, `__set`, `__isset` inside `isset()`, `__unset`) and `__callStatic`, an output parameter that is a property (`curl_multi_exec($mh, $this->active)`), `(int)` of a handle, `use (&$x)` as a cell shared with its frame (the promise closure of the multi handler reads `$promise` after the frame ended), a class whose trait or parent a later include declares, `extract()`, `get_defined_constants(true)` by category
- phpscript #208 (a static call opens no span) merged and followed

## 2026-11-05 - upstream de399bb

- follow `de399bb`, 20 commits ahead of the last one
- `migrations/run.sh` wrote the generated tree again, 3 files dropped

## 2026-11-05 - library coverage, and what it found

- `phpx/testdata/packages/coverage.sh` measures statement coverage of the nine libraries under phpx and gates each at 80%; seven pass, guzzle (53.4%, no curl) and console (64.2%) do not. The table, its blocker column and the by-design list are in `phpx/TESTING.md`
- found by the new tests and fixed: `break N` and `continue N`; `json_encode` flags (slash and unicode escaping, pretty print, errors) and a `json_decode` that returns null and sets `json_last_error`; `set_error_handler` receiving `trigger_error` and a failed `fopen`; braced `namespace X { }`; `#[Attr]` inside a parameter list; a catch type written through a `use` import; `date()`/`gmdate()` with every format letter; `filter_var` with its constants; `gz*`, `inet_*`, `hash_init`/`update`/`final`, `strnat*`, `wordwrap`, `array_replace_recursive`, `grapheme_*`, `stream_copy_to_stream`, `stat`, `symlink`, `get_class_methods`; `data:` URIs; `stream_get_meta_data`; `parse_url` of `host:port` and IPv6; `isset` of a string offset past the end (also inside a namespace, where it was always true); a by-value `foreach` over rows a binding returned; sibling classes calling an overridden protected method

## 2026-11-05 - follow upstream af52ba4 (Runtime\Go, bindlint)

- upstream added the `Runtime\Go` bindings (#202) and `bindlint` (#203), a gate that fails a binding taking or narrowing a runtime value type. Overlays rebased with a three-way merge against the old base: `parser/parser.go` (`goRuntime` field, `[]T` types, `use` hook, kept `private \stdClass $x;`), `internal/phpval/phpval.go`, `runner/expr/closure.go`, `direct_operators.go`, `flatstack.go`, `runner_assign.go`, `helpers.go`
- upstream's `phpArith` now answers `(value, error)`, so the fork's division-by-zero check sits at one seam, the head of `phpArith` (overlaid `runner/operators.go`), where it sat in three call sites and a `DivZero` hook before. The `DivZero` field of `expr.Helpers` is gone. Upstream still answers 0 for `/` and `%` by zero
- `$s = Go\make("[]string", 2)` stays a slice: the fork's copy of a binding-returned slice into an array (`phpxOwnedArray`) is skipped once the program imported a namespace (`rt.phpxOwned`)
- `get_object_vars` hands a value that is no script object to upstream's binding, which now reads the exported fields of a struct
- `is_a`/`gettype` no longer narrow to `*model.Object` (bindlint). The 25 findings left in the fork's bindings are a baseline in `internal/cmd/bindlint/phpx_baseline_test.go` (overlaid `main_test.go`): the count may fall and may not grow. Fixing them is the next chore: each takes `any` and reads through `model.RangeValues`
- `list-apis` regenerated the two reference pages

## 2026-11-04 - what 75 more library tests found

- `/` and `%` by zero raise `DivisionByZeroError` ("Division by zero", "Modulo by zero") on every path: the interpreter, the bytecode engine and `/=`, `%=`. They answered 0 before, which hid a failed `intdiv`-style guard in library code. An error the runtime raises now answers `getFile()`/`getLine()` of the statement that failed (interpreter only; the bytecode engine keeps no site)
- `is_subclass_of` takes a class name and an interface (and `Stringable` for anything with `__toString`); `interface_exists('Throwable')`; `gettype` of a stream is `resource`; `defined('Enum::Case')`; `PCRE_VERSION`; `fileinode`
- reflection reports class names the way php does: `?Foo\Bar`, `Foo\Bar|int`, `?Closure` for `Closure|null`, through the namespace and the `use` imports (`Param.Resolved`, `FuncDecl.ResolvedReturn`, `Field.Resolved` beside the spelling the formatter prints). `ReflectionParameter::getDefaultValue` on an interface method
- regex: the `A` modifier anchors at the offset (symfony's `StringInput` tokenizer needs it), possessive quantifiers after a group (`(?:...)++`) and a count (`{2}+`) read as greedy
- a class whose parent only an autoloader can supply is declared where its statement runs, not at hoist (php's own order: `require 'vendor/autoload.php'; class Cmd extends Vendor\Base {}` in one file); `is_iterable` is true for a `Generator`, an `ArrayIterator` and an `IteratorAggregate`; `mb_detect_encoding` and `mb_convert_encoding` convert between UTF-8, ASCII and ISO-8859-1 and map arrays (they returned their input)
- the library suites: 120 tests, php 120, phpx 120. Not done: `idn_to_ascii` (needs `golang.org/x/net/idna`, which go.mod does not carry)

## 2026-11-04 - exceptions know where they were made

- `getFile()` and `getLine()` answer the file and line of the `new`, for a built-in (`new RuntimeException`, through the context the constructor is given) and for a declared subclass whose constructor never calls the parent (`runner/phpx_throwsite.go`, stamped in `helperNew`). `classes/exception_site_test.php` was settled against php 8.4. Errors the runtime raises itself (TypeError, DivisionByZeroError) still answer an empty file and line 0; `getTrace()` is still empty

## 2026-11-04 - phpunit mocks run

- `createStub`, `createMock`, `getMockBuilder`, `expects($this->once())`, `->with()`, `willReturn*`, `willThrowException`, `never()`, `exactly()`, mocking an interface and a class, and a failed expectation's message: `tests/MockTest.php` (9 tests) prints what php prints. What it took: `eval()` of source (declarations are hoisted, a return gives the value, a bad source raises `ParseError`), `serialize`/`unserialize`, `func_num_args`/`func_get_arg`, `debug_backtrace` with objects (phpunit finds the running TestCase on it), an abstract trait method answered by a concrete one, `ReflectionClass::isReadOnly`, and an array callable `[Class::class, 'method']` that autoloads its class. `eval` is in upstream's "Won't build" table; the fork builds it, for generated class source

## 2026-11-04 - token_get_all as php has it

- `token_get_all` and `token_name` in `stdlib/phpx/tokenizer.go` return php 8.4's token ids and all of its tokens: the 152 `T_*` constants, `TOKEN_PARSE`, names (`T_NAME_QUALIFIED`), casts, `T_ATTRIBUTE`, `T_ENUM`, `T_MATCH`, `T_THROW`, heredoc and interpolation pieces, `T_DOC_COMMENT`, the two `&` kinds, `T_OPEN_TAG_WITH_ECHO`. `parser/tokenizer.go` (overlaid) carries php's numbering for the ids it had, so upstream's tests agree. The generated table is `stdlib/phpx/tokens_gen.go` (`php phpx/scripts/gen-tokens.php`). Checked against php token for token on 1,100 files from phpunit, sebastian, symfony, guzzle, monolog, doctrine, php-parser and the fixtures: no difference; `functions/token_get_all_test.php` pins ten samples. This is what phpscript needs too: to be proposed upstream once it has settled here
- not done: `PhpToken` (the class form), backtick strings, `public(set)` asymmetric visibility tokens

## 2026-11-04 - nikic/php-parser leaves the library set

- the library suites are nine, not ten: `nikic/php-parser` is not wanted in the set (an earlier note said it needed the tokenizer extension; `token_get_all` was in phpscript all along, and what it lacked was constants). `PhpParserTest.php` and the entry in `composer.json` and `package-survey.sh` are gone (phpunit still requires it as a dependency). 45 of 45 pass

## 2026-11-04 - 46 of 50, and a measurement of each unit

- `unpack()`, the `PATHINFO_*` constants and `pathinfo()` with flags, `IteratorAggregate::getIterator()` returning a generator, and a generator closure that keeps what its enclosing closure captured: symfony/console 5 of 5, monolog 6 of 6, 46 of 50 across the libraries; only php-parser (tokenizer extension) is short
- `phpx/scripts/measure.py` measures each fixture area and phpunit file on its own through a coverage build: [phpx/MEASUREMENTS.md](phpx/MEASUREMENTS.md)

## 2026-11-04 - follow upstream 8960ad6, phpx.go is gone

- phpscript #207 reads `argv[0]` as `<binary>-<tool>` and puts the tool first (`toolchain.CommandArgs`), and a `-cover` build started without `GOCOVERDIR` sets `cover/<binary>` and starts itself again (`toolchain.CoverRelaunch`). `phpx.go` and its test are removed; `phpx test` reaches `phpx-test` through upstream's discovery and writes `cover/phpx-test`. The coverage guide in `docs/guides` is upstream's

## 2026-11-04 - upstream 8960ad6

- follow `8960ad6`, 16 commits ahead of the last one
- `migrations/run.sh` wrote the generated tree again, 1 files dropped

## 2026-11-04 - follow upstream 2011c01, toolchain discovery

- followed upstream to `2011c01`: phpscript #206 discovers `<binary>-<tool>` on PATH, so `phpx test` reaches `phpx-test` without the fork's exec. `phpx.go` keeps what discovery cannot do: a discovered tool is given its arguments without its own name, so `phpx-test` puts `test` in front, and it sets `GOCOVERDIR` and re-execs itself because the coverage runtime reads that before any init. `godoc.json` and the reference pages were regenerated

## 2026-11-04 - upstream 2011c01

- follow `2011c01`, 15 commits ahead of the last one
- `migrations/run.sh` wrote the generated tree again, 1 files dropped

## 2026-11-04 - phpx-test, the coverage build

- `go build -cover -covermode=atomic -coverpkg=./... -tags phpx -o bin/phpx-test .` is the coverage build (`atkins -f atkins.phpx.yml build:test-binary`, which also installs it under `PHPX_TEST_PREFIX`, `/usr/local/bin` by default). `phpx.go` makes `phpx test ...` exec a `phpx-test` found beside it or on PATH, and `phpx-test` set `GOCOVERDIR` (`cover/phpx-test`) and re-exec itself, because the coverage runtime reads that variable before any init runs. `atkins -f atkins.phpx.yml cover:phpx-test` runs both matrices through it and writes `cover/phpx-test.cov`. `go install` has no `-o`, so the job is `go build`. phpscript issue #195 asks for the general `<binary>-<tool>` discovery; the `phpx.go` hand-off is the fork's version until that lands

## 2026-11-04 - a phpunit suite for each of the ten libraries

- `phpx/testdata/packages/tests`: 10 files, 50 tests, `summary.sh` prints php against phpx. 10 of 50 passed at first, 39 of 50 now. Generators, SPL file iterators, `WeakMap`, `array_map` keys, `parse_url` and a dozen functions landed for it; Monolog waits on the date classes and php-parser on the tokenizer extension

## 2026-11-04 - more phpunit suites

- `OopTest`, `FunctionsTest`, `EnumMatchTest` join the five phpunit files: 8 files, 56 tests, all printing what php prints. They found four bugs, now fixed: `use (&$x)` did not write back (a closure that outlives its frame still does not), instantiating an abstract class raised no catchable `Error`, `UnitEnum`/`BackedEnum`/`UnhandledMatchError` were not classes, and a trait in another file was not found
- open: `createMock`/`createStub` (next wall: `ReflectionClass::isReadOnly`, then `eval`)

## 2026-11-04 - upstream 2a897ed

- follow `2a897ed`, 14 commits ahead of the last one
- `migrations/run.sh` wrote the generated tree again, 1 files dropped

## 2026-11-04 - follow upstream f57b4e7, and the prelude the follow dropped

- followed upstream `8b460c4` -> `f57b4e7` (`build(deps)`), by a run of the default pipeline at the end of the phpunit work
- found: `follow.sh` dropped `runner/phpx_prelude/reflection.php` and `spl.php`, the PHP the prelude embeds, because the fork-owned rule matched `phpx_*.go` and not a directory named `phpx_*`; the build broke on the `go:embed` and the regenerated `godoc.json` and reference pages lost the fork's entries. `fork_owns` now takes `*/phpx_*/*`; the files and the regenerated documents are back in the commit after the follow

## 2026-11-04 - upstream f57b4e7

- follow `f57b4e7`, 12 commits ahead of the last one
- `migrations/run.sh` wrote the generated tree again, 3 files dropped

## 2026-11-04 - phpunit runs

- phpunit 11.5 runs from its source tree: test discovery, `setUp`, assertions, `#[DataProvider]`, `#[TestWith]`, `#[Depends]`, exceptions, skips. `cd phpx/testdata/phpunit && ./compare.sh phpx` runs five test files under php and phpx and diffs what phpunit prints: all five print the same. [TESTING.md](phpx/TESTING.md) carries the before and after numbers and the coverage. `phpunit/phpunit/src` lints 895 pass, 93 warn, 0 fail (was 795, 483, 25); the top ten composer libraries fail 18 files (was 71)
- syntax: destructuring (`[$a, [$b]] = `, `foreach ... as [$a, $b]`), heredoc and nowdoc, `and`/`or`/`xor`, `for` lists and `for (;;)`, `a ?? b = c`, assignment expressions on any target, named arguments, array spread, `print`, `match`, `throw` and `clone` as expressions, `enum`, `use A\{B, C}`, bare blocks, `$cls::NAME`/`$cls::method()`, string offset writes
- runtime: Reflection with attributes and doc comments (a PHP prelude over a source scan), array value semantics, `json_encode` of `JsonSerializable`, `php://memory`, the SPL structures, Iterator/IteratorAggregate/Countable/ArrayAccess on script classes, `new Exception($m, $c, $previous)`, `instanceof Throwable`, object comparison, `parent::` through a chain, a namespaced `catch`, Go slices that become script arrays, the `E_*`/`JSON_*`/`SORT_*` constants and about forty functions
- dropped, because they pinned decisions this fork reverses: `tests/fixtures/arrays/unset_native_slice_reindexes.phpt`, `tests/fixtures/stdlib/json_encode_flags.phpt`, two tests in `stdlib/core/json_test.go`, two cases in `tests/bindings_test.go` (reworded). Each has a fixture under `phpx/testdata` that settles the new answer against php
- upstream: [#204](https://github.com/titpetric/phpscript/pull/204) (`files.RegisterWrapper`) and [#205](https://github.com/titpetric/phpscript/pull/205) (`Runtime.RegisterClassSource`); [phpx/testdata/UPSTREAM.md](phpx/testdata/UPSTREAM.md) lists the seams that retire the other overlays
- open: a throwable's file and line are not recorded, `createMock` needs `eval`, `class_alias` does not feed `instanceof`, braced `namespace X { }` does not parse
- limit: the bytecode engine keeps upstream's answer for a string offset write and a Go slice result; two fixtures opt it out

## 2026-11-04 - `?->`

- `?->` parses and runs, on a property, a method call and a chain. `phpxNullsafe` and `phpxWrapGuards` (`parser/phpx_nullsafe.go`) record each guard in the postfix loop of the overlaid `parser/expr.go` and wrap the finished chain in `($t = a) === null ? null : <chain on $t>`, so a null skips the rest of the chain and its arguments, and the base evaluates once. A second `?->` nests in the first's else arm. The lexer (overlaid) reads `?->` as one operator
- `phpx/testdata/functions/nullsafe_test.php`, expected output from `php`: property, method, chain, short-circuited arguments, single evaluation, `??`, a condition, string concatenation, a loop. Matrix: 47 of 47 on all three runtimes
- limit: `$a?->b = 1` is not refused where php fatals; `?->` inside a double-quoted string is not read
- upstream still `8b460c4`; `go get ./...` and the atkins pipeline were not run
- next: `match`, `enum`

## 2026-11-04 - `$obj::class`

- `$obj::class`, `f()::class` and `$this->o::class` parse and run. `phpxClassOf` (`parser/phpx_classof.go`) is one seam in the postfix loop of the overlaid `parser/expr.go` and lowers the form to `get_class($obj)`; `Name::class` is read earlier and unchanged. Upstream is still `8b460c4`, so no follow; `go get ./...` was not run
- `phpx/testdata/functions/class_of_object_test.php`, expected output from `php`. Matrix: 46 of 46 on all three runtimes
- limit: a non-object operand returns get_class's answer, where php 8 throws a TypeError
- next: `?->` (nullsafe), `match`, `enum`

## 2026-11-04 - follow upstream 8b460c4, `<=>`

- followed upstream `ff20a8c` -> `8b460c4` (`list --classes --json`). No overlay drifted. `test:introspection` regenerated `godoc.json` and two reference pages, and `TestRunListsClasses` passed again. `go get ./...` was not run
- `<=>` parses and runs. The lexer reads it, `binPrec` puts it level with `==`, and `phpxBinary` (`parser/phpx_spaceship.go`) lowers it to `__phpx_spaceship(a, b)`, bound in `stdlib/phpx/spaceship.go` on `phpval.Compare`; NAN answers 1 as php does. Lowering to a call avoids an operator in both engines and in `runner/expr`, at the cost of the formatter printing the call
- `phpx/testdata/functions/spaceship_test.php`, expected output from `php`: ints, strings, numeric strings, arrays, null and bool, NAN, single evaluation of operands, `usort`, precedence against `+`. Matrix: 45 of 45 on all three runtimes
- next: `$obj::class` (32 files), `enum` (13), `match` (8), `?->`

## 2026-11-04 - upstream 8b460c4

- follow `8b460c4`, 11 commits ahead of the last one
- `migrations/run.sh` wrote the generated tree again, 2 files dropped

## 2026-11-04 - `??` with a call on the left

- `f()['k'] ?? b`, `$o->m()->p ?? b` and `f() ?? b` parse and run; the 6 files that stopped on this wall, `Application.php:1138` among them, now get past it. `phpxCoalesce` stores the left side once in a generated temporary, `($t = left) !== null ? $t : b`, and keeps `isset(a) ? a : b` for a side-effect free left. The runtime answers null for a missing key or property, so no `isset()` is needed
- found: a variable assigned inside an expression and read again in that expression read null the first time it ran, because the compiled expression reads the slot Eval bound before the assignment (`(($t = f()) !== null) ? $t : 'd'`, and `isset()` on it, are php-valid). The `AssignExpr` case in the overlaid `runner/expr/direct.go` now writes the slot too. An upstream bug, a candidate for a `phpx:` PR
- `phpx/testdata/functions/coalesce_call_test.php`, expected output from `php`: call under an index, a missing key, a null result, a chain, a builtin call, a deep miss. Re-lint of the phpunit tree: phpunit/phpunit src 83 failing files (was 84), sebastian 23 (24), phar-io 1 (3), nikic/php-parser 11
- next: `$obj::class` (32 files), `enum` (13), `match` arms (8)

## 2026-11-04 - scope: OOP baseline, extensions out

- `README.phpx.md` gains a Scope section: the OOP baseline is the scope, extensions are out of it, and a wall is classed in scope or out before it counts as a blocker. PDO and more standard library surface are the only candidates. phpunit runs minimally: a launcher calls `PHPUnit\TextUI\Application::run` and skips the entry script's `extension_loaded()` gate, with no `phpunit.xml`; coverage, XML logging and the phpt runner are unsupported
- `phpx/scripts/extension-tally.php` counts calls and class uses per extension with php's own `ReflectionFunction::getExtensionName`. `phpx/testdata/report.md` now holds the totals for phpunit, the ten packages and the four frameworks, and the biggest non-OOP blocker of each. dom (326 uses) and xmlwriter (186) lead in phpunit; mbstring leads everywhere else; PDO (306 and 329 uses) is the one extension in reach, in Laravel and Filament
- `packages/README.md`, `frameworks/README.md` and `ACCEPTANCE.md` point at the scope check. No Go code changed; no agents were dispatched, the tally was one script

## 2026-11-04 - follow upstream ff20a8c, arrow functions

- followed upstream `e40e38b` -> `ff20a8c` (trace_level, `Runtime\Router::Bind`, `list` API docs). The overlays of `runner/runner.go` and `stdlib/imports.go` held the old upstream files and the follow reverted `traceIncludes`, the `TraceLevel` gate and the `stdlib/runtime` import, which failed three observer tests; both overlays now carry upstream's hunks. `go get ./...` was not run
- `fn($x) => expr` parses and runs. `parser/phpx_arrow.go` parses the parameters, an optional return type, `=>` and an expression into a `model.Closure` whose body is one `Return`, and whose `Uses` are the variables the body's tokens name that are not parameters (php's by-value auto capture). `fn` not followed by a parameter list and `=>` falls back to upstream's reading, which `TestIfConditionSyntax` needs (`fn()` as a function name). Two seams in the overlaid `parser/expr.go` call `phpxClosureOf`; `static fn` takes the same path
- `phpx/testdata/functions/arrow_function_test.php`, expected output from `php`: two parameters, capture by value, nested arrows, typed, `$this` in a method, `static fn`, a parameter shadowing a captured name, `array_map`, `usort`
- limit: over-capture is harmless but an unset outer variable becomes null inside. A nested closure's own parameters are captured from the outer frame too, unused
- `atkins.phpx.yml` `test:introspection` lacked upstream's `list-apis --index` line, so `internal/apidoc/godoc.json` stayed upstream's and `list`/`list-apis` tests failed after a follow (the fork's registrations are not in upstream's index). Added, and both `list-apis` calls now run with `-tags phpx`; the index is regenerated and committed. Run `test:introspection` after every follow
- found: `<=>` is a parse error. Next wall for phpunit with `?->`, `match`, `enum`, and `extension_loaded()`

## 2026-11-04 - upstream ff20a8c

- follow `ff20a8c`, 10 commits ahead of the last one
- `migrations/run.sh` wrote the generated tree again, 1 files dropped

## 2026-11-04 - `??` and `??=`

- upstream is still at `e40e38b`; `go get ./...` was not run
- upstream parses neither `??` nor `??=`: `$a ?? 1` stopped at "unexpected token ?". The lexer (overlaid `parser/lexer.go`) now reads both, `parseTernary` parses the coalesce level through `phpxParseCoalesce` (right associative, tighter than `?:`, looser than `||`), and `??=` lowers to `a = isset(a) ? a : b` in `parseAssign`. The code is `parser/phpx_coalesce.go`; the overlaid `expr.go` and `parser.go` carry the hooks and the `??=` entry in `isAssignOp`
- `phpx/testdata/functions/null_coalesce_test.php`, expected output from `php`: undefined variable, null/false/0/"" left, missing and nested index, property, chain, `??=` on variable, index and property, precedence against `?:`
- limit: the left side is evaluated twice (isset, then read), so a call there is refused at parse time; php accepts it. Formatter prints the lowered form
- next walls for phpunit: `fn()`, `?->`, `match`, `enum`, then `extension_loaded()` for dom/libxml/tokenizer/xmlwriter

## 2026-11-04 - argument unpacking, PHP_BINARY, version_compare `=`

- upstream is still at `e40e38b`; `go get ./...` was not run
- `f(...$args)`, `$o->m(...$a)`, `C::s(...$a)`, `new C(...$a)` and `$f(...$a)` parse and run, with several spreads and spreads mixed with plain arguments before them. The parser (overlaid `parser/expr.go`, `parseArgs`) wraps the operand in a `model.Unary` with op `...`, so `model.Call` and the formatter need no change and the bytecode engine declines the unknown operator and the whole program runs on the interpreter. `runner/expr/phpx_spread.go` compiles a call that holds a spread into `spreadCall`, one seam at the top of `compile` in the overlaid `runner/expr/direct.go`. A call with no spread compiles as before
- `phpx/testdata/functions/argument_unpacking_test.php`, expected output from `php`, and `TestParseArgumentUnpacking`
- limits: only arrays unpack (a `Traversable` raises "Only arrays and Traversables can be unpacked"), and a string key is passed positionally where php names the argument. php's compile error for a positional argument after a spread is not raised
- `PHP_BINARY` is the running binary (`stdlib/phpx/environment.go`); phpunit reads it
- `version_compare()` accepted `==` and `eq` but answered null for `=`, which phpunit's entry script uses (`version_compare(PHP_VERSION, PHP_VERSION, '=')`). Added to `phpx/testdata/constants/version_compare.phpt`
- real phpunit 11 now parses `Assert/Functions.php` and starts its entry script. The walls behind it, none of them OOP: `extension_loaded()` answers false, so the script stops at "requires the dom, filter, json, libxml, mbstring, tokenizer, xmlwriter extensions"; with that check cut out, `TextUI/Application.php:1138` stops on `??`. Across `phpunit/` and `sebastian/` of the install, 39 of 1481 files use `??`, 49 use `fn(`, 13 `match`, 13 `enum`. Those are recorded in `phpx/testdata/report.md` as outside the five reversals and not started

## 2026-11-04 - namespaced top-level statements and conditional function declarations

- upstream is still at `e40e38b`; `go get ./...` was not run
- the parser no longer refuses a top-level statement in a namespaced file (overlaid `parser/parser.go`); the include runs it as php does. `TestParseNamespacedStatementAccepted` replaces upstream's `TestParseNamespacedStatementRejected` in the overlaid `parser/parser_test.go`
- a `function` statement below the top level declares its function when it runs: `phpxDeclareFunc` in `runner/phpx_funcdecl.go`, one seam in `exec` (overlaid `runner/runner.go`), untagged stub keeps upstream's no-op. `if (!function_exists('f')) { function f() {} }` and a function declared inside another function's body work; a name already registered is left alone
- the bytecode engine declines a program that declares a function below the top level (overlaid `flatstack/engine/program_compiler.go`), so the matrix passes on all three runtimes
- `phpx/testdata/functions/conditional_decl_test.php` with `conditional_decl_ns.inc`, expected output from `php`
- real phpunit 11 now gets past `deep_copy.php`; the next blocker is argument unpacking `f(...$args)` (parse error, `Assert/Functions.php:3226`), a general language gap. Also found: a namespaced `const X = 1;` defines `X`, not `Lib\X`

## 2026-11-04 - Closure::bindTo and Closure::call

- upstream is still at `e40e38b`; `go get ./...` was not run
- `$f->bindTo($this, $scope)` answered "undefined method" and `$f->call($o, ...$args)` reached upstream's Go `Callable.Call` and passed `$o` as an ordinary argument. `phpxClosureMethod` (`runner/phpx_closure.go`) answers both on a closure literal through `PhpxRebind`; `call` runs in the object's class scope. One seam at the top of `callGoMethod` in the overlaid `runner/dispatch_method.go`; untagged stub answers "not handled"
- `phpx/testdata/classes/closure_bindto.phpt`, expected output from `php`: private read through bindTo, call with an argument, a bound-method closure rebound to another class, bindTo without a scope still refused, a static closure bound to a class, `call(5)` TypeError
- limits: `Closure::fromCallable` scope, and a closure the bytecode engine built, still answer unchanged

## 2026-11-04 - Runtime.Vendor and Closure::bind scope

- upstream is still at `e40e38b`; `go get ./...` was not run
- new objective from the run input: phpunit runs from its source tree, `go:embed`ed into phpx, and not from `phpunit.phar`. `rt.Vendor(fsys)` (`runner/phpx_vendor.go`) layers a tree behind `Options.RootFS` (the root wins a path both hold) and, when `Options.Include` is empty, sets it to `vendor/autoload.php`. Opt in: a runtime that never calls it reads the root alone, and the `--include` flag was already the opt-in for the CLI. `TestPhpxVendorAutoloadsFromAnEmbeddedTree`
- found by running composer's phpunit 11 install under `phpx run --include vendor/autoload.php`: composer's `ClassLoader` reads its private properties through `Closure::bind(fn, null, ClassLoader::class)`, which property visibility (previous entry) now refused. Upstream's `Closure::bind` answered the closure unchanged and refused a non-null `$this`. `stdlib/phpx/classes.go` re-registers it on `Runtime.PhpxRebind` (`runner/phpx_closure.go`): `$this` and the class scope are replaced in the closure's environment. A closure the bytecode engine built still answers unchanged
- `phpx/testdata/classes/closure_bind.phpt`, expected output from `php`. Upstream's `stdlib/core/closure_test.go` pinned the `$this` refusal and is overlaid without that case
- blocker, next for phpunit: the loader's first `files` entry, `myclabs/deep-copy/src/DeepCopy/deep_copy.php`, is a namespaced file with a top-level `if (!function_exists())`; the parser refuses it (`parser/parser.go:196`, "Known divergences" in `docs/README.md`). Lifting it means running a namespaced file's statements at include time. Not attempted this run: it reverses a phpscript decision and needs its own design
- limits: `bindTo()`, `call()` and `Closure::fromCallable` scope are not wired; a scope given as an object is read as its class

## 2026-11-04 - a readonly read before initialization raises

- upstream is still at `e40e38b`; `go get ./...` was not run
- reading a readonly property the object does not hold yet answered null. `phpxUninitializedRead` (`runner/phpx_properties.go`, one seam in `helperGet` in the overlaid `runner/helpers.go`) raises php's `Typed property A::$x must not be accessed before initialization` with the declaring class; `isset()` swallows it through the existing `Quiet` mark
- `phpx/testdata/classes/readonly_uninit.phpt`, expected output from `php`
- limit: only readonly properties are uninitialized; a plain typed property without a default still starts as null

## 2026-11-04 - readonly properties

- upstream is still at `e40e38b`; `go get ./...` was not run
- `public readonly int $x;` did not parse, and a promoted `readonly` parameter was a plain property. `model.Field` gains `Readonly`; the parser reads the modifier in a class or trait body, the formatter prints it back, and a `readonly class` or a promoted `readonly` parameter marks the field at hoist (`phpxPromote`)
- a readonly field is absent from the new object (`phpxUninitialized`, one seam in `helpers.go`), so "set" means "initialized". `phpxPropAccess` refuses a second write and a first write from outside the declaring class; `phpxPropUnset` and `phpxIndirectWrite` (a seam in the `Index` case of `execAssign`) cover `unset` and `$o->list[] = 1`. Messages are php 8.4's
- seams: `model/ast.go`, `parser/parser.go`, `formatter/print_decl.go`, `runner/helpers.go`, `runner/runner_assign.go`, `flatstack/engine/program_compiler.go`, all overlaid; untagged stubs in `phpx_properties_stub.go`
- `phpx/testdata/classes/readonly.phpt`, expected output from `php` 8.4; `TestPhpxReadonlyWrite`
- limits: a read before initialization answers null where php throws; an uninitialized property is listed after the initialized ones; php 8.3 words the scope message differently
- found: the bytecode engine now declines every class with a non-public or readonly property, so most OOP code runs on the interpreter; giving `flatHost` the caller class is still the open optimisation

## 2026-11-03 - property visibility

- upstream is still at `e40e38b`; `go get ./...` was not run
- a `protected` or `private` property read, written, unset or listed from anywhere. Hoist now records each one's visibility and declaring class (`phpxBindProps`, `runner/phpx_properties.go`, a table keyed by the hoisted class so an ordinary object costs one map miss), and the access seams (`helperGet`, `execAssign`, `assignTo`, `readLValue`, `execUnset`, the `foreach` over an object) refuse as php does with `Cannot access protected property B::$b`. A private property of an ancestor reads as undefined from a child
- `isset()` and `empty()` swallow the refusal: the expression compiler wraps their arguments (`runner/expr/phpx_quiet.go`, one seam in the overlaid `direct.go`) and an error marked `Quiet` becomes null
- `get_object_vars()` is re-registered in `stdlib/phpx/classes.go` and lists what the calling frame may read
- the bytecode engine reads a property with no caller class, so it declines a class declaring a non-public property (`flatstack/engine/program_compiler.go`, overlaid). Trade-off: most OOP programs now run on the interpreter. A follow-up could give `flatHost` the caller class and lift the decline
- `phpx/testdata/classes/property_visibility.phpt`, expected output from `php`; `TestPhpxPropVerdict`, `TestQuietOnRefusal`
- limits: a write to an ancestor's private property from a child shares the slot; `print_r`, `var_dump`, `json_encode` and `(array)` list every property; `__get`/`__set` unsupported; phpscript has no `??` operator, so there is no seam for it

## 2026-11-03 - constructor property promotion

- upstream is still at `e40e38b`; `go get ./...` changed no dependency
- `__construct(private int $x = 3)` parsed and then did nothing: the object had no `$x`. Hoist now declares each promoted parameter as a property with its visibility and type, and prepends `$this->x = $x;` to the constructor. `phpxPromote` (`runner/phpx_promotion.go`) works on a copy, so the declaration the formatter prints back keeps its modifiers.
- seams: one call in `runner/phpx_inheritance.go`; `runner/runner.go` (overlaid) stopped re-pointing `c.Methods` at the parsed method, which discarded the copy; the bytecode engine declines a class with a promoted parameter (`flatstack/engine/program_compiler.go`, overlaid)
- `phpx/testdata/classes/promotion.phpt`, expected output from `php`: defaults, mixed visibility, `readonly`, a parent constructor chain, an inherited promoted constructor, a body that rewrites the property
- limits: a promoted constructor inside a trait declares no property; named arguments (`new P(y: 9)`) do not parse; `readonly` is not enforced
- found: property visibility is still unenforced, so `get_object_vars($o)` outside the class lists private and protected properties where php lists the public ones. It is the next step.

## 2026-11-03 - the (string) cast refuses an object with no string form

- upstream is still at `e40e38b`; `go get ./...` changed no dependency
- `(string) new C` answered the empty string where php raises `Object of class C could not be converted to string`. `helperCast` returns no error, so `phpxCastString` panics with the `callbackPanic` the call boundary already unwraps, and the expression engine's `PanicError` now goes through `recoveredError` instead of wrapping every panic as a host panic
- seams: `runner/convert.go` and `runner/helpers.go`, both overlaid; untagged answer in `phpx_stringer_stub.go`
- `stringable.phpt` gains the cast case, expected output from `php`

## 2026-11-03 - Stringable and the object string refusal

- upstream is still at `e40e38b`; `go get ./...` was not run
- a class declaring `__toString` was not `Stringable`, and `echo new C` or `"v" . new C` on an object with no `__toString` wrote nothing where php raises `Object of class C could not be converted to string`. `phpxBindStringer` adds `stringable` to `Implements` (inherited and trait-supplied methods too), and `phpxStringOf` sends an object through `phpval.StringContext` in `echo` and concatenation
- seams: `runner/runner.go` (echo) and `runner/runtime_eval.go` (`joinParts`), both overlaid; untagged answer in `phpx_stringer_stub.go`
- `phpx/testdata/classes/stringable.phpt`, expected output from `php`; `TestPhpxStringOf` and the untagged `TestPhpxStringOfUntagged`
- limits: `(string) $obj` goes through `helperCast`, which has no error return, so it still answers the empty string; `class_implements()` is still unbound
- found: untagged `go test ./tests` fails two fixtures that the overlay patched (`global` no-op); they pass under `-tags phpx`, which the gate uses

## 2026-11-03 - shared static properties

- upstream is still at `e40e38b`; `go get ./...` changed no dependency
- each class kept its own copy of an inherited static, so `B::inc()` bumping `static::$n` left `A::$n` alone. php keeps one storage per declaration. Hoist records the declaring class under `MethodOwner["$name"]` for each static a class inherits without redeclaring (`runner/phpx_statics.go`), and `staticStorageFor` routes the read or write to that class's bag. A redeclared static keeps its own
- seams: the call sites in `runner/static.go`, `runner/runner_assign.go` and `runner/flatstack_call.go`, all overlaid; untagged answer in `phpx_statics_stub.go`
- `phpx/testdata/classes/shared_statics.phpt`, expected output from `php`
- found: property visibility and `__call`/`__get` remain unenforced or unsupported

## 2026-11-03 - method visibility

- upstream is still at `e40e38b`; `go get ./...` changed no dependency
- `protected` and `private` methods ran from anywhere. `runner/phpx_visibility.go` now refuses the call as php does, in `helperCall`, `helperStaticCall` and the constructor call of `new`; seams in `runner/helpers.go` and `runner/static.go`, overlaid, with an untagged no-op in `phpx_visibility_stub.go`
- fixed on the way: in `use T { say as private; say as greet; }` the alias took the visibility the earlier rule gave the original. Aliases now apply first, from the method as the trait declared it
- `phpx/testdata/classes/method_visibility.phpt`, expected output from `php`; `TestPhpxCheckVisible`
- limits: property visibility, `__call` fallback and callables are not checked

## 2026-11-03 - extends a built-in throwable

- upstream is still at `e40e38b`; `go get ./...` changed no dependency
- `class AppException extends Exception` failed: `parent::__construct()` raised `static call Exception::__construct(): unknown class`, and an instance had no `getMessage()`, `getCode()` or `getPrevious()`. `new Exception` is a Go value, and a host parent contributed a name and no member
- `runner/phpx_throwable.go` holds php's 22 built-in throwable names with their parents. `extends` on one parses a small php declaration of the members (`message`, `code`, `previous`, the getters, `__toString`), builds it through the ordinary hoist, and registers it and its ancestors in the class table so `parent::` resolves. `ErrorException` gets its own constructor and `getSeverity()`. Any other host parent keeps the old answer
- `catch (Exception)` took a user class by the suffix of its name, so `class Fault extends Error` was caught as an Exception. `phpxIsErrorClass` reads the thrown object's chain; the seam is in `runner/runner_control.go`, overlaid, with the untagged answer in `phpx_inheritance_stub.go`
- `phpx/testdata/classes/extends_throwable.phpt`, expected output from `php`; `TestPhpxThrowableHierarchy` and `TestPhpxIsErrorClass`
- limits: `getFile()` answers an empty string, `getLine()` 0, `getTrace()` an empty array, `getTraceAsString()` `#0 {main}`

## 2026-11-03 - trait adaptations

- upstream is still at `e40e38b`; `go get ./...` was not run
- `use A, B { A::f insteadof B; B::f as g; f as protected; }` was a parse error. The parser reads the block into `model.ClassDecl.TraitRules` and `TraitDecl.TraitRules` (`parser/phpx_traits.go`), the formatter prints it back (`formatter/phpx_traits.go`), and `runner/phpx_traitrules.go` applies it to the trait bodies before they compose
- seams: the `TraitRules` fields in `model/ast.go` and the call in `parser/parser.go`, both overlaid; the type is in fork-owned `model/phpx_traitrules.go`
- `phpx/testdata/traits/adaptations.phpt`, expected output from `php`
- found: method visibility is not enforced at a call (`ACCEPTANCE.md`)

## 2026-11-03 - __CLASS__ and __TRAIT__ in a trait

- upstream is still at `e40e38b`; `go get ./...` was not run
- `__CLASS__` inside a trait method named the trait; it now reads the frame's `__class__`, which hoist set to the using class. `__TRAIT__` is a literal of the trait's name. `get_class()` with no argument answered false and names the running method's class (`stdlib/phpx/classes.go` wraps upstream's binding)
- seams: `parser/parser.go` (a `trait` field) and `parser/expr.go` (two cases), both overlaid; the logic is in fork-owned `parser/phpx_traits.go`
- `phpx/testdata/traits/magic_names.phpt`, expected output from `php`

## 2026-11-03 - __toString

- `echo $obj`, `"x" . $obj`, `strval`, `(string)`, `sprintf('%s')`, `implode` and `str_replace` now call `__toString`, own, inherited or from a trait; upstream gave an object no string form
- `model.Class.Stringer` (overlay) is set at hoist by `runner/phpx_stringer.go`; `internal/phpval` (`phpval.go`, `collection.go`, overlaid) calls it. The runtime reaches phpval through the class, so concurrent runtimes share no global
- `phpx/testdata/classes/tostring.phpt`, expected output from `php`
- upstream `e40e38b` followed first; `go get ./...` changed nothing
- found: a plain object with no `__toString` echoes empty where php raises; `class_implements()` still unregistered (`ACCEPTANCE.md`)

## 2026-11-03 - upstream e40e38b

- follow `e40e38b`, 5 commits ahead of the last one
- `migrations/run.sh` wrote the generated tree again, 1 files dropped

## 2026-11-03 - top-ten package survey and the bindings it named

- upstream is still at `2983cb8`; `go get ./...` was not run
- `phpx/testdata/packages/` lists the ten most downloaded composer libraries with code (psr/log, symfony/console, psr/http-message, guzzlehttp/psr7, psr/container, symfony/finder, guzzlehttp/guzzle, nikic/php-parser, monolog/monolog, doctrine/inflector), and `phpx/scripts/package-survey.sh` lints each one with `phpx lint --include vendor/autoload.php` into `survey.md`
- the lint named 40 undefined functions. Those with at most two arguments and no constant are bound in `stdlib/phpx/`, one file per group; the rest are flagged in `ACCEPTANCE.md` as compatibility issues
- `runner/phpx_classinfo.go` gives `class_parents()` the chain `extends` recorded
- `phpx/testdata/functions/` settles the new names against php (`vendor_basics.phpt`, `clock_arrays.phpt`)
- the widest wall is `??`: 99 of the failing files parse-fail on it, ahead of the namespaced-file statement rule (32) and array spread (19)

## 2026-11-03 - late static binding

- `static::` names the called class and not the class that wrote the body: `new static`, `static::class`, `static::CONST`, `static::method()` and `get_called_class()` answer `Admin` for `Admin::make()` where the inherited body belongs to `Model`. The frame records it as `__static__` beside `__class__`; `self::` and `parent::` forward it, a named class resets it, and a closure made in the method carries it
- `new self` and `new parent` resolve to the declaring class. Upstream refused both and `new static` as won't-build; `tests/fixtures/oop/new_self_refused.phpt` and `new_static_refused.phpt` are dropped through `patches/REMOVED`, and `phpx/testdata/classes/late_static_binding.phpt` settles php's answer
- `static::$prop` keeps resolving to the declaring class's storage, which is php's answer for a property a child inherits; a child that redeclares it shares with its parent here and not in php
- new fork-owned files: `runner/phpx_called.go` and `stdlib/core/phpx_called.go`; the seams sit in `static.go`, `runner_frame.go`, `runtime_eval.go` and `helpers.go`, all overlaid
- found: `fn() => expr` is a parse error, `__TRAIT__` is undefined, `class_parents()` and `strval()` are unregistered, and `echo $obj` with `__toString` prints nothing (a magic method, outside the fork's scope); recorded in `ACCEPTANCE.md`
- `go get ./...` was not run, upstream is still at `2983cb8`

## 2026-11-03 - version_compare()

- upstream is still at `2983cb8`; `go get ./...` was not run, so no dependency changed
- `version_compare()` was unregistered. `stdlib/core/phpx_version_compare.go` registers it with php's canonical form (separators and digit/letter boundaries split parts), the special-form order `dev < alpha < beta < RC < # < pl`, and every operator spelling (`<`, `lt`, `<=`, `le`, `>`, `gt`, `>=`, `ge`, `==`, `eq`, `!=`, `<>`, `ne`)
- `phpx/testdata/constants/version_compare.phpt` settles it against php on all three runtimes
- found while writing the fixture: `foreach ($rows as [$a, $b])` is a parse error, recorded in `ACCEPTANCE.md`

## 2026-11-02 - PHP_VERSION reaches 8.4.27

- upstream is still at `2983cb8`; `go get ./...` changed no dependency
- `PHP_VERSION` read `8.4.0`, so composer's `platform_check.php` refused a vendor tree that requires 8.4.1 or later (`laravel/framework`). `stdlib/phpx/version.go` sets `PHP_VERSION`, `PHP_RELEASE_VERSION`, `PHP_VERSION_ID`, `phpversion()` and `zend_version()` to 8.4.27, the php on this machine
- the binding sits in a new package `stdlib/phpx`, because a file in `stdlib/core` named `phpx_*.go` initialises before `platform.go` and upstream's registration would overwrite it. `stdlib/imports.go` (overlay) imports the package; the untagged build compiles an empty one
- `stdlib/core/platform_test.go` (overlay) pinned `8.4.0`; it now checks the three spellings agree
- `phpx/testdata/constants/php_version.phpt` settles it against php
- found while probing: `version_compare()` is unregistered, recorded in `ACCEPTANCE.md`

## 2026-11-02 - is_a()

- upstream is still at `2983cb8`; `go get ./...` changed no dependency
- `is_a($o, 'A')` raised `call to undefined function is_a()` while `is_subclass_of()` existed. `stdlib/core/phpx_isa.go` registers it over `Runtime.PhpxIsA` (`runner/phpx_isa.go`), which reuses `instanceof` for an object and reads a class name's ancestors and interfaces when `$allow_string` holds
- `phpx/testdata/classes/is_a.phpt` settles object, name, case, leading backslash, interface, scalar and unknown-class cases against php on all three runtimes
- the `is_a` item leaves "Discovered, not implemented" in `ACCEPTANCE.md`

## 2026-11-02 - a class name is declared once

- upstream is still at `2983cb8`; `go get ./...` changed no dependency; the previous run's commit `b938ab9` is pushed now
- `phpxClassChain` refuses a second declaration of one class name in a program, and a class named `stdClass`, with php's `Cannot redeclare class A`; the check reads the `decls` index hoist already builds, so no upstream file changes
- `phpx/testdata/classes/` holds the two fixtures; `runner/phpx_inheritance_test.go` covers the seam
- limit: php appends `(previously declared in FILE:LINE)`, and two declarations in different included files pass

## 2026-11-02 - self:: inside a constant expression

- upstream is still at `2983cb8`, the commit the last follow landed on, so there was no follow to absorb; `go get ./...` changed no dependency
- `const B = self::A + 1` raised `class constant self::A: unknown class`: `helperClassConst` evaluated a constant's expression in the scope that asked for it, which carries no class. It now evaluates in the declaring class's scope (`classDefaultScope`, the one a static default already ran in), and an interface constant in a scope named for the interface. `runner/runtime_eval.go` is an upstream file, so the whole-file copy joins `phpx/testdata/patches/runner/`
- `phpx/testdata/constants/self_in_constant.phpt` settles it against php on all three runtimes: `self::` and `K::` in a class constant, a static default, and an interface constant read through the interface and through an implementing class
- found while probing: this was the wall behind the `Illuminate` constant tables the framework probes will meet next, and it closes the discovery item in `ACCEPTANCE.md`
- limit: an inherited constant evaluates `self::` as the inheriting class, because `model.Class.Consts` keeps no declaring class; recorded under the new acceptance box

## 2026-11-02 - interfaces: the contract meets the finished class

- `implements` is checked against the class the fork hoists, not the class's own AST: a method a parent carried, a trait contributed, or an autoloaded parent brought satisfies the contract, which is how php reads the finished class. `runner/phpx_interfaces.go` moves the check to the end of `hoist`; the untagged build keeps upstream's AST check in the stub beside it
- `model/phpx_interfaces.go` adds the effective method set the linter needs and `PhpxCheckInterfaces`; `lint/lint.go` joins the overlay so `phpx lint` reads the same set and stops reporting `class B extends A implements I` as missing `f()`, and a class whose parent the file does not declare is left to the runtime
- an interface contributes its constants: `Store::LIMIT` through the implementing class and `B::LIMIT` through an interface that extends the declarer, with a class constant of the same name winning; `runner/phpx_interfaces.go` folds them into the class and into each interface declaration at hoist, and reads an interface the runtime registered from an earlier included file, which is the one-interface-per-file layout a framework uses
- a class that includes the interface's file itself (the file hoists before its include statements run), or that leaves the interface to the autoloader, folds no constant: the runtime must know the interface when the class is built
- the bytecode engine's contract check moves after its decline loop, so a program that declares `extends` or a trait is declined instead of being refused by a check that cannot see the parent; the fixture's Flat stack column skips
- `tests/fixtures/oop/interface_constants_not_acquired.phpt` and `interface_constants_extends.phpt` drop through `phpx/testdata/patches/REMOVED`: their own descriptions name php's answer, which is now the fork's
- `phpx/testdata/interfaces/interfaces.phpt` settles the whole behaviour against php on the matrix; `runner/phpx_interfaces_test.go` covers the seam and `runner/phpx_interfaces_stub_test.go` pins the untagged build
- discovered and not implemented: a constant expression that names `self::` does not resolve (`class constant self::A: unknown class` where php answers 2)
- `go get ./...` changed no dependency

## 2026-11-02 - upstream 2983cb8

- follow `2983cb8`, 4 commits ahead of the last one
- `migrations/run.sh` wrote the generated tree again, 1 file dropped
- the overlay copies of `model/ast.go`, `parser/parser.go`, `lint/names.go`, `formatter/print_decl.go` and `flatstack/engine/program_compiler.go` take upstream's hunks: `#194` adds `PropAccess.NameExpr` and has the bytecode engine decline a dynamic property name, `#191` keeps the leading backslash on a name equal to the namespace, and a whole-file copy that misses either drops the change on the floor; the build caught the missing field and the fixture tray caught the engine compiling a name it must decline
- upstream `#194` implements `$o->{'name'}` and `$o->$name`, which drops the discovery that waited on an issue; `#192` reads an out-parameter live and `#193` lists class and interface methods with `--api`
- `go get ./...` changed no dependency

## 2026-11-02 - autoloaded parents, typed property types, frameworks

- upstream is still at `672379e`, the commit the last follow landed on, and `go get ./...` changed no dependency
- `extends` resolves a parent no declaration in the program answers through the autoload queue. composer registers its loader with `--include vendor/autoload.php`, and `runner/phpx_inheritance.go` calls `rt.autoload` after the declaration list and the class table; `runner/phpx_autoload_test.go` covers the resolved parent and the unchanged missing-parent fatal
- a property type may open with `\` (`private \stdClass $providers;`), and `parser/parser.go` accepts it where a property type starts; `phpx/testdata/syntax/typed_property.phpt` settles a nullable, a union and a static typed property against php on all three runtimes, and `phpx/testdata/patches/parser/parser.go` carries the overlay copy
- the four framework snippets land under `phpx/testdata/frameworks/`, each a `composer.json` and its entry program, and `phpx/testdata/report.md` records the first wall and the walls behind it for each
- the walls the probes met, all outside the five reversals: a namespaced file with a top-level statement (composer's `nikic/fast-route` autoload file), a function declared inside a top-level `if` (the `autoload_files` guard), `??`/`??=` and `fn($x) => ...`, and `PHP_VERSION` reading `8.4.0` against a vendor tree that requires `8.4.1`
- `atkins.phpx.yml` skips `phpx/testdata/frameworks/` in the fixture loop, because a blocked framework holds no `.phpt` and the runner treats an empty tree as an error; a framework that runs adds its fixture and drops the skip
- `README.phpx.md` ticks `phpx/testdata/report.md`, and `ACCEPTANCE.md` gains an "OOP syntax" section with the typed-property box and two discovered items

## 2026-11-02 - stdClass

- `json_decode` answers stdClass for a JSON object, which is PHP's default shape and the one framework code reads with `->name`; upstream built an `*model.Array` and refused `$associative = false`
- `stdlib/core/phpx_json.go` re-registers the name from a fork-owned file: the bindings apply in contribution order, `phpx_json.go` sorts after `json.go`, and the untagged build keeps upstream's decoder
- `$associative = true` keeps the array shape; `false`, `null` and an omitted argument answer stdClass; a nested object follows at any depth; the document's order survives a decode and re-encode; a decoded object is an ordinary property bag (instanceof, foreach, unset, dynamic property, `(array)` cast)
- the extension `JSON\Decoder` is left as upstream built it, because it has no PHP counterpart and the fork changes the PHP name alone
- `phpx/testdata/stdclass/decode.phpt` settles it against php on all three runtimes: the php column passes, and the bytecode engine runs the program rather than skipping it
- `stdlib/core/phpx_json_test.go` covers the same shape from the binding under `-tags phpx`
- `phpx/testdata/README.md` moves the `stdclass/` row to done and `README.phpx.md` ticks the fifth reversal
- three gaps the reversal met and did not take, recorded under "Discovered, not implemented" in `ACCEPTANCE.md`: `is_a()` is unregistered where `is_subclass_of()` exists, `$o->{'name'}` is a parse error where PHP allows a braced property name, and a class, interface or trait named `stdClass` is accepted where php refuses to redeclare it
- `go get ./...` changed no dependency

## 2026-11-02 - abstract

- `abstract class` lands: `runner/phpx_abstract.go` refuses `new` on a class that declared the modifier, and `model.Class.Abstract` carries the flag the runtime reads
- hoist refuses a concrete class that leaves a method to a child: the class's own abstract declarations first, then what a parent left, then what a trait contributed, which is php's order and php's message (`Class C contains 2 abstract methods and must therefore be declared abstract or implement the remaining methods (B::b2, A::a1)`)
- a child that implements every method instantiates and inherits the concrete methods; an abstract class with no abstract method instantiates through a concrete child; a trait's abstract method is satisfied by the using class or by a concrete method the class inherits, and the inherited body is the one that runs
- a class may not turn a method its parent implements back into a declaration without a body: php's `Cannot make non abstract method P::f() abstract in class C`
- a class's own abstract declaration is reported before traits bind and a parent's members arrive, which is php's compile-time order, and a private method without a body is refused (`Abstract function A::f() cannot be declared private`); a trait may declare a private abstract method, because the using class supplies the body
- `parent::f()` on an abstract method is php's `Cannot call abstract method A::f()`; the message names the declaring class, which `phpxDeclOwner` reads from the class table rather than from the instance
- the bytecode engine declines a program that declares an abstract class or an abstract method, so the whole program runs on the interpreter
- the linter stops reporting `abstract` as a no-op, in `lint/walk.go` and its test
- `phpx/testdata/abstract/` settles it against php: one happy-path fixture and four `error:` fixtures for the four fatals, with the php column skipped on the error fixtures because the host renders an uncaught error where php writes stderr
- the parser already recorded `abstract` on the class and the method and the formatter printed both back, so this reversal needed no overlay syntax half
- `go get ./...` changed no dependency

## 2026-11-02 - traits

- `trait Name { ... }` parses into `model.TraitDecl`, and `use A, B;` inside a class or trait body records `model.ClassDecl.Traits`; the formatter prints both back, so a trait file lints and reformats in either build
- `runner/phpx_traits.go` folds a trait's properties, statics, constants and methods into each class that names it, nested `use` lines included, and it holds the collision bookkeeping and php's three messages
- a class's own method wins over the trait's, which is php's rule; a class's own property or constant is composed with the trait's and must agree on visibility, type and default, or php's fatal names both
- two traits declaring one name compose only when their definitions agree: a shared method is always a collision, a property or constant collides only when the definitions differ, and the three messages are php's word for word (`Trait method B::foo has not been applied as C::foo`, `A and B define the same property ($x) in the composition of C`)
- a trait contributes no ancestor: `instanceof TraitName` is false, `method_exists` finds what the trait installed, and `self::` in a trait method resolves to the using class
- the bytecode engine declines a program that declares or uses a trait, so the whole program runs on the interpreter
- `phpx/testdata/traits/` settles it against php: one happy-path fixture and two `error:` fixtures for the method and property fatals, with the php column skipped on the latter because the host renders an uncaught error where php writes stderr
- limits recorded in `README.phpx.md`: `insteadof`/`as` adaptations are a parse error, a trait declared in one included file is not seen by a class in another, `__CLASS__` in a trait method answers the trait's name, and an abstract trait method the class leaves unimplemented is accepted until the abstract reversal lands
- the linter walks trait bodies and counts a trait declaration as a declared name
- the syntax half lives in the overlay, because a follow lands upstream's file verbatim: `parser/parser.go`, `model/ast.go`, the three `formatter` files and `lint/names.go` join `phpx/testdata/patches/`, and the untagged build parses a trait while taking nothing from it
- `go get ./...` changed no dependency

## 2026-11-02 - inheritance

- `extends` lands: `runner/phpx_inheritance.go` flattens the parent's fields, statics, constants and methods into the child at hoist, the child's own declaration of a name wins, and `model.Class` grows no parent field
- `parent::method()` and `parent::__construct()` resolve against the neighbour in the chain and keep `$this` bound; a closure declared in a method carries the same parent
- `self::` inside an inherited body resolves to the class that declared it (`model.Class.MethodOwner`), which is php's rule and not the called class
- `instanceof`, `get_parent_class`, `is_subclass_of` and a catch clause naming an ancestor read `model.Class.Ancestors`
- a parent that is a host class (`extends Exception`) contributes its name and no member; `parent::__construct()` on one fails, and a static property keeps one bag per class rather than php's shared storage
- the bytecode engine declines a program that declares `extends`, so the whole program runs on the interpreter
- the linter stops reporting `extends`; upstream's `oop/extends_noop.phpt` and `oop/parent_refused.phpt` are dropped through `patches/REMOVED`, and `phpx/testdata/inheritance` settles the behaviour against php
- `go get ./...` changed no dependency
- `migrations/004-atkins-state.sh` adds `.atkins/atkins.db*` to `.gitignore`: atkins' run database and its SQLite side files are run state, and a pipeline run left a tree `follow.sh` refuses as dirty

## 2026-11-02 - upstream 672379e

- fork of `titpetric/phpscript` at `672379e`, with `origin` on `titpetric/phpx`
- `README.phpx.md` holds the standing prompt, and `README.md` is the copy of it
- `atkins.phpx.yml` holds the fork pipeline, and `phpx/scripts/follow.sh` the one-way follow
- `migrations/` holds the rewrites, one step per file, run again on every follow
- `migrations/001-module-path.sh` renamed the module to `github.com/titpetric/phpx` over 446 files
- `lefthook.yml` refuses a `README.md` staged by hand and a push where the copy drifted
- `atkins.phpx.yml` carries a counterpart for all 24 upstream jobs, with splint measured as a delta
- `ACCEPTANCE.md` holds the checklist, one box per claim and the command that settles it
- Docker Hub serves `titpetric/phpx:latest`, with the git stamp and `phpscript` linked to `phpx` inside it
- `phpx/testdata/patches/` holds the fork's version of an upstream file, copied over the tree on every follow
- the first reversal landed: `global` binds a name to the global frame, settled against php
- the five reversals wait for a worker: `global`, `extends`, traits, `abstract class`, `stdClass`
