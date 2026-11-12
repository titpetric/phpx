# phpx testing report

What has been tested against php, what a green row means, and what the numbers say to a team that wants to run
its own code on phpx. Last measured 2026-11-05 against `php 8.4.27` as the reference, on this checkout with
`-tags phpx`. Everything below regenerates: the commands sit beside each table.

## Verdict

phpx is **not yet a drop-in runtime for a Laravel or Filament application**, and it is a good fit for libraries.
The five applications measured here parse almost completely (99.8% to 100% of their files) and three of them load
every class under php and phpx alike, but none of them boots a request: each stops at a named wall, listed below,
and the walls are PHP 8.4 syntax and built-in classes and functions, not gaps in the OOP baseline.

| Question | Answer |
|----------|--------|
| Can I run a composer library on phpx? | Yes for the nine measured: 120 of 120 tests, 80% statement coverage on 8 of 9 (`symfony/console` is 64%). |
| Can I run a framework's own code? | Yes for what the shims reach: Slim, Latte, Laravel, Filament run their `launcher.php tests/` with php's totals. |
| Can I load a Laravel/Filament app's classes? | Mostly: 1,263 of the 1,280 classes php loads (98.7%) load under phpx on four apps; the fifth loads none. |
| Can I boot one and answer a request? | No. 0 of 5 reach `GET /`. Four stop in the vendor tree, one in the autoloader. |
| What is the next blocker for the most apps? | PHP 8.4 property hooks in `symfony/http-foundation` 7.4+ (every Laravel 13 app). |

## What a pass means

| Row says | It means | It does not mean |
|----------|----------|------------------|
| a `.phpt` fixture passes on `php`, `runtime`, `flatstack` | the three columns print what php 8.4 printed for that program (the expected output is php's, not phpx's) | the feature is complete; a fixture pins the cases someone thought of |
| a library test passes | the library's code path ran to the same assertion result under phpx as under php | every path of the library runs; see its statement coverage |
| a library's coverage is 80%+ | 80% of the library's statements ran while those tests ran, files no test reached counted as uncovered | the tests are good, or the lines are right: a line that ran and returned a wrong value counts as covered |
| a file `parses` (lint PASS or WARN) | phpx's parser built an AST for it | it runs: a file that parses can still call an undefined function or a missing class |
| a class `loads` | autoload found the file, parsed it, and every parent, interface and trait it names resolved with php's composition rules | its methods work; loading executes declarations and constant expressions, not method bodies |
| an app `boots` to stage N | the artisan bootstrap got through stage N with php's behaviour at every call so far | stage N+1 works |
| app statements run | statements of `app/` files that executed while phpx booted, out of the statements of the `app/` files loaded by then | a coverage of the app: the denominator is only the files reached |

## Measuring real applications

Five open-source apps, cloned at the commit shown, installed with composer. The scripts are in
[testdata/apps/](testdata/apps/README.md); `testdata/apps/run.sh <phpx> <apps> <apps-prod>` prints every table below.

| App | Stack | Commit | Framework | Why it is here |
|-----|-------|--------|-----------|----------------|
| [`relaticle/relaticle`](https://github.com/relaticle/relaticle) | Laravel 13, Filament 5, Livewire | `da2f5fd` | `laravel/framework` v13.40 | the most current Filament app: an open CRM |
| [`andrewdwallo/erpsaas`](https://github.com/andrewdwallo/erpsaas) | Laravel 12, Filament 3 | `565501e` | v12.25 | accounting platform on Filament panels |
| [`devaslanphp/project-management`](https://github.com/devaslanphp/project-management) | Laravel 9, Filament 2 | `5498216` | v9.45 | a Filament 2 admin, an older and larger-installed shape |
| [`realodix/urlhub`](https://github.com/realodix/urlhub) | Laravel 13, Fortify | `044bc14` | v13.40 | a small plain Laravel app, no Filament |
| [`akaunting/akaunting`](https://github.com/akaunting/akaunting) | Laravel 10, modules | `8c4a5e5` | v10.50 | a large plain Laravel app: 1,085 classes |

Five tests, the same for each app. They run in order of what they cost to pass.

| # | Consumer test | Command | Pass means |
|--:|---------------|---------|-----------|
| 1 | **Parse**: lint the app and its framework trees | `testdata/apps/survey.sh` | no file fails to parse |
| 2 | **Class load**: `class_exists` every class of `app/`, one process each | `testdata/apps/load.sh` | every class php loads, phpx loads |
| 3 | **Boot**: `bootstrap/app.php`, the console kernel, providers booted | `testdata/apps/boot.sh` | stage 6, the version string |
| 4 | **First request**: `GET /` through the HTTP kernel | stage 9 of `boot.php` | phpx answers the status php answers |
| 5 | **Measure**: statements of `app/` that ran under `--cover` | `boot.sh` | reported, not gated |

The runs of tests 2 to 5 use `composer install --no-dev`, which is what ships; test 1 uses the dev install so
`tests/` is present.

### Test 1: parse

Files that parse, as `passing + warnings` out of the files in the tree. A warning is a lint finding (an unknown
class, an unused variable), not a parse failure. `app/` is the app's own code; the vendor rows show the walls it
inherits.

| App | `app/` | `tests/` | `laravel/framework` | `filament` | `livewire` | `symfony` |
|-----|-------:|---------:|--------------------:|-----------:|-----------:|----------:|
| `relaticle` | 739 of 740 | 2,828 of 2,828 | 2,443 of 2,443 | 5,099 of 5,099 | 348 of 348 | 3,720 of 3,728 |
| `erpsaas` | 425 of 425 | 24 of 24 | 1,960 of 1,960 | 3,369 of 3,369 | 277 of 277 | 3,886 of 3,889 |
| `project-management` | 152 of 152 | 4 of 4 | 1,590 of 1,590 | 1,794 of 1,794 | 148 of 148 | 1,955 of 1,957 |
| `urlhub` | 52 of 52 | 72 of 72 | 2,489 of 2,489 | 2,691 of 2,691 | 345 of 345 | 3,203 of 3,209 |
| `akaunting` | 2,263 of 2,267 | 104 of 104 | 1,735 of 1,735 | n/a | 423 of 423 | 2,795 of 2,799 |

Verdict: **parse is not the blocker.** 28 failing files in the five apps (a vendor file counts once per app that
installs it), in six causes, each one construct:

| Parse failure | Files | Where | Language feature |
|---------------|------:|-------|------------------|
| `public ParameterBag $attributes { ... }` | 4 | `symfony/http-foundation` `Request`, `Response` (relaticle, urlhub) | PHP 8.4 property hooks |
| `public public(set) readonly array $v` | 4 | `symfony/http-kernel` `Cache`, `ViewEvent` (relaticle, urlhub) | PHP 8.4 asymmetric visibility |
| `isset(self::${$property}[$u])`, `static::${$name}` | 5 | `symfony/error-handler` `DebugClassLoader` (every app), `spatie/invade` | variable static property name |
| `{$this->indent(...)}` inside a nowdoc/heredoc closed by `EOF;` | 5 | `symfony/routing` `CompiledUrlMatcherDumper` (every app) | heredoc interpolation of a method call |
| `yield $r => $chunk = new X`, `(int) $ch = $info['h']` | 4 | `symfony/http-client` `AsyncResponse`, `CurlResponse` | assignment as an operand after `=>` and after a cast |
| `new readonly class(...) implements I` | 1 | `relaticle` `ConfirmIdentityAction` | `readonly` anonymous class |
| a statement starting with the token `Documents` in a class body | 4 | `akaunting` `app/BulkActions/Sales/*` | not traced |
| `??=` where an expression is expected | 1 | `nette/utils` `compatibility.php` | not traced |

### Test 2: class load

| App | Classes in `app/` | php loads | phpx loads | Both load | phpx / php | What stops the rest |
|-----|-----------------:|----------:|-----------:|----------:|-----------:|---------------------|
| `erpsaas` | 399 | 399 | 399 | 399 | **100%** | |
| `project-management` | 151 | 151 | 151 | 151 | **100%** | |
| `relaticle` | 691 | 679 | 675 | 663 | 97.6% | 15 classes extend `symfony/http-foundation` (property hooks); 1 `new readonly class`. php loads 679 of 691: 12 fatal on php 8.4 itself, because the app requires 8.5 |
| `urlhub` | 51 | 51 | 50 | 50 | 98.0% | `StoreUrlRequest` extends `FormRequest`, which loads `Request` (property hooks) |
| `akaunting` | 1,085 | 1,076 | 0 | 0 | **0%** | composer's autoload files call `stream_filter_register()` (`php-http/message`), which phpx does not bind, so `vendor/autoload.php` stops before any class |

Verdict: for two apps class loading **passes** in full and for two more it passes but for classes behind one parser wall, which means the class model (inheritance, traits, interfaces,
enums, attributes, promoted constructors) is sound at app scale, including 399 Filament classes. Three causes account for every
failure: property hooks, `new readonly class`, and the autoload files of one package.

### Tests 3 to 5: boot, first request, measure

Stages reached (1 autoload, 2 application, 3 kernel, 4 bootstrap, 5 container, 6 version, 7 routes, 8 request,
9 response). php reaches 9 for four of the apps. `relaticle` stops at 4 under php 8.4 because it needs 8.5.

| App | php reached | phpx reached | phpx stopped at | What it is | `app/` statements run |
|-----|-------------|--------------|-----------------|------------|----------------------:|
| `urlhub` | 9 (HTTP 500: no database) | 4 | `PDO::FETCH_OBJ`: unknown class | PDO is an extension | 21 of 62 |
| `relaticle` | 4 (php 8.4 fatal: trait property `$badgeColor` incompatible) | 9 (HTTP 301) | none: reaches a response where php cannot boot | php is stricter about trait property composition than phpx | 519 of 3,431 |
| `project-management` | 9 (HTTP 302) | 4 | `PDO::FETCH_OBJ`: unknown class | PDO is an extension | 34 of 1,087 |
| `erpsaas` | 9 (HTTP 302) | 9 (HTTP 500) | the 500 is `openssl_cipher_iv_length()` undefined, from `Illuminate\Encryption` (cookie encryption): an extension wall, out of scope | not yet read | 145 of 2,949 |
| `akaunting` | 9 (HTTP 302) | 1 | `stream_filter_register()` undefined | stream filters | 0 of 0 |

Verdict: **0 of 5 answer a request.** Reading across the table, the blockers by the number of apps they stop:

| Rank | Blocker | Apps stopped | In the fork's scope? | Size of the fix |
|-----:|---------|-------------:|----------------------|-----------------|
| 1 | PHP 8.4 property hooks and `public(set)` (`symfony/http-foundation` 7.4+) | 2 of 5 (and any Laravel 13 app) | yes: class syntax | a parser and runtime feature (get/set bodies, backed properties, the visibility pair) |
| 2 | PDO (`PDO::FETCH_OBJ`, the Eloquent connection) | 1 of 5 seen so far, all of them behind it | no: extension | an app that touches a database cannot boot without it |
| 3 | `DatePeriod`, `DateTime`-family built-ins | 1 of 5 seen so far | no: standard library, but Carbon is in every Laravel app | bind the class family |
| 4 | `stream_filter_register` and the stream filter functions | 1 of 5 | no: standard library | bind or stub |

The "seen so far" is the point of the ladder: a boot test stops at the first wall, so the counts above are a
floor. Each fix moves the stage and uncovers the next wall.

The measure column is deliberately modest. `app/` statements run is the part of the app that executed before the
wall; it is not a coverage figure for the app and must not be quoted as one.

### Fixed by this measurement

The first run of these tests stopped on bugs, not walls. Four were fixed, each with a fixture whose expected
output is php's:

| Bug | Seen in | Fixture |
|-----|---------|---------|
| `\false`, `\true`, `\null` read as an unknown constant | `google/apiclient` aliases (relaticle stopped at stage 1) | `testdata/syntax/qualified_constants.phpt` |
| a trait reached by two paths (class uses `T` and `U`, `U` uses `T`) raised "has not been applied ... collision" | Livewire's `WithFileUploads` and `WithRateLimiting`, 22 classes | `testdata/traits/shared_trait.phpt` |
| `IteratorIterator` did not forward unknown methods to its inner iterator | `symfony/finder`'s `ExcludeDirectoryFilterIterator::getFilename()`, stage 4 of two apps | `testdata/classes/iterator_iterator_forwards.phpt` |
| a protected method called from outside did not go to `__call`/`__callStatic` | Laravel's `Router::prefix()` through the `Route` facade, stage 4 of two apps | `testdata/classes/magic_call_inaccessible.phpt` |

## What the measurement means for someone who wants to integrate

| If you are... | The state today | What to do |
|---------------|-----------------|------------|
| shipping a library (PSR, Guzzle, Monolog, Symfony components) | supported: nine libraries run their own tests with php's results | run the library's phpunit suite under phpx; a difference in totals is a bug, file it with the fixture |
| embedding PHP business logic in a Go service (value objects, validation, rules) | supported for plain PHP and the OOP baseline | write the logic against the fixtures' feature list; avoid `eval`, references, `__destruct`, scalar type coercion |
| running a Laravel or Filament application | not yet: class loading works for most apps, booting does not | track the four walls above; the class-load test is the cheap gate to re-run after each |
| depending on PDO, DOM, libxml, mbstring or Carbon's date classes | out of scope for the fork | expect an issue before a box |
| wanting a percentage to quote | quote the class-load rate and the stage reached, with the app and commit | do not quote statement coverage of an app: it is not measured until a request completes |

How to read the three measures:

| Measure | What it tells you | What it hides |
|---------|-------------------|---------------|
| Parse rate | the language is covered; a failing file names the missing syntax | a file can parse and still not run |
| Class load rate | the object model composes the way php composes it, at 100s of classes | method bodies, calls, extension use |
| Stage reached | the first thing that stops a real program | what is behind the wall; the stage is a floor |

The measures that are not available yet, in order of value: a full request through an app (needs the walls above
and a database); `app/` statement coverage from the app's own test suite (needs the same, plus phpunit or Pest
under phpx, which are measured below for phpunit only); an OOP benchmark against phpscript (not built).

## Frameworks, libraries and the language

### Framework shims

Each shim is a small phpunit suite over the framework's public surface, in
[testdata/frameworks/](testdata/frameworks/). `cd phpx/testdata/frameworks/<name> && phpx run launcher.php tests/`.

| Framework | Version | Tests | php | phpx | Verdict |
|-----------|---------|------:|-----|------|---------|
| Slim | 4.15.3 | 5 | `OK (5 tests, 9 assertions)` | same | pass |
| Latte | 3.1.7 | 5 | `OK (5 tests, 9 assertions)` | same | pass |
| Laravel | 12.70.2 | 4 | `OK (4 tests, 11 assertions)` | same | pass: container, collections, support |
| Filament | on Laravel | 2 | `OK (2 tests, 6 assertions)` | same | pass: not a panel booted, see the app tests |

The shims prove the framework code runs. The app tests above prove it does not yet run as an application: Laravel's
12.70 shim passes while two Laravel 12 and 13 apps stop at stage 3 and 4.

### Libraries

| Library | Tests | php | phpx | Statements | Covered | Gate (80%) | Blocker |
|---------|------:|----:|-----:|-----------:|--------:|:----------:|---------|
| `psr/log` | 10 | 10 | 10 | 11 | 100.0% | pass | |
| `psr/http-message` | 9 | 9 | 9 | 0 | interfaces | pass | |
| `psr/container` | 8 | 8 | 8 | 1 | 100.0% | pass | |
| `guzzlehttp/psr7` | 15 | 15 | 15 | 1,903 | 84.0% | pass | |
| `guzzlehttp/guzzle` | 16 | 16 | 16 | 3,722 | 87.6% | pass | |
| `doctrine/inflector` | 17 | 17 | 17 | 740 | 99.5% | pass | |
| `symfony/finder` | 12 | 12 | 12 | 606 | 93.4% | pass | |
| `symfony/console` | 17 | 17 | 17 | 4,925 | 64.3% | **fail** | ceiling 88%: process, stopwatch, lock, pcntl, a tty; same-named private properties, unenforced scalar types, DOMDocument |
| `monolog/monolog` | 16 | 16 | 16 | 3,317 | 85.0% | pass | |

Passing here means phpx gives php's result on tests written against the library's surface. 1,111 tests run in
the coverage pass and phpx fails about 180 of them; those are the by-design blockers (`__destruct`, `__serialize`,
`&` references outside `foreach`, scalar parameter types) and are kept as regressions.

### Upstream suites

A library's own test suite is the widest check of php's behaviour the repository can run.

| Library | php | phpx | Verdict |
|---------|-----|------|---------|
| pimple | 68 tests, 135 assertions | same | pass |
| `doctrine/inflector` | 1,215 tests, 1,238 assertions | same | pass |
| `egulias/email-validator` | 358 tests, 477 assertions | 472 assertions, 5 errors | the intl `Spoofchecker` extension |

### Fixtures and the phpunit runner

| Suite | Count | Result |
|-------|------:|--------|
| `phpx test --matrix ./tests/fixtures/...` (upstream) | 298 | 298 pass on flatstack, runtime and php |
| `phpx test --matrix ./phpx/testdata/...` (the fork's) | 192 | 191 pass; the one failure is `patches/tests/fixtures/namespaces/instanceof_namespace.phpt`, an overlay copy that needs a sibling directory it does not have where it sits, and it passes in its place under `tests/fixtures` |
| `go test ./...`, `go test -tags phpx ./...` | all packages | pass |
| phpunit 11.5 from its source tree, `compare.sh` | 9 files | `same` for each; failures that php reports too are expected |

Open in phpunit: `createMock()` needs `eval`; a throwable records no file and line; `serialize` is process isolation
only. The `phpx phpunit` subcommand does not exist; `phpx/scripts/phpunit.sh` stands in for it.

### Go coverage of the fork

86% of the fork-owned statements (3,502 in 76 files) run under `go test` and the fixtures together; the whole module
is 88%. Per-file numbers are in the appendix and per-unit numbers in [MEASUREMENTS.md](MEASUREMENTS.md). This is
coverage of phpx's Go, which says the fork's code is exercised; it says nothing about how much PHP a program can
run, which is what the tables above measure.

## Open

| Item | Why it matters | Where |
|------|----------------|-------|
| property hooks, `public(set)` | rank 1 blocker for current Laravel and Symfony | parser, runtime |
| `stream_filter_register` family | akaunting stops at stage 1 | `stdlib` |
| PDO, `DatePeriod` | rank 2 and 3 | out of scope or standard library |
| trace the 5 unexplained parse failures (`Documents` x4, `??=` x1) | each is one construct | `phpx lint` on the file |
| a database for the boot test | stage 9 needs a connection to answer 200 rather than 500 | `testdata/apps/boot.php` |
| the apps' own test suites under phpx | the real coverage measure | needs Pest and phpunit with `createMock` |
| the OOP benchmark, phpscript against phpx | listed in the standing prompt, still not built | |

## Appendix: the 2026-11-04 run

The run behind this report took the fork from "phpunit stops on a parse error at `Application.php:835`" to
"phpunit 11.5 runs from its source tree: test discovery, `setUp`, assertions, `#[DataProvider]`, `#[TestWith]`,
`#[Depends]`, exceptions, output expectations, skips and incompletes, and it prints what php prints". Everything
below was measured on this checkout with `php 8.4.27` as the reference; the expected output of every new fixture is
`php`'s own, not phpx's.

### Before and after

| Measure | Before (`e1e250e`) | After |
|---------|-------------------:|------:|
| phpx fixtures (`phpx test --matrix phpx/testdata/...`: flatstack, runtime and php) | 47 of 47 | **81 of 81** |
| upstream fixtures (`phpx test --matrix tests/fixtures/...`) | 290 of 290 | **287 of 287** (3 dropped, see below) |
| `go test -tags phpx ./...` | green | green |
| `phpunit/phpunit/src`, files that parse (pass / warn / fail) | 795 / 483 / **25** | 895 / 93 / **0** |
| whole phpunit install (phpunit, sebastian, phar-io, myclabs, staabm, theseer) | 1025 / 746 / **36** | 1173 / 196 / **0** |
| top ten composer libraries, failing files | **71** | **18** |
| top ten composer libraries, passing files | 438 | 536 |
| `phpx run launcher.php tests/ATest.php` | parse error, nothing runs | `OK (1 test, 1 assertion)` |
| phpunit test files whose output equals php's (`compare.sh`) | 0 of 5 | **5 of 5** (the failure's file and line aside) |
| phpunit tests run, all five files | 0 | 31 tests, 87 assertions, 2 failures and 1 error that php reports too |
| undefined functions called by the phpunit tree | `assert` x318, `serialize`, `current`, `class_implements`, ... | extension functions only (`libxml_get_errors`, `proc_*`, `xdebug_*`, `pcntl_*`), `serialize`, `unserialize`, `eval` |

The ten libraries, per package (passing / warning / failing files):

| Package | Before | After |
|---------|-------:|------:|
| `guzzlehttp/psr7` | 11 / 49 / 4 | 13 / 55 / 0 |
| `guzzlehttp/guzzle` | 27 / 74 / 7 | 31 / 102 / 1 |
| `symfony/console` | 55 / 133 / 29 | 76 / 161 / 15 |
| `symfony/finder` | 14 / 9 / 4 | 16 / 15 / 2 |
| `monolog/monolog` | 35 / 305 / 16 | 75 / 172 / 0 |
| `doctrine/inflector` | 47 / 2 / 1 | 49 / 0 / 0 |
| `psr/*` (3 packages) | all pass | all pass |

Three upstream fixtures and six Go tests pinned decisions this fork reverses and are dropped or reworded through
`phpx/testdata/patches/` (`REMOVED`, and overlaid test files). Each has a fixture under `phpx/testdata/` that settles the
new answer against php:

| Dropped or reworded | The decision it pinned | Where the new answer is settled |
|---------------------|------------------------|---------------------------------|
| `tests/fixtures/arrays/unset_native_slice_reindexes.phpt`, two cases in `tests/bindings_test.go` | a Go slice a binding answers is not a script array | `functions/explode_array.phpt` |
| `tests/fixtures/stdlib/json_encode_flags.phpt`, `TestNoJSONFlagConstants`, `TestJSONFlagConstantThrows` | no `JSON_*` constant | `functions/json_constants_test.php` |
| `tests/fixtures/runtime/memory_alias.phpt`, `TestAliasCountedOnce` | `$b = $a` shares one array, counted once | `classes/array_value_semantics_test.php` |
| `TestFOpenUnimplementedWrapper` (asks about `php://filter` now) | `php://memory` answers false | `functions/php_memory_stream_test.php` |

### What landed

Syntax, each with a fixture whose expected output is php's:

| Feature | Seam | Fixture |
|---------|------|---------|
| `[$a, [$b, $c], 'k' => $d] = ...`, `foreach (... as [$a, $b])`, skipped slots | `parser/phpx_destructure.go` | `functions/destructuring_test.php` |
| `<<<EOT`, `<<<'EOT'`, closing-line indentation | `parser/phpx_heredoc.go` | `syntax/heredoc_test.php` |
| `and`, `or`, `xor`; `for (;;)`, `for ($i = 0, $c = n; ...; $i++, $j--)`; `a ?? b = c` | `parser/phpx_logical.go` | `syntax/word_operators_test.php` |
| assignment expressions on index, property and static targets, compound operators | `runner/expr/phpx_assign.go`, `runner/phpx_assign.go` | `syntax/word_operators_test.php` |
| named arguments (user functions, constructors, variadics, a table of built-ins) | `parser/phpx_named.go`, `runner/phpx_named.go` | `syntax/named_arguments_test.php` |
| `[...$a, ...$b]`, a string-keyed spread binding by name | `runner/expr/phpx_spread_array.go` | `syntax/named_arguments_test.php` |
| `print` | `parser/phpx_logical.go` | `syntax/print_test.php` |
| `match`, `throw` as an expression | `parser/phpx_match.go` | `syntax/match_throw_test.php` |
| `clone`, `__clone`, object `==` and `<` | `parser/phpx_match.go`, `runner/phpx_spl.go`, `internal/phpval/compare.go` | `classes/clone_test.php`, `classes/object_compare_test.php` |
| `enum`, backed and pure, `cases`, `from`, `tryFrom`, methods, constants, interfaces | `parser/phpx_enum.go`, `stdlib/phpx/enums.go` | `classes/enum_test.php` |
| `use A\{B, C as D}` | `parser/phpx_usegroup.go` | `syntax/group_use_test.php` |
| bare `{ ... }` block | `parser/parser.go` | `syntax/bare_block_test.php` |
| `$cls::NAME`, `$cls::$prop`, `$cls::method()`, `$cls::$method()` | `parser/phpx_classof.go` | `classes/static_on_value_test.php` |
| `$s[0] = 'x'` | `runner/phpx_stroffset.go` | `syntax/string_offset_write.phpt` |
| `instanceof self`, `static`, `parent` | `runner/expr/direct_operators.go` | `classes/instanceof_self_test.php` |

Runtime and library:

| Feature | Seam | Fixture |
|---------|------|---------|
| `Reflection{Class,Object,Method,Function,Property,Parameter,NamedType,Attribute}`; `getAttributes()`, `getDocComment()` read the declaration's source | `runner/phpx_prelude/reflection.php`, `runner/phpx_reflect.go`, `runner/phpx_attributes.go` | `classes/reflection_test.php`, `classes/attributes_test.php` |
| `SplObjectStorage`, `SplStack`, `SplQueue`, `SplFixedArray`, `ArrayObject`, `ArrayIterator`, `SplObserver`, `SplSubject`; `FilterIterator`, `CallbackFilterIterator`, `RecursiveFilterIterator`, `LimitIterator`, `IteratorIterator`, `RecursiveIteratorIterator`, `RecursiveArrayIterator`, `EmptyIterator` | `runner/phpx_prelude/spl.php` | `classes/spl_structures_test.php`, `classes/spl_iterators_test.php` |
| `Iterator`, `IteratorAggregate`, `Countable`, `ArrayAccess` on script classes; `iterator_to_array`, `iterator_count`, `is_countable`, `instanceof Traversable` | `runner/phpx_spl.go`, `stdlib/phpx/spl.go` | `classes/spl_interfaces_test.php`, `classes/countable_traversable_test.php` |
| arrays are values: an assignment from storage, a by-value parameter the body writes, a return of storage copy | `runner/phpx_value.go` | `classes/array_value_semantics_test.php` |
| `new Exception($m, $c, $previous)`, `getPrevious`, `getTraceAsString`, `$e instanceof Throwable` | `stdlib/phpx/exceptions.go`, `runner/phpx_instanceof.go` | `classes/exception_previous_test.php` |
| `catch (Name)` inside a namespace matches by last segment; `parent::` through a chain | `runner/phpx_inheritance.go` | `classes/catch_namespaced_test.php`, `classes/parent_chain_test.php` |
| `json_encode` of `JsonSerializable` and of public properties; `<`, `>`, `&` unescaped | `stdlib/phpx/json.go` | `classes/json_serializable_test.php` |
| `php://memory`, `php://temp`, `php://stdout`, `php://stderr`; `stream_get_contents` with length and offset | `stdlib/files/phpx_wrappers.go` | `functions/php_memory_stream_test.php`, `classes/json_serializable_test.php` |
| `\R`, `\h` and the POSIX classes in a pattern | `stdlib/compat/phpx_escapes.go` | `functions/preg_escapes_test.php`, `functions/preg_posix_test.php` |
| `class_implements`, `class_alias`, `constant('A::B')`, `is_numeric` of a numeric string, sort flags, callable strings with a leading backslash | `runner/phpx_classinfo.go`, `stdlib/phpx/*.go`, `runner/callable.go` | `functions/class_info_test.php`, `functions/sort_flags_test.php`, `functions/callable_strings_test.php` |
| `current`, `key`, `next`, `prev`, `reset`, `end` | `stdlib/phpx/pointer.go` | `functions/array_pointer_test.php` |
| `E_*`, `JSON_*`, `SORT_*`, `LC_*`, `PHP_URL_*` | `stdlib/phpx/constants.go` | `functions/json_constants_test.php` |
| `assert`, `printf`, `gc_*`, `setlocale`, `base_convert`, `array_diff_assoc`, three `libxml_*` no-ops ... | `stdlib/phpx/runtime_funcs.go` | exercised by the phpunit run |

### Coverage

Go statement coverage of the fork-owned files (76 files, 3502 statements), from two sources:
`go test -tags phpx -coverpkg=./... ./...`, and a binary built with `go build -cover` that ran every fixture under
`phpx/testdata` and `tests/fixtures` and the five phpunit test files (`GOCOVERDIR`, then `go tool covdata textfmt`). The
last column is the union of the two. The fork-owned files are 86% covered together (52% by `go test`, 81%
by the fixtures); the whole module is 26465 statements.

| File | Statements | go test | fixtures | together |
|------|-----------:|--------:|---------:|---------:|
| `formatter/phpx_traits.go` | 64 | 98% | 0% | 98% |
| `internal/phpval/phpx_numeric.go` | 2 | 100% | 100% | 100% |
| `lint/phpx_interfaces.go` | 10 | 90% | 0% | 90% |
| `model/phpx_interfaces.go` | 73 | 73% | 26% | 73% |
| `model/phpx_traitrules.go` | 1 | 100% | 100% | 100% |
| `parser/phpx_arrow.go` | 33 | 88% | 82% | 88% |
| `parser/phpx_classof.go` | 24 | 21% | 92% | 92% |
| `parser/phpx_coalesce.go` | 21 | 90% | 86% | 90% |
| `parser/phpx_destructure.go` | 68 | 84% | 84% | 84% |
| `parser/phpx_enum.go` | 94 | 99% | 99% | 99% |
| `parser/phpx_heredoc.go` | 74 | 82% | 89% | 91% |
| `parser/phpx_logical.go` | 47 | 83% | 87% | 91% |
| `parser/phpx_match.go` | 70 | 76% | 83% | 83% |
| `parser/phpx_named.go` | 11 | 91% | 91% | 91% |
| `parser/phpx_nullsafe.go` | 8 | 100% | 100% | 100% |
| `parser/phpx_spaceship.go` | 3 | 100% | 100% | 100% |
| `parser/phpx_traits.go` | 70 | 86% | 84% | 87% |
| `parser/phpx_usegroup.go` | 44 | 86% | 84% | 86% |
| `runner/expr/phpx_assign.go` | 12 | 58% | 92% | 92% |
| `runner/expr/phpx_builtin_params.go` | 29 | 17% | 97% | 97% |
| `runner/expr/phpx_offset.go` | 33 | 48% | 76% | 76% |
| `runner/expr/phpx_quiet.go` | 7 | 100% | 100% | 100% |
| `runner/expr/phpx_spread.go` | 118 | 25% | 92% | 92% |
| `runner/expr/phpx_spread_array.go` | 46 | 11% | 78% | 78% |
| `runner/phpx.go` | 10 | 100% | 80% | 100% |
| `runner/phpx_abstract.go` | 47 | 81% | 89% | 91% |
| `runner/phpx_assign.go` | 13 | 23% | 54% | 54% |
| `runner/phpx_attributes.go` | 208 | 0% | 89% | 89% |
| `runner/phpx_called.go` | 28 | 57% | 82% | 82% |
| `runner/phpx_classinfo.go` | 92 | 2% | 88% | 88% |
| `runner/phpx_closure.go` | 36 | 28% | 86% | 86% |
| `runner/phpx_funcdecl.go` | 9 | 89% | 78% | 89% |
| `runner/phpx_inheritance.go` | 152 | 84% | 87% | 91% |
| `runner/phpx_instanceof.go` | 23 | 17% | 96% | 96% |
| `runner/phpx_interfaces.go` | 58 | 90% | 95% | 95% |
| `runner/phpx_isa.go` | 15 | 60% | 93% | 100% |
| `runner/phpx_named.go` | 92 | 25% | 88% | 88% |
| `runner/phpx_prelude.go` | 15 | 33% | 67% | 67% |
| `runner/phpx_promotion.go` | 31 | 100% | 97% | 100% |
| `runner/phpx_properties.go` | 113 | 79% | 94% | 94% |
| `runner/phpx_reflect.go` | 182 | 1% | 80% | 80% |
| `runner/phpx_spl.go` | 116 | 22% | 78% | 78% |
| `runner/phpx_statics.go` | 25 | 88% | 88% | 92% |
| `runner/phpx_stringer.go` | 21 | 48% | 86% | 86% |
| `runner/phpx_stroffset.go` | 34 | 9% | 59% | 59% |
| `runner/phpx_throwable.go` | 37 | 89% | 89% | 89% |
| `runner/phpx_traitrules.go` | 47 | 98% | 89% | 98% |
| `runner/phpx_traits.go` | 181 | 80% | 78% | 82% |
| `runner/phpx_value.go` | 18 | 100% | 100% | 100% |
| `runner/phpx_vendor.go` | 10 | 80% | 0% | 80% |
| `runner/phpx_visibility.go` | 31 | 94% | 97% | 97% |
| `stdlib/compat/phpx_escapes.go` | 27 | 93% | 89% | 96% |
| `stdlib/core/phpx_called.go` | 3 | 67% | 100% | 100% |
| `stdlib/core/phpx_isa.go` | 3 | 67% | 100% | 100% |
| `stdlib/core/phpx_json.go` | 37 | 81% | 81% | 81% |
| `stdlib/core/phpx_version_compare.go` | 69 | 3% | 90% | 90% |
| `stdlib/files/phpx_wrappers.go` | 23 | 65% | 87% | 91% |
| `stdlib/phpx/arrays.go` | 22 | 14% | 100% | 100% |
| `stdlib/phpx/classes.go` | 41 | 76% | 93% | 93% |
| `stdlib/phpx/clock.go` | 26 | 19% | 92% | 92% |
| `stdlib/phpx/constants.go` | 4 | 100% | 100% | 100% |
| `stdlib/phpx/ctype.go` | 17 | 24% | 100% | 100% |
| `stdlib/phpx/enums.go` | 44 | 9% | 84% | 84% |
| `stdlib/phpx/environment.go` | 24 | 46% | 92% | 92% |
| `stdlib/phpx/exceptions.go` | 13 | 62% | 92% | 92% |
| `stdlib/phpx/json.go` | 85 | 60% | 84% | 84% |
| `stdlib/phpx/logical.go` | 22 | 36% | 91% | 91% |
| `stdlib/phpx/pointer.go` | 54 | 63% | 83% | 83% |
| `stdlib/phpx/reflection.go` | 48 | 38% | 79% | 81% |
| `stdlib/phpx/runtime_funcs.go` | 139 | 35% | 44% | 44% |
| `stdlib/phpx/sortflags.go` | 16 | 88% | 88% | 88% |
| `stdlib/phpx/spaceship.go` | 11 | 18% | 100% | 100% |
| `stdlib/phpx/spl.go` | 59 | 34% | 68% | 68% |
| `stdlib/phpx/strings.go` | 51 | 18% | 88% | 88% |
| `stdlib/phpx/types.go` | 49 | 27% | 88% | 88% |
| `stdlib/phpx/version.go` | 9 | 100% | 89% | 100% |
| **fork-owned files** | 3502 | 52% | 81% | 86% |
| whole module | 26465 | 83% | 60% | 88% |

### The nine composer libraries, run

`phpx/testdata/packages/tests/` holds two phpunit files for each library, a first pass over the surface and a `*MoreTest.php`
that reaches the corners (`summary.sh <phpx>` prints the table, `compare.sh` diffs the full output). php 8.4.27 passes all
120 tests; the third column is phpx. The first run of this table was 10 of 45. `nikic/php-parser` is not in the set by choice.

| Library | Tests | php | phpx | What stops the rest |
|---|--:|--:|--:|---|
| `psr/log` | 10 | 10 | **10** | |
| `psr/http-message` | 9 | 9 | **9** | |
| `psr/container` | 8 | 8 | **8** | |
| `guzzlehttp/psr7` | 15 | 15 | **15** | |
| `guzzlehttp/guzzle` | 16 | 16 | **16** | |
| `doctrine/inflector` | 17 | 17 | **17** | |
| `symfony/finder` | 12 | 12 | **12** | |
| `symfony/console` | 17 | 17 | **17** | |
| `monolog/monolog` | 16 | 16 | **16** | |
| **total** | **120** | **120** | **120** | 100% |

The 75 added tests failed 16 times on their first phpx run; the work list was: `/` and `%` by zero, `is_subclass_of` on class
names and interfaces, `interface_exists('Throwable')`, `gettype` of a stream, `defined('Enum::Case')`, `PCRE_VERSION`,
`fileinode`, class names in reflected types through the namespace and imports, the `A` regex modifier at an offset, possessive
quantifiers after a group, `mb_detect_encoding`/`mb_convert_encoding` (they were stubs), a class declared after the `require`
that registers its parent's loader, and `is_iterable` of a `Generator` or `ArrayIterator` (guzzle's `EachPromise` wrapped the
iterator in an array when it said no, and every `Pool` request then fulfilled with the iterator).

What the suites found and what fixed it: generators (`yield` ran on a goroutine; doctrine/inflector's rules and symfony/finder
iterate them), `array_map` keeping keys, `parse_url`, `http_build_query` with flags, the SPL file iterators
(`SplFileInfo`, `DirectoryIterator`, `RecursiveDirectoryIterator`, `WeakMap`), interface constants and parents loaded through the
autoloader, `!$x = f()`, a static closure's `self::`, possessive quantifiers, and enum names in a namespace.

### Library coverage

A library passes when 80% of its statements run under phpx while the suites in `phpx/testdata/packages/tests/` run.
`phpx/testdata/packages/coverage.sh [phpx] [profile]` builds the table; `coverage_launcher.php` runs every suite in one
phpunit process and then includes each library's source files, so a file no test reaches counts in the denominator
(php-code-coverage's whitelist does the same). `phpx/scripts/lib-coverage.py <profile> --files <library>` lists a library's
least covered files. The measure is phpx's own statement profile (`--cover`): php here has no xdebug or pcov, so the
php column of this table does not exist; php runs the same tests and passes all of them.

`*MoreTest.php` is the second test file of a library, written after the first pass over its surface to reach the corners
(`Psr7Test` and `Psr7MoreTest`). The `*CovNTest.php` files are the third round, written against the coverage list.

| Library | Statements | Covered | Coverage | Gate (80%) | Blocker |
|---|--:|--:|--:|:-:|---|
| `psr/log` | 11 | 11 | 100.0% | pass | |
| `psr/http-message` | 0 | 0 | interfaces only | pass | |
| `psr/container` | 1 | 1 | 100.0% | pass | |
| `guzzlehttp/psr7` | 1903 | 1598 | 84.0% | pass | |
| `guzzlehttp/guzzle` | 3722 | 3260 | 87.6% | pass | |
| `doctrine/inflector` | 740 | 736 | 99.5% | pass | |
| `symfony/finder` | 606 | 566 | 93.4% | pass | |
| `symfony/console` | 4925 | 3168 | 64.3% | **FAIL** | needs 770 more; ceiling 88% (about 600 statements need process, stopwatch, event-dispatcher, lock, pcntl or a tty); failing tests: same-named private properties in parent and child, unenforced scalar types, DOMDocument, RecursiveTreeIterator, attributes through reflection |
| `monolog/monolog` | 3317 | 2821 | 85.0% | pass | |
| **all nine** | 15225 | 12161 | 79.9% | 8 of 9 pass | |

The blocker text lives in `phpx/testdata/packages/blockers.json`; `lib-coverage.py` prints it, and it is rewritten with each
iteration. 1111 tests run, php passes all of them; phpx fails about 180 of them (errors, failures, incomplete, risky). The
failures are kept in the files as regressions, and what they have in common is on the "blockers by design" list:
`__destruct`, `__serialize`, `__sleep`, `__debugInfo`, `&` references outside `foreach` (so `invokeArgs` with a reference and
a promoted `int &$x`), and scalar parameter types, none of which phpx builds ([design](../docs/design.md)). Everything else
the agents found was fixed (see CHANGELOG).

### PHPUnit

`phpx/testdata/phpunit/` holds a composer tree of `phpunit/phpunit ^11.5` (git ignores `vendor/`), a launcher, five test
files and `compare.sh`:

```sh
cd phpx/testdata/phpunit && composer install --ignore-platform-reqs
phpx run launcher.php --testdox tests/ProviderTest.php
./compare.sh phpx            # each tests/*Test.php under php and phpx, diffed
```

The launcher skips phpunit's entry script, which refuses to start without dom, libxml, tokenizer and xmlwriter; those
extensions stay out of scope. `php launcher.php tests/X.php` is the reference.

| Test file | Covers | php | phpx |
|-----------|--------|-----|------|
| `ATest` | one assertion | `OK (1 test, 1 assertion)` | same |
| `CalcTest` | `setUp`, `expectException`, `assertCount`, a deliberate failure, `markTestSkipped` | `Tests: 7, Assertions: 21, Failures: 1, Skipped: 1.` | same |
| `ProviderTest` | `#[DataProvider]`, `#[TestWith]`, `#[Test]`, `#[CoversNothing]`, named and numbered data sets | `OK (7 tests, 7 assertions)` | same |
| `SkipOnlyTest` | `markTestSkipped`, `markTestIncomplete` | `Tests: 2, Assertions: 0, Skipped: 1, Incomplete: 1.` | same |
| `AssertionsTest` | objects, enums, traits, interfaces, strings, arrays, numbers, `#[Depends]`, output expectations, a diff on arrays, an uncaught `LogicException` | `Tests: 14, Assertions: 58, Errors: 1, Failures: 1.` | same |

### Open

- A throwable's file and line are not recorded, so a failure has no `file:line` under it and `getTrace()` is empty;
  `compare.sh` drops that line from both sides.
- `createMock()` and `createStub()` need `eval` of generated class source; `eval` is in the "Won't build" table.
- `serialize`, `unserialize`: phpunit uses them for process isolation only.
- the bytecode engine keeps upstream's answers where a fixture opts it out: a string offset write, a Go slice that is
  not a script array, and the array copies of `runner/phpx_value.go` (`flatstack: false` in `string_offset_write.phpt` and
  `explode_array.phpt`).
- `class_alias` does not feed `instanceof`; braced `namespace X { }` blocks do not parse; `feof()` on `php://memory`
  answers false until a read hits the end; `json_encode` writes `/` unescaped; `print_r` and `var_dump` do not print an
  enum as php does.
- an array copy is counted by the memory walk where php shares the storage until the first write.
- extension walls (dom, libxml, xmlwriter, tokenizer) are out of scope and are not listed as blockers. The three
  `libxml_*` functions phpunit calls without a guard answer as an empty error buffer; `extension_loaded('libxml')` stays
  false.

### Lint

`phpx/scripts/lint-delta.sh` still reports growth against upstream, and it was not brought to zero in this run:
`grouping/filename` (the exported `Phpx*` methods sit in files named for the area, not the symbol: 62 before this run, 63 now),
`func-args/order` (4, three from earlier fork files), `coverage/uncovered` (5), `godoc/outweighs` (+1) and one
`filecheck/long`. `imports/format` and `pairing/unpaired` were fixed at the end (`splint-fix.sh`, and a test file beside
each new fork file). The gate is not part of any test above.

### Upstream suites

A library's own phpunit suite is the widest check of PHP's behaviour this
repository can run. [testdata/upstream](testdata/upstream/README.md) clones a
library, installs it with phpunit and runs its tests under php and under phpx,
so a difference in the totals is a difference in the language. The fixtures that
come out of a failing test live in the area they belong to (`functions`,
`classes`, `syntax`), settled against php first.

| Library | php | phpx |
|---------|-----|------|
| pimple | 68 tests, 135 assertions | the same |
| doctrine/inflector | 1215 tests, 1238 assertions | the same |
| egulias/email-validator | 358 tests, 477 assertions | 472 assertions, five errors from the intl `Spoofchecker` |

### Upstream

Two pull requests against phpscript turn overlays into registrations; see [testdata/UPSTREAM.md](testdata/UPSTREAM.md).
