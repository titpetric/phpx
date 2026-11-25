#!/usr/bin/env bash
#
# run.sh builds the phpx binary and runs the testcov compose benchmark.
# Usage: phpx/bench/testcov/run.sh [--covermode modes]
# The modes reach the container as COVERMODES.
#
# The build context is staged in a temporary directory: the binary, the
# benchmark script, .env.testing and the fixture trees (tests/ and
# phpx/testdata/ without patches, frameworks and upstream, which are an overlay,
# vendored libraries and a cache rather than fixture areas). Staging keeps the
# image self-contained and the 1.6G checkout out of the docker build.
set -euo pipefail
here="$(cd "$(dirname "$0")" && pwd)"
root="$(cd "$here/../../.." && pwd)"
cd "$root"

modes="atomic count"
if [ "${1:-}" = "--covermode" ]; then
	modes="${2:-}"
fi

ctx="$(mktemp -d)"
trap 'rm -rf "$ctx"' EXIT
mkdir -p "$ctx/bin" "$ctx/phpx"

CGO_ENABLED=0 go build -tags phpx -o "$ctx/bin/phpx" .
cp "$here/bench.sh" "$ctx/bench.sh"
cp .env.testing "$ctx/.env.testing"
tar -C "$root" -cf - tests | tar -C "$ctx" -xf -
tar -C "$root/phpx" -cf - --exclude=frameworks --exclude=upstream --exclude=patches testdata |
	tar -C "$ctx/phpx" -xf -

TESTCOV_CONTEXT="$ctx" TESTCOV_DOCKERFILE="$here/Dockerfile" COVERMODES="$modes" \
	docker compose -f "$here/compose.yml" run --rm --build testcov
