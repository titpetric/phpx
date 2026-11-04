# phpx test report, 2026-11-04

The run behind this report took the fork from "phpunit stops on a parse error at `Application.php:835`" to
"phpunit 11.5 runs from its source tree: test discovery, `setUp`, assertions, `#[DataProvider]`, `#[TestWith]`,
`#[Depends]`, exceptions, output expectations, skips and incompletes, and it prints what php prints". Everything
below was measured on this checkout with `php 8.4.27` as the reference; the expected output of every new fixture is
`php`'s own, not phpx's.

## Before and after

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

## What landed

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

## Coverage

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

## The nine composer libraries, run

`phpx/testdata/packages/tests/` holds a phpunit suite for each library (`summary.sh <phpx>` prints the table, `compare.sh`
diffs the full output). php 8.4.27 passes all 45 tests; the third column is phpx. The first run of this table was 10 of 45. `nikic/php-parser` is not in the set by choice.

| Library | Tests | php | phpx | What stops the rest |
|---|--:|--:|--:|---|
| `psr/log` | 4 | 4 | **4** | |
| `psr/http-message` | 3 | 3 | **3** | |
| `psr/container` | 3 | 3 | **3** | |
| `guzzlehttp/psr7` | 5 | 5 | **5** | |
| `guzzlehttp/guzzle` | 5 | 5 | **5** | |
| `doctrine/inflector` | 9 | 9 | **9** | |
| `symfony/finder` | 5 | 5 | **5** | |
| `symfony/console` | 5 | 5 | **5** | |
| `monolog/monolog` | 6 | 6 | **6** | |
| **total** | **45** | **45** | **45** | 100% |

What the suites found and what fixed it: generators (`yield` ran on a goroutine; doctrine/inflector's rules and symfony/finder
iterate them), `array_map` keeping keys, `parse_url`, `http_build_query` with flags, the SPL file iterators
(`SplFileInfo`, `DirectoryIterator`, `RecursiveDirectoryIterator`, `WeakMap`), interface constants and parents loaded through the
autoloader, `!$x = f()`, a static closure's `self::`, possessive quantifiers, and enum names in a namespace.

## PHPUnit

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

## Open

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

## Lint

`phpx/scripts/lint-delta.sh` still reports growth against upstream, and it was not brought to zero in this run:
`grouping/filename` (the exported `Phpx*` methods sit in files named for the area, not the symbol: 62 before this run, 63 now),
`func-args/order` (4, three from earlier fork files), `coverage/uncovered` (5), `godoc/outweighs` (+1) and one
`filecheck/long`. `imports/format` and `pairing/unpaired` were fixed at the end (`splint-fix.sh`, and a test file beside
each new fork file). The gate is not part of any test above.

## Upstream

Two pull requests against phpscript turn overlays into registrations; see [testdata/UPSTREAM.md](testdata/UPSTREAM.md).
