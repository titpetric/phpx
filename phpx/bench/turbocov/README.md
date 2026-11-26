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
./run.sh        # three rounds of each mode; prints the best of each
./run.sh 10     # ten rounds
```

`run.sh` clones turbocov, builds phpx, stages both and the configuration into a
build context, builds the `phpx-turbocov` image and runs the four modes inside
it. The staging exists because the docker daemon of the bench host cannot see
the checkout's paths; nothing in the image reaches back to the host.

## Results

Measured on this machine (php 8.5.13, turbocov at the commit the image builds,
Xdebug 3.5.3, phpx at this checkout), best of three runs:

| mode | duration |
|------|----------|
| php, no coverage | 1.69 s |
| php + turbocov | 2.69 s |
| php + xdebug (coverage mode) | 25.8 s |

Every php run reports `Tests: 4447, Assertions: 5608` and writes the same
Clover report (488 KB). turbocov adds about 60% to the suite; Xdebug adds about
1,400%.

phpx cannot load a php extension, so it cannot run turbocov itself: turbocov
counts lines inside the php engine. What phpx runs is the same suite on its own
interpreter, and the same checkout of commonmark gives:

| suite | tests | passing | errors | failures | wall |
|-------|-------|---------|--------|----------|------|
| unit | 3,228 | 3,129 | 18 | 81 | 397 s |
| functional | 1,219 | 1,179 | 14 | 20 (6 skipped) | 109 s |

php measures 1.69 s for both suites together. phpx interprets them, so the
difference is the interpreter and not the coverage: at ~0.12 s a test it is
about 190x the php run with turbocov and 20x the php run with Xdebug.

phpx collects statement coverage itself, with no extension: the functional
suite takes 111 s with `--cover` against 109 s without, and the profile of
`src/` reports 36.4% of its statements covered by that suite.

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
