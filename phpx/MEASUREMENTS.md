# Measurements, one test unit at a time

Every unit below ran alone through a coverage build of phpx (`go build -cover -covermode=atomic -coverpkg=./... -tags phpx`)
with its own `GOCOVERDIR`, so each statement count belongs to that unit. `phpx/scripts/measure.py <binary>` regenerates the
table; [docs/guides](../docs/guides/how-to-collect-go-code-coverage-from-phpscript-phpx-test.md) explains the build. A unit is
a fixture area (`phpx test --matrix <dir>/...`, interpreter and bytecode engine) or one phpunit test file.

Columns: **Passed** is passes over tests (fixtures; for phpunit, tests minus failures, errors, skips and incompletes).
**Go statements covered** is every statement of the module the unit executed, **fork-owned** the part of that in `phpx*.go`
files and `stdlib/phpx/`, and **% of module** the share of the about 27,000 instrumented statements.

Read the phpunit rows with the deliberate cases in mind: `AssertionsTest` carries one failure and one error, `CalcTest` a
failure and a skip, `FunctionsTest` a failure, because php reports the same; No row's shortfall is a gap now. Seconds for a phpunit file are almost all phpunit's own start-up (3 seconds), not
the test.

| Unit | Kind | Passed | Seconds | Go statements covered | of which fork-owned | % of module |
|---|---|---:|---:|---:|---:|---:|
| `phpx/testdata/abstract` | fixtures | 7/7 | 0.0 | 3716 | 590 | 13.6% |
| `phpx/testdata/classes` | fixtures | 37/37 | 0.1 | 8425 | 2342 | 30.8% |
| `phpx/testdata/constants` | fixtures | 3/3 | 0.0 | 4418 | 536 | 16.1% |
| `phpx/testdata/functions` | fixtures | 23/23 | 0.0 | 7744 | 1396 | 28.3% |
| `phpx/testdata/globals` | fixtures | 1/1 | 0.0 | 3288 | 375 | 12.0% |
| `phpx/testdata/inheritance` | fixtures | 1/1 | 0.0 | 3976 | 615 | 14.5% |
| `phpx/testdata/interfaces` | fixtures | 1/1 | 0.0 | 3694 | 626 | 13.5% |
| `phpx/testdata/stdclass` | fixtures | 1/1 | 0.0 | 4085 | 486 | 14.9% |
| `phpx/testdata/syntax` | fixtures | 9/9 | 0.0 | 6747 | 1110 | 24.6% |
| `phpx/testdata/traits` | fixtures | 5/5 | 0.0 | 4233 | 790 | 15.5% |
| `tests/fixtures/arithmetic` | fixtures | 25/25 | 0.0 | 4678 | 393 | 17.1% |
| `tests/fixtures/arrays` | fixtures | 25/25 | 0.0 | 6240 | 602 | 22.8% |
| `tests/fixtures/autoloading` | fixtures | 4/4 | 0.0 | 5193 | 552 | 19.0% |
| `tests/fixtures/bindings` | fixtures | 37/37 | 0.0 | 7220 | 562 | 26.4% |
| `tests/fixtures/comparison` | fixtures | 1/1 | 0.0 | 3791 | 353 | 13.8% |
| `tests/fixtures/errors` | fixtures | 3/3 | 0.0 | 4050 | 370 | 14.8% |
| `tests/fixtures/exceptions` | fixtures | 10/10 | 0.0 | 4465 | 392 | 16.3% |
| `tests/fixtures/flatstack` | fixtures | 8/8 | 0.0 | 5584 | 548 | 20.4% |
| `tests/fixtures/functions` | fixtures | 19/19 | 0.0 | 5975 | 590 | 21.8% |
| `tests/fixtures/gd` | fixtures | 8/8 | 0.2 | 5456 | 530 | 19.9% |
| `tests/fixtures/github` | fixtures | n/a | 0.0 | 474 | 39 | 1.7% |
| `tests/fixtures/includes` | fixtures | 5/5 | 0.0 | 4342 | 447 | 15.9% |
| `tests/fixtures/mail` | fixtures | 3/3 | 0.0 | 4233 | 398 | 15.5% |
| `tests/fixtures/namespaces` | fixtures | 5/5 | 0.0 | 4774 | 591 | 17.4% |
| `tests/fixtures/oop` | fixtures | 24/24 | 0.0 | 6798 | 782 | 24.8% |
| `tests/fixtures/output` | fixtures | 4/4 | 0.0 | 4144 | 425 | 15.1% |
| `tests/fixtures/paths` | fixtures | 8/8 | 0.0 | 4913 | 481 | 17.9% |
| `tests/fixtures/pexec` | fixtures | 3/3 | 0.0 | 3805 | 352 | 13.9% |
| `tests/fixtures/regex` | fixtures | 10/10 | 0.0 | 4541 | 416 | 16.6% |
| `tests/fixtures/routing` | fixtures | n/a | 0.0 | 474 | 39 | 1.7% |
| `tests/fixtures/runtime` | fixtures | 14/14 | 0.0 | 5466 | 502 | 20.0% |
| `tests/fixtures/scaffold` | fixtures | 4/4 | 0.0 | 4207 | 370 | 15.4% |
| `tests/fixtures/sessions` | fixtures | 1/1 | 0.0 | 3832 | 353 | 14.0% |
| `tests/fixtures/stdlib` | fixtures | 27/27 | 0.1 | 6567 | 466 | 24.0% |
| `tests/fixtures/strings` | fixtures | 25/25 | 0.0 | 6611 | 586 | 24.1% |
| `tests/fixtures/syntax` | fixtures | 9/9 | 0.0 | 5877 | 524 | 21.5% |
| `tests/fixtures/types` | fixtures | 5/5 | 0.0 | 5121 | 455 | 18.7% |
| `phpx/testdata/phpunit/tests/ATest.php` | phpunit | 1/1 | 2.9 | 6771 | 1524 | 24.7% |
| `phpx/testdata/phpunit/tests/AssertionsTest.php` | phpunit | 12/14 | 3.9 | 7967 | 2170 | 29.1% |
| `phpx/testdata/phpunit/tests/CalcTest.php` | phpunit | 5/7 | 3.3 | 7107 | 1596 | 26.0% |
| `phpx/testdata/phpunit/tests/EnumMatchTest.php` | phpunit | 5/5 | 3.1 | 7084 | 1780 | 25.9% |
| `phpx/testdata/phpunit/tests/FunctionsTest.php` | phpunit | 8/9 | 3.4 | 7390 | 1687 | 27.0% |
| `phpx/testdata/phpunit/tests/OopTest.php` | phpunit | 6/6 | 3.2 | 7005 | 1705 | 25.6% |
| `phpx/testdata/phpunit/tests/ProviderTest.php` | phpunit | 7/7 | 3.3 | 7025 | 1624 | 25.7% |
| `phpx/testdata/phpunit/tests/SkipOnlyTest.php` | phpunit | 0/2 | 2.9 | 6847 | 1554 | 25.0% |
| `phpx/testdata/packages/tests/ConsoleTest.php` | phpunit | 5/5 | 3.3 | 7965 | 2168 | 29.1% |
| `phpx/testdata/packages/tests/FinderTest.php` | phpunit | 5/5 | 3.2 | 7125 | 1670 | 26.0% |
| `phpx/testdata/packages/tests/GuzzleTest.php` | phpunit | 5/5 | 3.3 | 7615 | 1933 | 27.8% |
| `phpx/testdata/packages/tests/InflectorTest.php` | phpunit | 9/9 | 3.5 | 7268 | 1785 | 26.5% |
| `phpx/testdata/packages/tests/MonologTest.php` | phpunit | 6/6 | 3.2 | 7600 | 2039 | 27.8% |
| `phpx/testdata/packages/tests/Psr7Test.php` | phpunit | 5/5 | 3.2 | 7333 | 1850 | 26.8% |
| `phpx/testdata/packages/tests/PsrContainerTest.php` | phpunit | 3/3 | 3.0 | 7005 | 1659 | 25.6% |
| `phpx/testdata/packages/tests/PsrHttpMessageTest.php` | phpunit | 3/3 | 3.0 | 6927 | 1638 | 25.3% |
| `phpx/testdata/packages/tests/PsrLogTest.php` | phpunit | 4/4 | 3.1 | 7090 | 1791 | 25.9% |

## Reading it

- **Breadth:** `phpx/testdata/classes` (36 fixtures) reaches 30.5% of the module and the most fork-owned code of any fixture
  area (2242 statements); `tests/fixtures/bindings` and `stdlib` reach the most of upstream's library.
- **Cost:** a fixture area runs in under 0.3 seconds on both engines; phpunit units take about 3 seconds each, so the 8 + 10
  phpunit files are the pipeline's slow part.
- **Overlap:** every phpunit unit covers 25 to 29% of the module and a fork-owned 1.5 to 2.2 thousand statements: they
  mostly run the same interpreter paths, which is why the union of all units is 88% and not the sum.
- **Empty units:** `tests/fixtures/github` and `tests/fixtures/routing` hold no `.phpt` fixtures, and their rows are the
  start-up cost alone (473 statements).
