# The turbocov benchmark

[turbocov](https://github.com/shipmonk-rnd/turbocov) is a php extension that
collects line coverage in the engine, so a suite collects coverage with far
less overhead than Xdebug or PCOV. Its own benchmark runs the test suite of
`thephpleague/commonmark` and reports the suite's duration with each driver.

This directory reproduces that comparison for php, and puts phpx beside it: the
same checkout, the same PHPUnit 13.4.8, one php image with turbocov and Xdebug
built from source.

## What runs

`phpunit-plain.xml`, `phpunit-turbocov.xml` and `phpunit-xdebug.xml` are the
same configuration for the 4,447 tests of commonmark 2.10.7. Two of them ask
for a Clover report and a different driver:

| mode | command |
|------|---------|
| plain | `php vendor/bin/phpunit -c phpunit-plain.xml` |
| turbocov | `php -d turbocov.collect_paths=src vendor/bin/phpunit -c phpunit-turbocov.xml` |
| xdebug | `php -d zend_extension=xdebug.so -d xdebug.mode=coverage vendor/bin/phpunit -c phpunit-xdebug.xml` |
| phpx | `PHPX_PHP_BINARY=phpx-shim.sh phpx -f phpx.yml run vendor/bin/phpunit -c phpunit-plain.xml --testsuite <leg>` |
| phpx, coverage | `phpx -f phpx.yml --coverage --coverfile=phpscript.cov run vendor/bin/phpunit -c phpunit-plain.xml --testsuite <leg>` |

The phpx rows run the unit and functional testsuites as two processes and
report the sum; `mergecov.py` joins the two coverage profiles.

turbocov refuses to run while Xdebug is loaded, so the image enables turbocov
in `conf.d` and the Xdebug mode names its extension with `-d` for that leg
alone. Turbocov never runs under phpx: the extension boundary is php's, and
the comparison reads turbocov's Clover beside the profile phpx writes for the
same suite. `phpx.yml` turns telemetry off for both phpx legs.

## Running it

```sh
./run.sh              # three rounds of each php mode; prints the best of each
./run.sh 10           # ten rounds
PHPX_COVER=1 ./run.sh # also collects phpx's profiles, merged and compared with turbocov's Clover
PHPX=0 ./run.sh       # php arms alone; the phpx legs take minutes
PHPX_GOGC=100 ./run.sh # collect garbage on the default schedule instead of the soft limit
```

`run.sh` clones turbocov, builds phpx, stages both and the configuration into a
build context, builds the `phpx-turbocov` image and runs the four modes inside
it. The staging exists because the docker daemon of the bench host cannot see
the checkout's paths; nothing in the image reaches back to the host.

## Results

Measured on this machine (php 8.5.13, turbocov at the commit the image builds,
Xdebug 3.5.3, phpx at this checkout), best of three runs of the full 4,447-test
suite:

| mode | duration | over plain |
|------|---------:|-----------:|
| php, no coverage | 1.68 s | — |
| php + turbocov | 2.62 s | +56% |
| php + xdebug (coverage mode) | 24.56 s | +1,362% |

Every php run reports `Tests: 4447, Assertions: 5608` and writes the same
Clover report (488 KB). turbocov adds about 56% to the suite; Xdebug adds about
1,360%.

phpx cannot load a php extension, so it cannot run turbocov itself: turbocov
counts lines inside the php engine. What phpx runs is the same suite on its own
interpreter, and the same checkout of commonmark now passes whole:

| leg | tests | passing | errors | failures | wall |
|-----|------:|--------:|-------:|---------:|-----:|
| unit | 3,228 | 3,228 | 0 | 0 | 306.7 s |
| functional | 1,219 | 1,213 and 6 skipped | 0 | 0 | 74.0 s |
| both | 4,447 | 4,441 and 6 skipped | 0 | 0 | 380.7 s |

The two legs run in one process each. The whole suite in one process grows the
heap past the soft limit with the collector off, and the collector then runs
for the length of the unit leg; each leg alone stays inside it. The wall for
both is the sum of the two legs.

php measures 1.68 s for both suites. phpx interprets them, so the difference is
the interpreter and not the coverage: at about 0.086 s a test, the two legs'
380.7 s is about 145x the php run with turbocov and 15x the php run with
Xdebug.

## Coverage overhead, the same suite on each stack

The absolute wall times are the interpreter's; the question coverage asks is
what measurement adds to the runtime beneath it.

| stack | plain | covered | overhead |
|-------|------:|--------:|---------:|
| php, full suite | 1.68 s | 24.56 s (xdebug) | +1,362% |
| php + turbocov, full suite | 1.68 s | 2.62 s | +56% |
| phpx, full suite | 380.7 s | 394.9 s | +3.7% |
| phpx, functional leg | 74.0 s | 77.1 s | +4.2% |
| phpx, unit leg | 306.7 s | 317.8 s | +3.6% |
| phpx, 20M-statement loop | 8.65 s | 8.89 s | +2.7% |
| `phpx test`, fixture suite | 1.178 s runtime | 1.208 s runtime | +2.5% |

The unit collector figure is measured at the default collector; the plain unit
leg is measured with the collector off, which is the faster policy for the
functional leg. The two agree with the controlled loop and the fixture suite:
phpx's statement collector costs about 4% of the runtime it measures.

Relative to the runtime it measures, phpx's coverage is an order of magnitude
cheaper than turbocov's and two orders cheaper than Xdebug's. phpx pays for
being an interpreter: its full suite takes minutes where php takes 1.68 s,
about 145x the php+turbocov run and 15x the php+xdebug run.

## The generated profile beside Clover

`phpx --coverfile` writes a Go statement profile; turbocov writes PHPUnit's
Clover report. They count different units — a profile block carries a
statement and its line range, Clover carries a statement line — so the
comparison is per file, not per line.

The two profiles cover the same suite: the phpx profile is the merge of the
unit and functional legs (`mergecov.py`). On the 237 `src/` files both tools
recorded, 234 agree within five percentage points. The statement-weighted
totals are 97.0% of 4,251 Clover lines and 97.6% of 3,797 profile statements.

Three files differ by more than five points, and each names a unit difference
rather than a missed statement:

| file | clover | phpx | why |
|------|-------:|-----:|-----|
| `src/Util/SpecReader.php` | 0/23 | 15/16 | PHPUnit runs the spec reader while resolving data providers, before a php coverage driver starts; phpx covers the whole process |
| `src/Util/RegexHelper.php` | 36/49 | 38/39 | one phpx statement spans a multi-line call that Clover counts as several lines |
| `src/Extension/FrontMatter/Data/LibYamlFrontMatterParser.php` | 0/9 | 2/9 | the yaml extension is absent in the image, so neither tool runs the parser's body; the two register different statements |

## What this run changed

The profile of the 2026-11-27 run put 78% of the unit suite's CPU and 69% of
its 48.1 GB of allocation in the destructor's reachability walk. Four changes
came out of it:

- **A frame is a witness root.** A dropped value the caller still held in a
  local was walked from the globals on the next statement, because no witness
  could name a frame local. `phpxWitnessRoots` now visits the live frames,
  innermost first, with a per-push frame id so a pooled scope answers for
  neither. The unit leg goes from 356.0 s to 306.7 s.
- **The mark sets are pooled, with the cap above one walk.** The three identity
  maps a walk dedups with are cleared and reused rather than allocated per
  drop. Allocation for the unit leg falls from 48.1 GB to 10.7 GB.
- **The collector is off for a measured run.** `GOGC=off` with
  `GOMEMLIMIT=6GiB` takes the functional leg from 93.3 s to 74.0 s, about 21%.
- **Two allocation fixes from the profile.** `model.Array.promote` boxes an
  implicit key once for the map and the order slice (493 to 249 allocs for a
  500-key promote), and `__phpx_refarg` takes its three parameters typed
  instead of variadic (80 B and 3 allocs to 16 B and 1), removing 2.28 GB from
  the suite's allocation profile.

What is left in the profile is the interpreter itself: `Array.promote`,
`DeepSize` (PHPUnit's `memory_get_usage`), `asciiLower`, and the per-call
argument slice in `expr.baseCall`. None of them is a coverage cost.

## What the measurement found

Three bugs came out of it, each pinned by a test:

- `declare(strict_types=1);` was registered as an executable statement, so
  every file opening with it reported a covered line php does not count, and a
  file holding only an interface reported 100%. A compile-time declare records
  no block now; a `ticks` directive and the block form still do.
- `--covermode`'s default was written back onto the option, and `phpx test`
  read that as the operator naming the mode: every test run collected coverage
  and wrote `phpscript.cov`, including the "uncovered" arm of the testcov
  benchmark. `Check` leaves the mode as written.
- An absolute path argument left each `.phpt` entrypoint counted as a covered
  source file; `fixtureCoverName` normalizes both sides.

## The destructor benchmark

`phpx/bench/turbocov` measures the suite; the destructor work it needed is
measured by the scripts under `phpx/bench/turbocov/destructor/`. `graph.php`
builds a tree of 4,000 roots with 20 children each and drops every root, which
is one reachability walk per root:

```sh
./destructor/bench.sh            # php and phpx, three sizes
```

| nodes | phpx before | phpx now | php |
|-------|-------------|----------|-----|
| 1,000 | 3.32 s | 0.17 s | 0.04 s |
| 2,000 | 13.6 s | 0.31 s | 0.04 s |
| 4,000 | 52.9 s | 0.64 s | 0.05 s |

"before" is commit `e7f859e1`, the parent of the commit that caches what a
subtree reaches; "now" is the tree at HEAD.

The profile comparison reads the Clover report turbocov writes and the
profile phpx writes from the same checkout:

```sh
php -d zend_extension=xdebug.so -d xdebug.mode=coverage vendor/bin/phpunit \
    -c phpunit-xdebug.xml --testsuite 'league/commonmark Functional Tests'
phpx --coverfile=/tmp/phpx.cov run vendor/bin/phpunit -c phpunit-plain.xml \
    --testsuite 'league/commonmark Functional Tests'
./compare.py build/logs/clover.xml /tmp/phpx.cov
```
