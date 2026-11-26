# testcov

The test suite under statement coverage, in a container. It answers two
questions: what does `phpx test` cost with coverage, and do the recording
modes agree on what ran?

`./run.sh [--covermode "atomic count"]` builds `bin/phpx`, stages the binary
and the fixture trees into a small image, and runs the workload once to warm
the page cache and five times per mode, printing every time and keeping the
best:

- **uncovered**: `phpx test --skip-php --matrix` over the coverage pipeline's
  fixture set: `./tests/...` and every area under `phpx/testdata` except
  `patches/`, which is the overlay and not a fixture tree.
- **atomic** and **count**: the same run with `--cover --covermode=<mode>`.

The table carries the run's own summary and, for each mode, the coverage the
generated profile holds: files and statements of the source the fixtures load,
with the `.phpt` entrypoints left out because they run by definition. It then
diffs the two profiles line for line. A mode may change the cost and the count
a line carries, never which statements ran.

The container joins the host network because `.env.testing` points the
database fixtures at `127.0.0.1`; start mysql and postgres the way the pipeline
does (`docker compose up -d --wait mysql postgres`) before a run.

The binary is built for the host with `CGO_ENABLED=0` and copied into the image,
so the numbers belong to the tree in this checkout and not to a published
image. The mounted tree supplies the fixtures.

On a shared host set `TESTCOV_CPUSET=1` to pin the container to one core. The
suite is start-up bound and a neighbour's load moves the wall time by more than
the recording mode does; pinning keeps the sample spread near five percent.

## Results

Measured on this machine, 546 fixtures, best of five, one container:

| Mode | Best (ms) | Runtime engine (ms) | Coverage |
|------|----------:|--------------------:|----------|
| uncovered | 1573 | 1178 | no profile |
| atomic | 1620 | 1208 | 141 files, 93.9% (2045/2178 statements), 7456 profile lines, 428,763 bytes |
| count | 1615 | 1211 | same statements, counts beside the marks |

Coverage costs 3.0% of the wall time and 2.5% of the runtime engine; atomic and
count are within noise of each other because the suite executes few statements
per fixture. The two modes mark the same statements.

## What this benchmark found

Before this tree, the `uncovered` arm was a coverage run. `flags.Check`
normalized `--covermode` to its default even when the operator named nothing,
and `phpx test` read a non-empty mode as `--covermode` having been given, so
every `phpx test` installed a collector and wrote `phpscript.cov` into the
working directory. That is why this benchmark's first results read uncovered
1.54 s and atomic 1.52 s: both runs collected. `Check` now leaves the mode as
the operator wrote it, the test command sees an empty one, and an uncovered run
allocates no collector and writes no profile. The regression is pinned by
`TestOptions_Check` in `internal/flags`.

The same run surfaced a second accounting bug: an absolute path argument left
the `.phpt` entrypoint counted as a covered source file, because the collector
drops the leading separator and the comparison against the argument did not.
A relative run reported `-` files for a folder of entrypoints while an absolute
run reported every one of them at 100%. `fixtureCoverName` normalizes both
sides, and `TestRunCommandCoverAbsolutePath` covers the two spellings.
