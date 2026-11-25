# testcov

The test suite under statement coverage, in a container. It answers one
question: what does `phpx test` cost with coverage, and what does the recording
mode change?

`./run.sh [--covermode "atomic count"]` builds `bin/phpx`, stages the binary
and the fixture trees into a small image, and runs the workload once to warm
the page cache and five times per mode, printing every time and keeping the
best:

- **uncovered**: `phpx test --skip-php --matrix` over the coverage pipeline's
  fixture set: `./tests/...` and every area under `phpx/testdata` except
  `patches/`, which is the overlay and not a fixture tree.
- **atomic** and **count**: the same run with `--cover --covermode=<mode>`.

The container joins the host network because `.env.testing` points the
database fixtures at `127.0.0.1`; start mysql and postgres the way the pipeline
does (`docker compose up -d --wait mysql postgres`) before a run.

The binary is built for the host with `CGO_ENABLED=0` and copied into the image,
so the numbers belong to the tree in this checkout and not to a published
image. The mounted tree supplies the fixtures.

On a shared host set `TESTCOV_CPUSET=1` to pin the container to one core. The
suite is start-up bound and a neighbour's load moves the wall time by more than
the recording mode does; pinning keeps the sample spread near five percent.
