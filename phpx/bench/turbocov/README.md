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
| phpx | `PHPX_PHP_BINARY=phpx-shim.sh phpx run vendor/bin/phpunit -c phpunit-plain.xml` |

turbocov refuses to run while Xdebug is loaded, so the image enables turbocov
in `conf.d` and the Xdebug mode names its extension with `-d` for that leg
alone.

## Running it

```sh
./run.sh              # three rounds of each php mode; prints the best of each
./run.sh 10           # ten rounds
PHPX_COVER=1 ./run.sh # also runs phpx with --coverfile
PHPX=0 ./run.sh       # php arms alone; the phpx legs take minutes
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
| php, no coverage | 1.66 s | — |
| php + turbocov | 2.59 s | +56% |
| php + xdebug (coverage mode) | 24.38 s | +1,366% |

Every php run reports `Tests: 4447, Assertions: 5608` and writes the same
Clover report (488 KB). turbocov adds about 56% to the suite; Xdebug adds about
1,370%.

phpx cannot load a php extension, so it cannot run turbocov itself: turbocov
counts lines inside the php engine. What phpx runs is the same suite on its own
interpreter, and the same checkout of commonmark gives:

| suite | tests | passing | errors | failures | wall |
|-------|-------|---------|--------|----------|------|
| unit | 3,228 | 3,129 | 18 | 81 | 397 s |
| functional | 1,219 | 1,179 | 14 | 20 (6 skipped) | 109 s |

php measures 1.66 s for both suites together. phpx interprets them, so the
difference is the interpreter and not the coverage: at ~0.12 s a test, the two
suites' 506 s is about 195x the php run with turbocov and 20x the php run with
Xdebug.

## Coverage overhead, the same suite on each stack

The absolute wall times are the interpreter's; the question coverage asks is
what measurement adds to the runtime beneath it.

| stack | plain | covered | overhead |
|-------|------:|--------:|---------:|
| php + turbocov, functional suite | 0.78 s | 1.54 s | +97% |
| php + xdebug, functional suite | 0.78 s | 10.56 s | +1,256% |
| phpx, 20M-statement loop | 8.65 s | 8.89 s | +2.7% |
| `phpx test`, fixture suite | 1.178 s runtime | 1.208 s runtime | +2.5% |

The commonmark functional suite is too noisy to separate the modes: eight
interleaved phpx runs took 105.2–118.6 s without coverage and 105.5–126.2 s
with it, means 114.4 s and 112.7 s. The controlled loop and the fixture suite
are small enough to sit still, and both put the collector at 2.5–3%.

Relative to the runtime it measures, phpx's coverage is about an order of
magnitude cheaper than turbocov's and two orders cheaper than Xdebug's. The
target the comparison set was turbocov's 56–97%; phpx is ahead of it. phpx
pays for being an interpreter: its full suite takes minutes where php takes
1.66 s, about 195x the php+turbocov run and 20x the php+xdebug run.

## The generated profile beside Clover

`phpx --coverfile` writes a Go statement profile; turbocov writes PHPUnit's
Clover report. They count different units — a profile block carries a
statement and its line range, Clover carries a statement line — so the
comparison is per file, not per line.

On the functional suite, for the 234 `src/` files both tools recorded: 227
agree within five percentage points, and the statement-weighted totals are
92.3% of 4,225 Clover lines and 91.5% of 3,772 profile statements. The files
that differ are the tests that fail under phpx (the embed extension and the
Nette-Schema configuration cases) and the code PHPUnit runs while resolving
data providers: phpx measures it because its collector covers the whole
process, while a PHP coverage driver starts after discovery.

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
