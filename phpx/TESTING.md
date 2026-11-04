# phpx test report, 2026-11-04

The run behind this report took the fork from "phpunit stops on a parse error at `Application.php:835`" to
"phpunit 11.5 starts from its source tree, discovers a test class, runs seven tests and reports a pass, a
failure and a skip". Everything below was measured on this checkout with `php 8.4.27` as the reference; the
expected output of every new fixture is `php`'s own, not phpx's.

## Before and after

| Measure | Before (`e1e250e`) | After |
|---------|-------------------:|------:|
| phpx fixtures (`phpx test --matrix phpx/testdata/...`, flatstack, runtime and php) | 47 of 47 | **70 of 70** |
| upstream fixtures (`phpx test --matrix tests/fixtures/...`) | 290 of 290 | **288 of 288** (2 dropped, see below) |
| `go test -tags phpx ./...` | green | green |
| phpunit source files that parse, `phpunit/phpunit/src` (pass / warn / fail) | 795 / 483 / **25** | 894 / 101 / **0** |
| whole phpunit install (phpunit, sebastian, phar-io, myclabs, staabm, theseer) | 1025 / 746 / **36** | 1171 / 217 / **0** |
| top ten composer libraries, failing files | **71** | **18** |
| top ten composer libraries, passing files | 438 | 535 |
| `phpx run launcher.php tests/ATest.php` | parse error, nothing runs | `OK (1 test, 1 assertion)` |
| `phpx run launcher.php tests/CalcTest.php` | nothing runs | 7 tests, 21 assertions: 5 pass, 1 failure, 1 skip (see "Open") |
| undefined functions called by the phpunit tree | `assert` x318, `serialize`, `current`, `class_implements`, ... | extension functions only (`libxml_*`, `proc_*`, `xdebug_*`, `pcntl_*`) and `serialize`, `unserialize`, `eval` |

The ten libraries, per package (passing / warning / failing files):

| Package | Before | After |
|---------|-------:|------:|
| `guzzlehttp/psr7` | 11 / 49 / 4 | 13 / 55 / 0 |
| `guzzlehttp/guzzle` | 27 / 74 / 7 | 31 / 102 / 1 |
| `symfony/console` | 55 / 133 / 29 | 75 / 163 / 15 |
| `symfony/finder` | 14 / 9 / 4 | 16 / 15 / 2 |
| `monolog/monolog` | 35 / 305 / 16 | 75 / 175 / 0 |
| `nikic/php-parser` | 231 / 59 / 10 | 258 / 37 / 0 |
| `doctrine/inflector` | 47 / 2 / 1 | 49 / 0 / 0 |
| `psr/*` (3 packages) | all pass | all pass |

Two upstream fixtures and four Go tests pinned decisions this fork reverses and are dropped or reworded through
`phpx/testdata/patches/` (`REMOVED`, and overlaid test files): `tests/fixtures/arrays/unset_native_slice_reindexes.phpt`
(a native slice renumbers on unset), `tests/fixtures/stdlib/json_encode_flags.phpt` (no `JSON_*` constant),
`TestNoJSONFlagConstants`, `TestJSONFlagConstantThrows`, and the two `tests/bindings_test.go` cases that pinned a
slice's limits. Each has a fixture under `phpx/testdata/` that settles the new answer against php.

## What landed

Syntax, each with a fixture whose expected output is php's:

| Feature | Seam | Fixture |
|---------|------|---------|
| `[$a, [$b, $c], 'k' => $d] = ...`, `foreach (... as [$a, $b])`, skipped slots | `parser/phpx_destructure.go` | `functions/destructuring_test.php` |
| `<<<EOT`, `<<<'EOT'`, closing-line indentation | `parser/phpx_heredoc.go` | `syntax/heredoc_test.php` |
| `and`, `or`, `xor`; `for (;;)`, `for ($i = 0, $c = n; ...; $i++, $j--)`; `a ?? b = c` | `parser/phpx_logical.go` | `syntax/word_operators_test.php` |
| assignment expressions on index, property and static targets, compound operators | `runner/expr/phpx_assign.go`, `runner/phpx_assign.go` | `syntax/word_operators_test.php` |
| named arguments (user functions, constructors, a table of built-ins) | `parser/phpx_named.go`, `runner/phpx_named.go` | `syntax/named_arguments_test.php` |
| `[...$a, ...$b]` | `runner/expr/phpx_spread_array.go` | `syntax/named_arguments_test.php` |
| `print` | `parser/phpx_logical.go` | `syntax/print_test.php` |
| `match`, `throw` as an expression | `parser/phpx_match.go` | `syntax/match_throw_test.php` |
| `clone`, `__clone`, object `==` and `<` | `parser/phpx_match.go`, `runner/phpx_spl.go`, `internal/phpval/compare.go` | `classes/clone_test.php`, `classes/object_compare_test.php` |
| `enum`, backed and pure, `cases`, `from`, `tryFrom`, methods, constants, interfaces | `parser/phpx_enum.go`, `stdlib/phpx/enums.go` | `classes/enum_test.php` |
| `use A\{B, C as D}` | `parser/phpx_usegroup.go` | `syntax/group_use_test.php` |
| bare `{ ... }` block | `parser/parser.go` | `syntax/bare_block_test.php` |
| `$cls::NAME`, `$cls::$prop`, `$cls::method()`, `$cls::$method()` | `parser/phpx_classof.go` | `classes/static_on_value_test.php` |
| `$s[0] = 'x'` | `runner/phpx_stroffset.go` | `syntax/string_offset_write.phpt` |

Runtime and library:

| Feature | Seam | Fixture |
|---------|------|---------|
| `Reflection{Class,Object,Method,Function,Property,Parameter,NamedType}`, `ReflectionException` (a PHP prelude over Go primitives; attributes and doc comments read as absent) | `runner/phpx_prelude/reflection.php`, `runner/phpx_reflect.go` | `classes/reflection_test.php` |
| `SplObjectStorage`, `SplStack`, `SplQueue`, `SplDoublyLinkedList`, `SplFixedArray`, `ArrayObject`, `ArrayIterator`, `SplObserver`, `SplSubject` | `runner/phpx_prelude/spl.php` | `classes/spl_structures_test.php` |
| `Iterator`, `IteratorAggregate`, `Countable`, `ArrayAccess` on script classes; `iterator_to_array`, `iterator_count` | `runner/phpx_spl.go`, `stdlib/phpx/spl.go` | `classes/spl_interfaces_test.php` |
| `new Exception($m, $c, $previous)`, `getPrevious`, `getTraceAsString`, `$e instanceof Throwable` | `stdlib/phpx/exceptions.go`, `runner/phpx_instanceof.go` | `classes/exception_previous_test.php` |
| `catch (Name)` inside a namespace matches by last segment | `runner/phpx_inheritance.go` | `classes/catch_namespaced_test.php` |
| `parent::` through a three-level chain no longer recurses | `runner/phpx_inheritance.go` | `classes/parent_chain_test.php` |
| `class_implements`, `class_alias`, `constant('A::B')` with inherited constants | `runner/phpx_classinfo.go`, `stdlib/phpx/reflection.go` | `functions/class_info_test.php` |
| `current`, `key`, `next`, `prev`, `reset`, `end` | `stdlib/phpx/pointer.go` | `functions/array_pointer_test.php` |
| a Go slice a binding answers becomes a script array where a variable or parameter takes it | `runner/phpx_named.go` | `functions/explode_array.phpt` |
| `E_*`, `JSON_*`, `SORT_*`, `LC_*`, `PHP_URL_*` constants | `stdlib/phpx/constants.go` | `functions/json_constants_test.php` |
| `fopen('php://stdout')`, `php://stderr` | `stdlib/files/phpx_wrappers.go` | exercised by the phpunit run |
| `assert`, `printf`, `gc_*`, `setlocale`, `base_convert`, `array_diff_assoc`, ... | `stdlib/phpx/runtime_funcs.go` | exercised by the phpunit run |
| `use function hrtime;` no longer shadows a class `HRTime` | `parser/parser.go` | `functions/class_info_test.php` |
| a bare constant in a ternary arm or the right of `&&`/`||` reads when it runs | `runner/expr/phpx_assign.go` | exercised by the phpunit run |

## Coverage

Go statement coverage of the fork-owned files, from two sources: `go test -tags phpx -coverpkg=./... ./...`, and a
binary built with `go build -cover` that ran every fixture under `phpx/testdata` and `tests/fixtures` and the phpunit
launcher (`GOCOVERDIR`, then `go tool covdata textfmt`). The last column is the union of the two.

| File | Statements | go test | fixtures | together |
|------|-----------:|--------:|---------:|---------:|
| `formatter/phpx_traits.go` | 64 | 98% | 0% | 98% |
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
| `runner/expr/phpx_spread.go` | 108 | 27% | 93% | 93% |
| `runner/expr/phpx_spread_array.go` | 46 | 11% | 78% | 78% |
| `runner/phpx.go` | 10 | 100% | 80% | 100% |
| `runner/phpx_abstract.go` | 47 | 81% | 89% | 91% |
| `runner/phpx_assign.go` | 13 | 23% | 54% | 54% |
| `runner/phpx_called.go` | 28 | 57% | 82% | 82% |
| `runner/phpx_classinfo.go` | 92 | 2% | 88% | 88% |
| `runner/phpx_closure.go` | 36 | 28% | 86% | 86% |
| `runner/phpx_funcdecl.go` | 9 | 89% | 78% | 89% |
| `runner/phpx_inheritance.go` | 152 | 84% | 87% | 91% |
| `runner/phpx_instanceof.go` | 18 | 17% | 100% | 100% |
| `runner/phpx_interfaces.go` | 58 | 90% | 95% | 95% |
| `runner/phpx_isa.go` | 15 | 60% | 93% | 100% |
| `runner/phpx_named.go` | 71 | 30% | 87% | 87% |
| `runner/phpx_prelude.go` | 15 | 33% | 67% | 67% |
| `runner/phpx_promotion.go` | 31 | 100% | 97% | 100% |
| `runner/phpx_properties.go` | 107 | 80% | 93% | 93% |
| `runner/phpx_reflect.go` | 182 | 1% | 77% | 77% |
| `runner/phpx_spl.go` | 110 | 16% | 77% | 77% |
| `runner/phpx_statics.go` | 25 | 88% | 88% | 92% |
| `runner/phpx_stringer.go` | 21 | 48% | 86% | 86% |
| `runner/phpx_stroffset.go` | 34 | 9% | 59% | 59% |
| `runner/phpx_throwable.go` | 37 | 89% | 89% | 89% |
| `runner/phpx_traitrules.go` | 47 | 98% | 89% | 98% |
| `runner/phpx_traits.go` | 181 | 80% | 78% | 82% |
| `runner/phpx_vendor.go` | 10 | 80% | 0% | 80% |
| `runner/phpx_visibility.go` | 31 | 94% | 97% | 97% |
| `stdlib/core/phpx_called.go` | 3 | 67% | 100% | 100% |
| `stdlib/core/phpx_isa.go` | 3 | 67% | 100% | 100% |
| `stdlib/core/phpx_json.go` | 37 | 81% | 81% | 81% |
| `stdlib/core/phpx_version_compare.go` | 69 | 3% | 90% | 90% |
| `stdlib/files/phpx_wrappers.go` | 4 | 100% | 75% | 100% |
| `stdlib/phpx/arrays.go` | 22 | 14% | 100% | 100% |
| `stdlib/phpx/classes.go` | 41 | 76% | 93% | 93% |
| `stdlib/phpx/clock.go` | 26 | 19% | 92% | 92% |
| `stdlib/phpx/constants.go` | 4 | 100% | 100% | 100% |
| `stdlib/phpx/ctype.go` | 17 | 24% | 100% | 100% |
| `stdlib/phpx/enums.go` | 44 | 9% | 84% | 84% |
| `stdlib/phpx/environment.go` | 24 | 46% | 92% | 92% |
| `stdlib/phpx/exceptions.go` | 13 | 62% | 69% | 69% |
| `stdlib/phpx/logical.go` | 22 | 36% | 91% | 91% |
| `stdlib/phpx/pointer.go` | 54 | 63% | 91% | 91% |
| `stdlib/phpx/reflection.go` | 40 | 38% | 75% | 78% |
| `stdlib/phpx/runtime_funcs.go` | 134 | 34% | 43% | 43% |
| `stdlib/phpx/spaceship.go` | 11 | 18% | 100% | 100% |
| `stdlib/phpx/spl.go` | 57 | 35% | 68% | 68% |
| `stdlib/phpx/strings.go` | 51 | 18% | 88% | 88% |
| `stdlib/phpx/types.go` | 40 | 25% | 85% | 85% |
| `stdlib/phpx/version.go` | 9 | 100% | 89% | 100% |
| **fork-owned files** | 3055 | 55% | 80% | 85% |
| whole module | 26010 | 84% | 59% | 88% |

The fork-owned files are 3055 statements, 85% covered together. `runner/phpx_reflect.go` and `runner/phpx_classinfo.go`
are covered through the fixtures only, and `stdlib/phpx/*.go` through the fixtures and the phpunit run.

## PHPUnit

`phpx/testdata/phpunit/` holds a composer tree of `phpunit/phpunit ^11.5` (git ignores `vendor/`), a launcher, and two
test files:

```sh
cd phpx/testdata/phpunit && composer install --ignore-platform-reqs
phpx run launcher.php --testdox tests/CalcTest.php
```

The launcher skips phpunit's entry script, which refuses to start without dom, libxml, tokenizer and xmlwriter; those
extensions stay out of scope. `php launcher.php tests/CalcTest.php` is the reference.

| | php 8.4.27 | phpx |
|---|---|---|
| progress line | `.....FS` | `.....FR` |
| summary | `Tests: 7, Assertions: 21, Failures: 1, Skipped: 1.` | `Tests: 7, Assertions: 21, Failures: 1, Skipped: 1, Risky: 1.` |
| `ATest` (one assertion) | `OK (1 test, 1 assertion)` | `OK (1 test, 1 assertion)` |

## Open

- `testSkipped` is reported skipped and also risky ("did not perform any assertions"): `TestRunner` should see
  `$e instanceof SkippedTest` for the rethrown `SkippedWithMessageException` and skip the check. The same instanceof
  holds in a standalone script, so the difference is somewhere in the runner's path. One test in the file shows `?` in
  testdox where php shows a check mark.
- Attributes (`#[DataProvider]`, `#[Test]`) and doc comments are not parsed, so phpunit data providers, `@dataProvider`
  and `#[Test]` are absent. `getAttributes()` answers `[]`.
- `createMock()` needs `eval` of generated class source; `eval` is in the "Won't build" table.
- `serialize`, `unserialize`: phpunit uses them for process isolation only.
- the bytecode engine keeps upstream's answers where a fixture opts it out: a string offset write, and a Go slice
  that is not a script array (`flatstack: false` in `string_offset_write.phpt` and `explode_array.phpt`).
- `class_alias` does not feed `instanceof`; braced `namespace X { }` blocks do not parse; `php://memory` stays closed
  because an upstream test pins it.
- extension walls (dom, libxml, xmlwriter, tokenizer) are out of scope and are not listed as blockers.

## Upstream

Two pull requests against phpscript turn overlays into registrations; see [testdata/UPSTREAM.md](testdata/UPSTREAM.md).
