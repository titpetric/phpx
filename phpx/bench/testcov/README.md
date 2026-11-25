# testcov

The test suite under statement coverage, in a container. It answers one
question: what does `phpx test` cost with coverage, and what does the recording
mode change?

`./run.sh [--covermode "atomic count"]` builds `bin/phpx`, stages the binary
and the fixture trees into a small image, and runs the workload five times per
mode, keeping the best wall time:

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
