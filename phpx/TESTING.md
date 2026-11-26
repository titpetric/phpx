# phpx testing report

What has been measured against php, what a green row means, and what the numbers say to a team that wants to run its
own code on phpx. The app and framework tables are the 2026-11-26 run against `php 8.4.27` as the reference, on this
checkout with `-tags phpx`. Every table regenerates from the command above it.

## Verdict

phpx is **not yet a drop-in runtime for a Laravel or Filament application**, but the five consumer apps now answer
php's request. They load every class php 8.4 loads, 2,356 of 2,356, plus 21 that php 8.4 itself refuses. They parse
every file of their own code and every file of their framework and vendor trees: 35,669 of 35,669 in all. Every one
reaches stage 9, and four answer the status php answers: erpsaas, project-management and akaunting answer the 302 of
a redirect, and urlhub answers php's 500 because the production install omits a dev-only debugbar class. The fifth,
relaticle, answers a 301 where php 8.4 cannot boot the app at all.

**2026-11-26 closed the last walls between the boot and the response.** `openssl` and the DOM subset are bound in Go
(`stdlib/px/phpx/openssl.go`, `stdlib/px/phpx/dom.go`). Then the request path itself exposed four semantic gaps, each
now settled by a php-first fixture: `count($array, COUNT_RECURSIVE)` counted one level, so laravel's `MessageBag`
counted 0 and every validator passed; a fresh `RecursiveIteratorIterator` answered `valid()` false, so urlhub's
helper provider never required `settings()`; `include` of a missing file threw where php warns and answers false, so
carbon's locale loader ended akaunting's request; and `ftruncate()` was unbound, so phpunit's run history stopped a
suite after its first test. The same run made a caught Go error keep its message, file and line through the cause it
unwraps to, and made `instanceof Throwable` agree with `catch (Throwable)`.

What is still missing is the suite-facing surface, not the request: phpunit's XML configuration loader needs
`DOMDocument::loadXML` and `DOMXPath`, so every suite here runs with `--no-configuration`; `createMock()` and
`partialMock()` need `eval`, which is out of scope; and `sqliteCreateFunction()` registered after a connection
opened does not reach an in-memory sqlite database, which stops nine of urlhub's unit tests.

The suite itself runs now. urlhub's whole `tests/Feature` tree, 191 tests and 500 assertions on the dev install with
`--no-configuration`, answers 162 passing, 22 failures and 7 errors; the first run of the tree passed 143. Six bugs
stood behind one another, each now fixed with a php-first fixture. `debug_backtrace()` named no function, and laravel
guesses a relation's foreign key from the caller's name, so every factory-created visit inserted an `_id` column;
once a frame carried its name, `AboutPageTest` passed 16 of 16. The `DOMDocument` constructor refused php's two
arguments, which ended the password-reset request in the mail renderer. Livewire's `#[Computed]` discovery needed
`is_a()` to autoload a class-name string (8 failures), and the wall behind it was
`call_user_func([self::class, 'filterEndDate'])`, which built a fresh empty instance where php binds the caller's.
Livewire's `Synthesizer::getKey()` then needed `property_exists()` to read a class name, and its `IsLazy` needed
`ReflectionClass::isUninitializedLazyObject()`. The in-memory sqlite state the morning's run reported
(`cannot VACUUM from within a transaction`, `table "migrations" already exists`, 14 errors) does not reproduce
after these fixes.

What stops the rest, in the order it is hit: `$errors` nil in a Livewire blade render (5 errors,
`call getBag on nil`); `route('link_detail', ...)` with its required parameter missing (2 errors); `DOMXPath` in the
mail renderer (1); the late `sqliteCreateFunction` (10 `REGEXP` queries); `call getQueryString on nil` in Livewire's
table renderer (7); php's parsing of the dashboard's date string (2); the intl formatter (1); and the unbound GD
extension (1).

| Question | Answer |
|----------|--------|
| Can I run a composer library on phpx? | Yes for the nine measured: 120 of 120 tests, 80% statement coverage on 8 of 9 (`symfony/console` is 64%). |
| Can I run a framework's own code? | Yes for 20 of the 21 shims measured; Twig is blocked by one array write. |
| Can I load a Laravel/Filament app's classes? | Yes for all five: 2,356 of the 2,356 classes php loads (100%), plus 21 that php 8.4 itself cannot load (12 in relaticle, 9 in akaunting). |
| Can I boot one and answer php's request? | Yes for four of five: erpsaas, project-management and akaunting answer php's 302, and urlhub answers php's 500 for the same missing class. relaticle answers a 301 where php 8.4 cannot boot it. The body and the exception are not compared. |
| Can I run an app's own test suite? | Partly. urlhub's whole `tests/Feature` tree runs: 191 tests, 500 assertions, 162 passing, 22 failures and 7 errors on the dev install with `--no-configuration`; the first run of the tree passed 143. Its `RegisterTest` passes 13 of 13 under phpx, where it failed 10 on the 2026-11-26 morning, and a `--cover` pass over it records 43 of 94 `app/` statements in the files it loaded. Its unit service suite passes 5 of 19: nine errors are the late `sqliteCreateFunction`, five failures are `partialMock` and `eval`. Suites run with `--no-configuration` until `loadXML` lands. |
| What is the next blocker for the most apps? | Not the request any more. The suite-facing surface: phpunit's XML configuration loader (`loadXML`, `DOMXPath`), late `sqliteCreateFunction`, and `eval` for mocks. The first two are in the fork's scope. |
| What does a green row mean? | That the measured behaviour matches php's on the code that ran; it does not mean the app is production ready, and it does not mean the suite covers the app. The tables below say which is which. |

## The five apps and the five tests

Five open-source apps, cloned at the commit shown, installed with composer. The scripts are in
[testdata/apps/](testdata/apps/README.md); `testdata/apps/run.sh <phpx> <apps> <apps-prod>` prints every app table
below. The list is capped at five by the workflow that owns it; adding a sixth means dropping one.

| App | Stack | Commit | Framework | Why it is here |
|-----|-------|--------|-----------|----------------|
| [`relaticle/relaticle`](https://github.com/relaticle/relaticle) | Laravel 13, Filament 5, Livewire | `da2f5fd` | `laravel/framework` v13.40 | the most current Filament app: an open CRM, 691 classes |
| [`andrewdwallo/erpsaas`](https://github.com/andrewdwallo/erpsaas) | Laravel 12, Filament 3 | `565501e` | v12.25 | accounting platform on Filament panels, 399 classes |
| [`devaslanphp/project-management`](https://github.com/devaslanphp/project-management) | Laravel 9, Filament 2 | `5498216` | v9.45 | a Filament 2 admin, an older and larger-installed shape |
| [`realodix/urlhub`](https://github.com/realodix/urlhub) | Laravel 13, Fortify | `044bc14` | v13.40 | a small plain Laravel app, no Filament |
| [`akaunting/akaunting`](https://github.com/akaunting/akaunting) | Laravel 10, modules | `8c4a5e5` | v10.50 | a large plain Laravel app: 1,085 classes |

Five tests, the same for each app, in order of what they cost to pass. A test is a gate on the one before it, so a
higher number is a floor and not a grade.

| # | Consumer test | Command | Pass means | It does not mean |
|--:|---------------|---------|-----------|------------------|
| 1 | **Parse**: lint the app and its framework trees | `testdata/apps/survey.sh` | no file fails to parse in any tree | that a file runs; parsing is not execution |
| 2 | **Class load**: `class_exists` every class of `app/`, one process each | `testdata/apps/load.sh` | every class php loads, phpx loads | that a method body works |
| 3 | **Boot**: `bootstrap/app.php`, the console kernel, providers booted | `testdata/apps/boot.sh` | stage 9, the response, is reached | that the response is correct |
| 4 | **First request**: `GET /` through the HTTP kernel | stage 9 of the same script | phpx answers the status php answers | that it is the same body or the same exception |
| 5 | **Measure**: `--cover` over the boot and the request | `boot.sh` | reported, not gated | a coverage figure for the app |

Tests 2 to 5 run on `composer install --no-dev`, which is what ships; test 1 runs on the dev install so `tests/` is
present. Test 5 reports the statements that ran in three trees — the app's own `app/`, `vendor/laravel/framework/src`,
and all of `vendor/` — and the classes each runtime declared. The denominator of each is the statements in files that
were loaded, so a wall makes the percentage look better, not worse; it is a statement about how far the request got.

## Per-project verdicts

One table per app: the five tests, what each returned, the first wall, and whether the fork can move it. The
tables above compare the apps with each other; these say what stands between each one and php's request.

### `relaticle` — Laravel 13, Filament 5

| Test | Result | Blocker | In the fork's scope? |
|------|--------|---------|----------------------|
| 1 Parse | `app/` 691 of 691, `tests/` 655 of 655; all 1,615 vendor files | none | |
| 2 Class load | 691 of 691, 12 beyond the 679 php 8.4 loads | none | |
| 3 Boot | stage 9: 267 routes, the console kernel, providers booted | none | |
| 4 First request | HTTP 301 | php 8.4.27 cannot boot the app: a trait property needs 8.5 | no: php's version gate, not phpx |
| 5 Measure | 519 of 3,431 `app/` statements (15.1%); 3,731 of 17,716 framework (21.1%); 10,186 of 61,794 vendor (16.5%); 1,192 classes declared | the redirect is answered before more `app/` code runs | reported, not gated |

**Verdict: the only app phpx boots past php.** Nothing in the fork stands between it and a response, and every
file parses. The reference request has to come from php 8.5, and a redirect leaves most of its `app/` code
unmeasured.

### `erpsaas` — Laravel 12, Filament 3

| Test | Result | Blocker | In the fork's scope? |
|------|--------|---------|----------------------|
| 1 Parse | `app/` 400 of 400, `tests/` 12 of 12; all 2,671 vendor files | none | |
| 2 Class load | 399 of 399 | none | |
| 3 Boot | stage 9: 102 routes, a response built | none | |
| 4 First request | HTTP 302, php's status | none | |
| 5 Measure | 145 of 2,949 `app/` statements (4.9%); 2,836 of 14,128 framework (20.1%); 6,856 of 41,396 vendor (16.6%); 781 classes declared, php 950 | the redirect is answered before more `app/` code runs | reported, not gated |

**Verdict: answers php's redirect.** The openssl binding closed the cookie-encryption wall, and the request now
returns the 302 php returns. The class count is still below php's because a redirect runs less of the framework;
the earlier gap was the code behind encryption. The next measurement is the app's own test suite, not the request.

### `project-management` — Laravel 9, Filament 2

| Test | Result | Blocker | In the fork's scope? |
|------|--------|---------|----------------------|
| 1 Parse | `app/` 151 of 151, `tests/` 4 of 4; all 954 vendor files | none | |
| 2 Class load | 151 of 151 | none | |
| 3 Boot | stage 9: 77 routes, a response built | none | |
| 4 First request | HTTP 302, php's status | none | |
| 5 Measure | 53 of 1,087 `app/` statements (4.9%); 2,813 of 12,465 framework (22.6%); 6,259 of 30,887 vendor (20.3%); 716 classes declared, php 886 | the redirect is answered before more `app/` code runs | reported, not gated |

**Verdict: PDO and openssl move it from the bootstrap to php's redirect.** The Eloquent connection, its fetch
modes and its `PRAGMA` calls all answer, the routes register (77 of php's 77), and the request returns php's 302.
The next measurement is the app's own test suite.

### `urlhub` — Laravel 13, Fortify

| Test | Result | Blocker | In the fork's scope? |
|------|--------|---------|----------------------|
| 1 Parse | `app/` 52 of 52, `tests/` 53 of 53; all 1,412 vendor files | none | |
| 2 Class load | 51 of 51 | none | |
| 3 Boot | stage 9: 58 routes, a response built | none | |
| 4 First request | HTTP 500, php's status, on the same missing class | the production install omits the dev-only `Fruitcake\LaravelDebugbar` facade, and both runtimes stop there | no: the production install's own dev-dependency gap |
| 5 Measure | 32 of 81 `app/` statements (39.5%); 3,499 of 17,439 framework (20.1%); 6,864 of 37,050 vendor (18.5%); 608 classes declared, php 776 | the exception is rendered from a dev-only class, so both sides answer the same status from the same wall | reported, not gated |

**Verdict: the request matches php, and the app's own suite runs.** Both answer 500 on the facade the production
install does not ship, which is where php stops too. On the dev install, the whole `tests/Feature` tree runs under
phpx with `--no-configuration`: 191 tests, 500 assertions, 162 passing, 22 failures and 7 errors, where the morning's
first run passed 143, and no test in it stops on the request. `RegisterTest` passes 13 of 13 (75 assertions) where
the morning's run failed 10: `settings()` needed the `RecursiveIteratorIterator` start, the unique and required
rules needed `count(..., COUNT_RECURSIVE)`, and phpunit's run history needed `ftruncate()`. A coverage pass over the
same file records 43 of 94 `app/` statements in the files it loaded (45.7%), 6,765 of 25,432 framework (26.6%) and
18,485 of 85,938 vendor (21.5%). `--cover` over `tests/Feature/Config/AboutPageTest` records 73 of 236 `app/`
statements (30.9%): the same measure, on a suite file that passes. The unit suite is the next measurement: 9 of its
19 failures are the late `sqliteCreateFunction`, 5 are `partialMock` and `eval`.

### `akaunting` — Laravel 10, modules

| Test | Result | Blocker | In the fork's scope? |
|------|--------|---------|----------------------|
| 1 Parse | `app/` 1,086 of 1,086, `tests/` 52 of 52; all 1,168 vendor files | none | |
| 2 Class load | 1,085 of 1,085 (the 1,076 php loads, and 9 php 8.4 itself refuses) | none | |
| 3 Boot | stage 9: 613 routes, a response built | none | |
| 4 First request | HTTP 302, php's status | none | |
| 5 Measure | 235 of 3,165 `app/` statements (7.4%); 2,955 of 13,318 framework (22.2%); 6,778 of 35,636 vendor (19.0%); 764 classes declared, php 931 | the redirect is answered before more `app/` code runs | reported, not gated |

**Verdict: the destructor closed the route gap, the highlighter the renderer, and `include` the request.**
Laravel's `Route::resource()` registers its routes in a `PendingResourceRegistration` destructor, so phpx registers
all 613 routes where php registers 613. The request then stopped in carbon's `AbstractTranslator`, which probes
locale files with `@include`; php warns and returns false, phpx threw `Errno: no such file or directory`, and the
binding now answers as php does. The app returns php's 302. The classes php declared behind the response are still
more than phpx's, because a redirect runs less of the framework.

## Across the apps

### Test 1: parse

Files that parse. A file parses when the linter reports `PASS` or `WARN` on it; a `WARN` is a lint finding (an
unknown class, an unused variable), not a parse failure. `app/` is the app's own code; the vendor rows show the
walls it inherits.

| App | `app/` | `tests/` | `laravel/framework` | `filament` | `livewire` | `symfony` | All trees |
|-----|-------:|---------:|--------------------:|-----------:|-----------:|----------:|----------:|
| `relaticle` | 691 of 691 | 655 of 655 | 1,746 of 1,746 | 4,887 of 4,887 | 306 of 306 | 1,615 of 1,615 | 10,355 of 10,355 |
| `erpsaas` | 400 of 400 | 12 of 12 | 1,454 of 1,454 | 3,274 of 3,274 | 240 of 240 | 2,671 of 2,671 | 8,246 of 8,246 |
| `project-management` | 151 of 151 | 4 of 4 | 1,186 of 1,186 | 1,760 of 1,760 | 124 of 124 | 954 of 954 | 4,298 of 4,298 |
| `urlhub` | 52 of 52 | 53 of 53 | 1,746 of 1,746 | 2,530 of 2,530 | 306 of 306 | 1,412 of 1,412 | 6,182 of 6,182 |
| `akaunting` | 1,086 of 1,086 | 52 of 52 | 1,287 of 1,287 | n/a | 244 of 244 | 1,168 of 1,168 | 6,588 of 6,588 |
| **all five** | **2,380 of 2,380** | **776 of 776** | **7,419 of 7,419** | **12,451 of 12,451** | **1,220 of 1,220** | **7,820 of 7,820** | **35,669 of 35,669** |

Verdict: **parse is not the blocker.** Every file of every app and its framework and vendor trees parses. The ten
files that failed on 2026-11-23 were three constructs, and all three are now implemented:

| Parse failure | Files | Where | Language feature | Status |
|---------------|------:|-------|------------------|--------|
| `isset(self::${$property}[$use])`, `self::${$property}[$class] = ...` | 5 | `symfony/error-handler` `DebugClassLoader` (every app) | a static property named by an expression, with an array index | fixed 2026-11-23 |
| `yield $r => $chunk = new X`, `(int) $ch = $info['h']` | 4 | `symfony/http-client` `AsyncResponse`, `CurlResponse` (akaunting, relaticle) | assignment as an operand after `=>` and after a cast | fixed 2026-11-24 |
| `$this->t ??= clone $t ??= ...` | 1 | `symfony/intl` `EmojiTransliterator` (erpsaas) | `??=` as an operand of `clone` | fixed 2026-11-24 |

The five `symfony/routing` files that failed on `{$this->indent(...)}` in a heredoc now parse: the heredoc body is
scanned as a body and not as a quoted literal, and `scanInterpBody` keeps the expression's own source
(`phpx/testdata/syntax/heredoc_test.php`, `parser/interp_test.go`). The four akaunting `Use Documents;` files parse
since 2026-11-21, because php reads its keywords without regard to case. The static property expression landed
2026-11-23, and the assignment as an operand and `??=` under `clone` landed 2026-11-24
(`phpx/testdata/syntax/static_property_name.phpt`, `assign_operand_test.php`,
`assign_operand_generator_test.php`).

### Test 2: class load

| App | Classes in `app/` | php loads | phpx loads | Both load | phpx / php | What stops the rest |
|-----|-----------------:|----------:|-----------:|----------:|-----------:|---------------------|
| `erpsaas` | 399 | 399 | 399 | 399 | **100%** | |
| `project-management` | 151 | 151 | 151 | 151 | **100%** | |
| `relaticle` | 691 | 679 | 691 | 679 | **100%** | php loads 679 of 691: 12 fatal on php 8.4 itself, because the app requires 8.5. All 691 load under phpx |
| `urlhub` | 51 | 51 | 51 | 51 | **100%** | |
| `akaunting` | 1,085 | 1,076 | 1,085 | 1,076 | **100%** | |

Verdict: class loading **passes on all five apps**: 2,356 of the 2,356 classes php loads. The class model
(inheritance, traits, interfaces, enums, attributes, promoted constructors, property hooks, `readonly`) composes at
app scale across 2,377 classes and the Filament, Livewire and Symfony trees they extend. phpx loads 21 more than
php 8.4 does, which php refuses at class composition time.

### Tests 3 to 5: boot, first request, measure

Stages reached (1 autoload, 2 application, 3 kernel, 4 bootstrap, 5 container, 6 version, 7 routes, 8 request,
9 response). All five reach 9 under phpx. php reaches 9 for four; relaticle stops at 4 under php 8.4 because it
needs 8.5.

| App | php | phpx | phpx stopped at | `app/` statements | framework statements | vendor statements | classes php / phpx |
|-----|-----|------|-----------------|------------------:|---------------------:|-----------------:|-------------------:|
| `akaunting` | 302 | 302 | nothing: php's redirect | 235 of 3,165 (7.4%) | 2,955 of 13,318 (22.2%) | 6,778 of 35,636 (19.0%) | 931 / 764 |
| `erpsaas` | 302 | 302 | nothing: php's redirect | 145 of 2,949 (4.9%) | 2,836 of 14,128 (20.1%) | 6,856 of 41,396 (16.6%) | 950 / 781 |
| `project-management` | 302 | 302 | nothing: php's redirect | 53 of 1,087 (4.9%) | 2,813 of 12,465 (22.6%) | 6,259 of 30,887 (20.3%) | 886 / 716 |
| `relaticle` | 4 (php 8.4 fatal) | 301 | nothing: a response where php cannot boot | 519 of 3,431 (15.1%) | 3,731 of 17,716 (21.1%) | 10,186 of 61,794 (16.5%) | n/a / 1,192 |
| `urlhub` | 500 | 500 | the dev-only debugbar class, where php stops too | 32 of 81 (39.5%) | 3,499 of 17,439 (20.1%) | 6,864 of 37,050 (18.5%) | 776 / 608 |

Verdict: **four of five answer php's status, and the fifth answers a response where php 8.4 cannot boot.** The
routes register on every app: akaunting 613 of php's 613, urlhub 58 of 58, project-management 77 of 77, erpsaas 102
of 102, relaticle 267 where php cannot count. What stands between the measurement and an app's own suite:

| Rank | Blocker | Tests stopped | In the fork's scope? | State |
|-----:|---------|--------------:|----------------------|-------|
| 1 | phpunit's XML configuration loader (`DOMDocument::loadXML`, `DOMXPath`) | every suite, which runs `--no-configuration`; the same `DOMXPath` stops the reset-password mail | yes: the DOM binding | pending |
| 2 | `$errors` and the request nil in Livewire's renderer (`call getBag on nil`, `call getQueryString on nil`), and `route()`'s required parameter | 14 of urlhub's 29 `Feature` failures and errors | not established: not root-caused | pending |
| 3 | `sqliteCreateFunction()` after the connection opened, on `:memory:` | 9 of urlhub's 19 unit tests and 10 `Feature` failures | yes: the PDO binding | pending |
| 4 | `createMock()` and `partialMock()` need `eval` | 5 of urlhub's 19 unit tests | no: `eval` is on the won't-build table | documented |
| 5 | the dev-only debugbar class the production install omits | urlhub's request, on php and phpx alike | no: the app's install | documented |
| 6 | the intl formatter and GD | 2 urlhub `Feature` failures | no: extensions off the table | documented |

The walls this run closed, each with the fixture that settles it: `count(..., COUNT_RECURSIVE)`
(`functions/count_recursive.phpt`), the `RecursiveIteratorIterator` start and `isDot` forwarding
(`classes/recursive_directory_iterator.phpt`), `include` of a missing file answering false
(`syntax/include_missing.phpt`), and `ftruncate()` (`functions/ftruncate.phpt`). The same run made a caught Go error
keep its message and site through the cause it unwraps to, and made `instanceof Throwable` agree with
`catch (Throwable)`.

The ladder is the point: a boot test stops at the first wall, so the counts are a floor. Each fix moves the stage
and uncovers the next wall. The measure column is deliberately modest: it is the part of the app that executed
before the wall, not a coverage figure for the app, and it must not be quoted as one. The class count is the honest
companion: phpx declares fewer classes than php on every app, because a redirect or an error page runs less of the
framework, not because a class failed to load.

## What a pass means

| Row says | It means | It does not mean |
|----------|----------|------------------|
| a `.phpt` fixture passes on `php`, `runtime`, `flatstack` | the three columns print what php 8.4 printed for that program (the expected output is php's, not phpx's) | the feature is complete; a fixture pins the cases someone thought of |
| a library test passes | the library's code path ran to the same assertion result under phpx as under php | every path of the library runs; see its statement coverage |
| a library's coverage is 80%+ | 80% of the library's statements ran while those tests ran, files no test reached counted as uncovered | the tests are good, or the lines are right: a line that ran and returned a wrong value counts as covered |
| a file `parses` (lint PASS or WARN) | phpx's parser built an AST for it | it runs: a file that parses can still call an undefined function or a missing class |
| a class `loads` | autoload found the file, parsed it, and every parent, interface and trait it names resolved with php's composition rules | its methods work; loading executes declarations and constant expressions, not method bodies |
| an app `boots` to stage N | the artisan bootstrap got through stage N with php's behaviour at every call so far | stage N+1 works |
| an app answers php's status | the HTTP status is the same | the body, the exception or the code path is the same; urlhub answers php's 500 on the same missing class, and the other four are redirects |
| an app's test file passes under phpx | the test's assertions held on the code path it exercised, with the app's own bootstrap | the whole suite passes; urlhub's `Feature` tree is 162 of 191, its `RegisterTest` 13 of 13 and its unit service suite 5 of 19 |
| app statements run | statements of `app/` files that executed while phpx booted, out of the statements of the files loaded by then | a coverage of the app; the denominator is only the files reached |
| classes declared (php / phpx) | how much of the runtime's code the request reached | that the same classes ran the same way |

## What the measurement means for someone who wants to integrate

| If you are... | The state today | What to do |
|---------------|-----------------|------------|
| shipping a library (PSR, Guzzle, Monolog, Symfony components) | supported: nine libraries run their own tests with php's results | run the library's phpunit suite under phpx; a difference in totals is a bug, file it with the fixture |
| embedding PHP business logic in a Go service (value objects, validation, rules) | supported for plain PHP and the OOP baseline | write the logic against the fixtures' feature list; avoid `eval`, references and scalar type coercion |
| running a Laravel or Filament application | close: every file parses, every class loads, every app boots, four of five answer php's status, and a whole feature directory runs with `--no-configuration`; the failures left are named walls, not the request | run the app's `GET /`, then a feature directory, and `phpx --cover --coverfile=... run vendor/bin/phpunit --no-configuration <file>`; file a fixture for the first wall, not the whole app |
| depending on `__destruct` (Laravel's `Route::resource()`, an observer that flushes on destruction) | supported: the destructor runs when the last reference goes and again for what is alive at the end | nothing; `phpx/testdata/classes/destruct.phpt` names the cases that hold |
| depending on PDO | supported over sqlite, mysql and pgsql, with the limits named below | use it; `phpx/testdata/pdo/` settles the modes the apps call, and register `sqliteCreateFunction` before the first query |
| using Twig | blocked: `$weakMap[$node][] = $value` is a nested append into an `ArrayAccess` object, and phpx refuses the write | wait for the write path, or avoid the pattern |
| depending on openssl, DOM, libxml, mbstring or Carbon's date classes | `openssl` and a DOM subset are bound; libxml, mbstring and the rest are out of scope | expect an issue before a box; `phpx/testdata/report.md` records what each project reached for |
| running phpunit with a `phpunit.xml` | blocked: the configuration loader needs `loadXML` and `DOMXPath` | pass `--no-configuration` and the bootstrap and env the file names, as `testdata/apps` does |
| wanting a percentage to quote | quote the class-load rate and the stage reached, with the app and commit | do not quote app statement coverage: it is not measured until a request completes |

How to qualify an app of your own, in the order that costs least:

1. `phpx lint --include vendor/autoload.php ./app/...` — a parse failure is the cheapest thing to file.
2. `phpx/testdata/apps/load.sh bin/phpx <app>` — one process per class, so a fatal costs one class.
3. `phpx/testdata/apps/boot.sh bin/phpx <app>` — the boot and the request, and the wall's name.

The four measures this report leans on:

| Measure | What it tells you | What it hides |
|---------|-------------------|---------------|
| Parse rate | the language is covered; a failing file names the missing syntax | a file can parse and still not run |
| Class load rate | the object model composes the way php composes it, at 100s of classes | method bodies, calls, extension use |
| Stage reached | the first thing that stops a real program | what is behind the wall; the stage is a floor |
| Statements run | how much code the wall let through, in three trees | whether the code that ran was correct |

The fifth measure, `app/` statement coverage from the app's own suite, is available for a file now: urlhub's
`RegisterTest` under `phpx --cover` records 43 of 94 `app/` statements in the files it loaded. The denominator is
the files the suite reached, so it is a floor in the same way the request measure is, and a suite-wide number
waits on the two binding walls named above.

## Bugs the measurement found and fixed

The first run of these tests stopped on bugs, not walls: four, each fixed with a fixture whose expected output is
php's. The 2026-11-17 run found no new one: the parse column fell from 28 failing files to 19 because
property hooks and `readonly` landed in between, and class load rose from 98.7% to 100% on the same four apps.
The 2026-11-21 run found php's keyword case; the 2026-11-23 run found the heredoc body and the missing highlighter;
the 2026-11-24 run found the assignment as an operand, the last parse failure. The 2026-11-26 run found no parse
failure at all and turned to what the request does after it parses: the recursive count, the iterator start, the
missing include and the unbound `ftruncate()`. The suite run that followed found six more, each behind the last:
a frame with no function name for `debug_backtrace()`, the `DOMDocument` constructor's arity, `is_a()` not
autoloading a class-name string, `call_user_func([self::class, ...])` building a fresh instance, `property_exists()`
answering false for a class name, and an unbound `ReflectionClass::isUninitializedLazyObject()`.

| Bug | Seen in | Fixture |
|-----|---------|---------|
| `self::${$property}[$class]` read a property named by nothing, so the value was null | `symfony/error-handler`'s `DebugClassLoader` (5 files, every app) | `testdata/syntax/static_property_name.phpt` |
| `Use Documents;` read as an expression: php reads keywords without regard to case | akaunting `app/Imports/*/Sheets` (4 files, 1,072 -> 1,085 classes) | `testdata/syntax/keyword_case_test.php` |
| a heredoc body embedding `{$this->indent("y")}` was lost silently, and an unreadable body lost the literal | `symfony/routing` `CompiledUrlMatcherDumper` (5 files, every app) | `testdata/syntax/heredoc_test.php`, `parser/interp_test.go` |
| `highlight_file()` was not bound, so the Symfony error renderer failed on the exception's source | akaunting's request (stage 8 -> 9) | `testdata/functions/highlight_test.php`, `stdlib/ps/core/phpx_highlight_test.go` |
| `WeakMap` named itself, not the key object, in `Object ... not contained in WeakMap` | the Twig shim | no fixture: php's object handle is not reproducible, the class name is |
| an assignment as a prefix operator's operand (`(int) $ch = $h`, `yield $r => $chunk = new X`, `clone $c ??= new X`) parsed as an empty variable name or dropped the write | `symfony/http-client` `AsyncResponse`, `CurlResponse`; `symfony/intl` `EmojiTransliterator`; the five apps' last parse failures | `testdata/syntax/assign_operand_test.php`, `assign_operand_generator_test.php` |
| `\false`, `\true`, `\null` read as an unknown constant | `google/apiclient` aliases (relaticle stopped at stage 1) | `testdata/syntax/qualified_constants.phpt` |
| a trait reached by two paths (class uses `T` and `U`, `U` uses `T`) raised "has not been applied ... collision" | Livewire's `WithFileUploads` and `WithRateLimiting`, 22 classes | `testdata/traits/shared_trait.phpt` |
| `IteratorIterator` did not forward unknown methods to its inner iterator | `symfony/finder`'s `ExcludeDirectoryFilterIterator::getFilename()`, stage 4 of two apps | `testdata/classes/iterator_iterator_forwards.phpt` |
| a protected method called from outside did not go to `__call`/`__callStatic` | Laravel's `Router::prefix()` through the `Route` facade, stage 4 of two apps | `testdata/classes/magic_call_inaccessible.phpt` |
| `count($array, COUNT_RECURSIVE)` counted one level, so laravel's `MessageBag::count()` was 0 and every validator rule passed | urlhub's `RegisterTest` (8 of 13 tests failed) | `testdata/functions/count_recursive.phpt` |
| a fresh `RecursiveIteratorIterator` answered `valid()` false, and `isDot()` was not forwarded, so urlhub's helper provider required no file and `settings()` was undefined | urlhub's request (5 of 13 tests) | `testdata/classes/recursive_directory_iterator.phpt` |
| `include` of a missing file threw `Errno: no such file or directory` where php warns and answers false | carbon's `AbstractTranslator::resetMessages()` (akaunting's request) | `testdata/syntax/include_missing.phpt` |
| `ftruncate()` was not bound, so phpunit's run history stopped a suite after its first test | urlhub's suite | `testdata/functions/ftruncate.phpt` |
| `debug_backtrace()` answered `''` for every frame's `function`, so laravel's `belongsTo()` guessed the foreign key `_id` instead of `url_id` | every factory-created `Visit` (16 of urlhub's 23 Feature errors) | `testdata/functions/debug_backtrace.phpt` |
| `new DOMDocument('1.0', 'UTF-8')` answered an arity error: the bound constructor took no arguments | laravel's mail renderer, and with it the reset-password notification | `testdata/classes/dom_document.phpt` |
| `is_a()` answered false for a class-name string because it looked only in the loaded table; php autoloads the name | livewire's `#[Computed]` discovery, so `$this->visibleColumns` was an unknown property (8 urlhub Feature failures) | `testdata/classes/is_a_autoload_test.php` |
| `call_user_func([self::class, 'method'])` built a fresh empty instance, so a protected method read `__get`'s answer instead of the receiver's property; `is_callable()` answered true for a non-static pair php refuses | CarbonPeriod's filters, `Unknown getter 'options'` in urlhub's charts | `testdata/functions/call_user_func_scope.phpt` |
| `property_exists()` answered false for a class name; php reads the class's declarations, statics included, and autoloads the name | Livewire's `Synthesizer::getKey()`, so every array property on a component answered "You need to define static $key property" | `testdata/functions/property_exists_class.phpt` |
| `ReflectionClass::isUninitializedLazyObject()` was unbound; php 8.4 has it | Livewire's `IsLazy` asks it of every hydrated model | `testdata/classes/reflection_lazy_object.phpt` |
| a caught Go error was unwrapped to its cause, which lost php's message and the file and line, and `instanceof Throwable` disagreed with `catch (Throwable)` | akaunting's exception, `Errno` where php names `Error` | `go test -tags phpx ./runner -run TestPhpxSite` (no fixture: php has no `Errno`) |

## Frameworks and libraries

### Framework shims

Each shim is a small phpunit suite over the framework's public surface, in
[testdata/frameworks/](testdata/frameworks/). `cd phpx/testdata/frameworks/<name> && phpx run launcher.php tests/`;
`testdata/frameworks/summary.sh bin/phpx` prints the table.

| Framework | Version | Tests | php | phpx | Verdict |
|-----------|---------|------:|-----|------|---------|
| Slim | 4.15.3 | 5 | 5 | 5 | pass |
| Latte | 3.1.7 | 5 | 5 | 5 | pass |
| Laravel | 12.71.0 | 4 | 4 | 4 | pass: container, collections, support |
| Filament | 4.16.1 on Laravel | 2 | 2 | 2 | pass: not a panel booted, see the app tests |
| Twig | 3.30.0 | 5 | 5 | 0 | **fail**: `$weakMap[$node][] = $value`, a nested append into an ArrayAccess object |
| CommonMark | 2.10.7 | 4 | 4 | 4 | pass |
| Symfony YAML | 7.4.23 | 5 | 5 | 5 | pass |
| Carbon | 3.15.1 | 4 | 4 | 4 | pass |
| nikic/php-parser | 5.9.0 | 3 | 3 | 3 | pass |
| Mustache | 3.2.0 | 4 | 4 | 4 | pass |
| ramsey/uuid | 4.9.4 | 3 | 3 | 3 | pass |
| league/csv | 9.29.0 | 3 | 3 | 3 | pass: php prints three deprecations, and `summary.sh` counts a `⚠` testdox row as a pass |
| brick/math | 0.12.3 | 3 | 3 | 3 | pass |
| symfony/string | 7.4.22 | 2 | 2 | 2 | pass |
| egulias/email-validator | 4.0.4 | 2 | 2 | 2 | pass |
| vlucas/phpdotenv | 5.8.0 | 2 | 2 | 2 | pass |
| ramsey/collection | 2.1.1 | 4 | 4 | 4 | pass |
| symfony/event-dispatcher | 7.4.17 | 2 | 2 | 2 | pass |
| symfony/routing | 7.4.23 | 3 | 3 | 3 | pass: the compiled matcher reads its heredoc now |
| fakerphp/faker | 1.24.1 | 2 | 2 | 2 | pass |
| pimple/pimple | 3.6.2 | 2 | 2 | 2 | pass |

The shims prove the framework code runs. The app tests above prove the same at application scale: the Laravel 12
shim passes, and every Laravel 9 to 13 app now boots and answers the response its runtime gives.

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
the coverage pass and phpx fails about 180 of them; those are the by-design blockers (`__serialize`, `&` references
outside `foreach`, scalar parameter types) and are kept as regressions. The numbers are the 2026-11-18 run.

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
| `phpx test --matrix ./tests/fixtures/...` (upstream) | 327 | 327 pass on flatstack, runtime and php |
| `phpx test --matrix ./phpx/testdata/...` (the fork's, whole tree) | 229 | 228 pass; the one failure is `patches/tests/fixtures/namespaces/instanceof_namespace.phpt`, whose `instanceofns/lib.php` support file is not in the overlay, and the pipeline's job does not run that tree |
| the fork's fixture job (`test:phpx:fixtures`: every `testdata/<area>/`, patches and frameworks skipped) | 219 | 219 pass |
| `go test ./...`, `go test -tags phpx ./...` | all packages | pass |
| phpunit 11.5 from its source tree, `compare.sh` | 9 files | `same` for each; failures that php reports too are expected |

Open in phpunit: `createMock()` and `partialMock()` need `eval`; a Go error's `getTrace()` is empty, and its file
and line come from the statement that raised it; `serialize` is process isolation only. The `phpx phpunit`
subcommand does not exist; `phpx/scripts/phpunit.sh` stands in for it.

### Go coverage of the fork

86% of the fork-owned statements (3,502 in 76 files, measured 2026-11-04) run under `go test` and the fixtures
together; the whole module is 88%. Per-file numbers are in the appendix and per-unit numbers in
[MEASUREMENTS.md](MEASUREMENTS.md). This is coverage of phpx's Go, which says the fork's code is exercised; it says
nothing about how much PHP a program can run, which is what the tables above measure.

### The PDO binding

`stdlib/px/phpx/pdo.go` (with `pdo_statement.go`, `pdo_consts.go` and `pdo_error.go`) binds `PDO`,
`PDOStatement` and `PDOException` over `database/sql`: sqlite through the private modernc driver, mysql and
pgsql through the keyword DSN of each. The fixtures in [testdata/pdo/](testdata/pdo/) settle the connection, both
bind styles, the fetch modes, transactions and error modes against php; `go test -tags phpx ./stdlib/px/phpx`
holds the DSN and placeholder halves.

`fetchAll()` answers php's whole mode list: `FETCH_GROUP`, `FETCH_UNIQUE`, `FETCH_FUNC`, `FETCH_NAMED` and
`FETCH_KEY_PAIR` beside the flat modes, `FETCH_CLASS` runs the class's constructor after it fills the row's
properties, and `FETCH_INTO` fills the object `setFetchMode` was given, with php's errors for a mode whose
argument is missing. [testdata/pdo/fetchmodes_test.php](testdata/pdo/fetchmodes_test.php) settles each against php.

It is written for the callers the apps make, and its limits name where that stops:

- rows are read at `execute()` and kept on the statement, so `fetch()` walks one slice; a table larger than memory
  is not streamed.
- `sqliteCreateFunction` registers on the connection's private driver. A function added after the connection
  opened reopens a file database; an in-memory database lives in the connection it would close, so it gets no
  function added this way. urlhub's `REGEXP` is registered in a service provider after the connection opened, so
  nine of its unit tests answer `no such function: REGEXP`; registering the function before the first query works.
- `bindParam` keeps the value given at the call, not the variable's later value; `bindColumn` is not implemented,
  and `getColumnMeta` answers name and type alone.
- `PDO::ATTR_SERVER_VERSION` is read on first use, and `ATTR_PERSISTENT` is recorded and not honoured.

## Open

| Item | Why it matters | Where |
|------|----------------|-------|
| phpunit's XML configuration loader (`DOMDocument::loadXML`, `DOMXPath`) | rank 1: every app suite runs `--no-configuration`; the DOM binding answers the console renderer and not the loader. The same `DOMXPath` stops laravel's mail renderer at `CssToInlineStyles`, so the reset-password notification is the wall seen from the other side | `stdlib/px/phpx/dom.go` |
| comparing two `DateTime` subclasses (`$a < $b`, `$a == $b`) | CarbonPeriod's iteration: phpx orders two Carbon objects by their property table, and `constructedObjectId` decides before the date does. urlhub's charts stop there | `internal/phpval/compare.go` |
| `call getQueryString on nil` in urlhub's table renderer | 7 of urlhub's `AuthPage` failures: the renderer reaches a host request value that is nil | not root-caused; the trace names the table component, not the binding |
| `$errors` nil in a Livewire blade render (`call getBag on nil`) | 5 urlhub `Feature` errors: the validation partial reaches a view-error bag that nothing shared | not root-caused; `ShareErrorsFromSession` and the compiled view are both in the path |
| `route('link_detail', ...)` with its required `url` parameter missing | 2 urlhub `Feature` errors: a route whose URI is `+{url}` | not root-caused; `Illuminate\Routing\UrlGenerator` |
| the dashboard's date string (`Failed to parse time string (1 )`) | 2 urlhub `Feature` failures: rendering a visit row reaches `new DateTime('1 ')` | not root-caused; the value comes from the model the component reads |
| `sqliteCreateFunction()` after the connection opened, on `:memory:` | 9 of urlhub's 19 unit tests and 10 `Feature` failures answer `no such function: REGEXP` | `stdlib/px/phpx/pdo.go` |
| `createMock()` and `partialMock()` | 5 of urlhub's 19 unit tests; `eval` is on the won't-build table | no fork work |
| nested append into an `ArrayAccess` object (`$weakMap[$node][] = $value`) | the Twig shim; `runner/assign_container.go` resolves an array's container and refuses an object's | `runner/assign_container.go`, `runner/phpx_spl.go` |
| `bindColumn` and the stream filter append family | the registries are bound; reading through a registered filter or binding a column needs the reference machinery at the call and is not wired | `stdlib/px/phpx` |
| the apps' own test suites under phpx | the real coverage measure: urlhub's whole `Feature` tree runs, 162 of 191, and the rest waits on the walls above | `testdata/apps` |
| the OOP benchmark times `oop.php` on php and phpx alone | phpscript refuses the program at its first interface, so there is no upstream run to compare; the flat program is 47% faster on phpscript than on phpx (2026-11-23) | `phpx/testdata/bench/upstream.sh` |

## Appendix: the 2026-11-04 run

The run behind this report took the fork from "phpunit stops on a parse error at `Application.php:835`" to
"phpunit 11.5 runs from its source tree: test discovery, `setUp`, assertions, `#[DataProvider]`, `#[TestWith]`,
`#[Depends]`, exceptions, output expectations, skips and incompletes, and it prints what php prints". Everything
below was measured on that checkout with `php 8.4.27` as the reference; the expected output of every new fixture is
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
`phpx/testdata/patches/` (`REMOVED`, and overlaid test files). Each has a fixture under `phpx/testdata/` that
settles the new answer against php:

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

### Arrays copy on write

`model.Array` is a handle over storage, and the copy php makes at assignment waits for the first write. `Clone`
(`model/array.go`) answers a second handle over the same storage and marks both; every mutator calls `Split` first,
which copies the storage and sets a separate handle on each nested array. The first write to either side is the
only copy the pair pays, where the earlier runtime deep-copied at the assignment. `modelutil.CopyArray`, and
`CopyValue` through it, are that clone.

`phpx/testdata/arrays/copy_on_write.phpt` settles it against php on all three columns: the top-level copy, the
nested copy, a by-reference loop, a reference to an element, a builtin that writes in place, a binding that builds
a result, and both sides writing. The limits: the memory walk counts a clone's entries twice, because it dedups the
handle alone while the two handles read one storage through separate pointers; and a copy of an array that shares a
nested array with a value the host holds stays the host's until the host writes.

### Coverage

Go statement coverage of the fork-owned files (76 files, 3502 statements), from two sources:
`go test -tags phpx -coverpkg=./... ./...`, and a binary built with `go build -cover` that ran every fixture under
`phpx/testdata` and `tests/fixtures` and the five phpunit test files (`GOCOVERDIR`, then `go tool covdata textfmt`).
The last column is the union of the two. The fork-owned files are 86% covered together (52% by `go test`, 81%
by the fixtures); the whole module is 26465 statements. Measured 2026-11-04.

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

`phpx/testdata/packages/tests/` holds two phpunit files for each library, a first pass over the surface and a
`*MoreTest.php` that reaches the corners (`summary.sh <phpx>` prints the table, `compare.sh` diffs the full output).
php 8.4.27 passes all 120 tests; the third column is phpx. The first run of this table was 10 of 45.
`nikic/php-parser` is not in the set by choice. Measured 2026-11-18.

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

What the suites found and what fixed it: generators (`yield` ran on a goroutine; doctrine/inflector's rules and
symfony/finder iterate them), `array_map` keeping keys, `parse_url`, `http_build_query` with flags, the SPL file
iterators (`SplFileInfo`, `DirectoryIterator`, `RecursiveDirectoryIterator`, `WeakMap`), interface constants and
parents loaded through the autoloader, `!$x = f()`, a static closure's `self::`, possessive quantifiers, and enum
names in a namespace.

### Library coverage

A library passes when 80% of its statements run under phpx while the suites in `phpx/testdata/packages/tests/` run.
`phpx/testdata/packages/coverage.sh [phpx] [profile] [--covermode count|atomic]` builds the table; `coverage_launcher.php` runs every suite in
one phpunit process and then includes each library's source files, so a file no test reaches counts in the
denominator (php-code-coverage's whitelist does the same). `phpx/scripts/lib-coverage.py <profile> --files <library>`
lists a library's least covered files. The measure is phpx's own statement profile (`--cover`): php here has no
xdebug or pcov, so the php column of this table does not exist; php runs the same tests and passes all of them.

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

The blocker text lives in `phpx/testdata/packages/blockers.json`; `lib-coverage.py` prints it, and it is rewritten
with each iteration. 1111 tests run, php passes all of them; phpx fails about 180 of them (errors, failures,
incomplete, risky). The failures are kept in the files as regressions, and what they have in common is on the
"blockers by design" list: `__serialize`, `__sleep`, `__debugInfo`, `&` references outside `foreach` (so
`invokeArgs` with a reference and a promoted `int &$x`), and scalar parameter types, none of which phpx builds
([design](../docs/design.md)). Everything else the agents found was fixed (see CHANGELOG).

### PHPUnit

`phpx/testdata/phpunit/` holds a composer tree of `phpunit/phpunit ^11.5` (git ignores `vendor/`), a launcher, five
test files and `compare.sh`:

```sh
cd phpx/testdata/phpunit && composer install --ignore-platform-reqs
phpx run vendor/bin/phpunit --testdox tests/ProviderTest.php
./compare.sh phpx            # each tests/*Test.php under php and phpx, diffed
```

phpunit's entry script starts because the seven extension markers it gates on answer through Go bindings. The
launcher that calls `PHPUnit\TextUI\Application::run($argv)` directly stays for a tree that lacks one of them.
`php launcher.php tests/X.php` is the reference for that route.

| Test file | Covers | php | phpx |
|-----------|--------|-----|------|
| `ATest` | one assertion | `OK (1 test, 1 assertion)` | same |
| `CalcTest` | `setUp`, `expectException`, `assertCount`, a deliberate failure, `markTestSkipped` | `Tests: 7, Assertions: 21, Failures: 1, Skipped: 1.` | same |
| `ProviderTest` | `#[DataProvider]`, `#[TestWith]`, `#[Test]`, `#[CoversNothing]`, named and numbered data sets | `OK (7 tests, 7 assertions)` | same |
| `SkipOnlyTest` | `markTestSkipped`, `markTestIncomplete` | `Tests: 2, Assertions: 0, Skipped: 1, Incomplete: 1.` | same |
| `AssertionsTest` | objects, enums, traits, interfaces, strings, arrays, numbers, `#[Depends]`, output expectations, a diff on arrays, an uncaught `LogicException` | `Tests: 14, Assertions: 58, Errors: 1, Failures: 1.` | same |

## Open, in detail

- A throwable's file and line are not recorded, so a failure has no `file:line` under it and `getTrace()` is
  empty; `compare.sh` drops that line from both sides.
- `createMock()` and `createStub()` need `eval` of generated class source; `eval` is in the "Won't build" table.
- `serialize`, `unserialize`: phpunit uses them for process isolation only.
- the bytecode engine keeps upstream's answers where a fixture opts it out: a string offset write, a Go slice that
  is not a script array, and the array copies of `runner/phpx_value.go` (`flatstack: false` in
  `string_offset_write.phpt` and `explode_array.phpt`).
- `class_alias` does not feed `instanceof`; braced `namespace X { }` blocks do not parse; `feof()` on `php://memory`
  answers false until a read hits the end; `json_encode` writes `/` unescaped; `print_r` and `var_dump` do not print
  an enum as php does.
- an array copy shares the storage with its source until the first write, as php's does, but the memory walk counts
  each of the two handles' entries: it dedups the handle alone while the two handles read one storage.
- the extension markers phpunit's entry gate reads answer true for dom, libxml, xmlwriter and tokenizer through Go
  bindings, and the host's extensions stay off. The three `libxml_*` functions phpunit calls without a guard answer
  as an empty error buffer.

### Lint

`phpx/scripts/lint-delta.sh` still reports growth against upstream, and it was not brought to zero in this run.
The 2026-11-24 measurement returned 0 from the base worktree for every rule but `prose/shape`, so the growth
numbers name the fork's own files and not a regression: `godoc/internals` 137, `grouping/filename` 69 (the exported
`Phpx*` methods sit in files named for the area, not the symbol), `prose/placement` 26, `prose/ratio` 25,
`coverage/uncovered` 23, `pairing/orphan` 13, `func-args/order` 6, `godoc/outweighs` 6, `filecheck/long` 3. The
base worktree's splint reported one finding, so the comparison undercounts upstream and the gate is not part of
any test above.

### Upstream suites

A library's own phpunit suite is the widest check of PHP's behaviour this repository can run.
[testdata/upstream](testdata/upstream/README.md) clones a library, installs it with phpunit and runs its tests under
php and under phpx, so a difference in the totals is a difference in the language. The fixtures that come out of a
failing test live in the area they belong to (`functions`, `classes`, `syntax`), settled against php first.

| Library | php | phpx |
|---------|-----|------|
| pimple | 68 tests, 135 assertions | the same |
| doctrine/inflector | 1215 tests, 1238 assertions | the same |
| egulias/email-validator | 358 tests, 477 assertions | 472 assertions, five errors from the intl `Spoofchecker` |

### Upstream

Two pull requests against phpscript turn overlays into registrations; see [testdata/UPSTREAM.md](testdata/UPSTREAM.md).
