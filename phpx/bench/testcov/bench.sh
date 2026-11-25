#!/bin/bash
#
# testcov runs the fixture suite the coverage pipeline runs, once without
# coverage and once per --covermode, and prints a markdown table: the five
# wall times, the best of them and the run's own summary.
#
# The workload is cover:phpx-test's: ./tests/... plus every fixture area under
# phpx/testdata except patches, which is an overlay and not a fixture tree.
# --skip-php leaves out php's column; the built-in engines are what this
# benchmark measures.
set -u
cd /app || exit 1
set -a
# .env.testing is a key=value file, not a shell script: the mysql URL holds
# parentheses. Export each line whole, which quotes the value.
if [ -f .env.testing ]; then
	while IFS= read -r line; do
		case "$line" in '' | '#'*) continue ;; esac
		export "$line"
	done <.env.testing
fi
set +a

profiles=/tmp/testcov
mkdir -p "$profiles"

paths="./tests/..."
for dir in phpx/testdata/*/; do
	name=${dir%/}
	[ "${name##*/}" = patches ] && continue
	paths="$paths ./$name/..."
done

bench() {
	local label=$1
	shift
	local best=999999
	local s e ms out bestout="" times=""
	# One unmeasured run: the first pass reads the fixture tree into the
	# page cache, and charging that to whichever mode runs first reads as
	# a mode difference.
	# shellcheck disable=SC2086
	phpx test --skip-php --matrix "$@" $paths >/dev/null 2>&1
	for _ in 1 2 3 4 5; do
		s=$(date +%s%N)
		# shellcheck disable=SC2086
		out=$(phpx test --skip-php --matrix "$@" $paths 2>&1)
		e=$(date +%s%N)
		ms=$(((e - s) / 1000000))
		times="$times $ms"
		if [ "$ms" -lt "$best" ]; then
			best=$ms
			bestout=$out
		fi
	done
	summary=$(printf '%s\n' "$bestout" | grep 'Matrix summary' | tail -1)
	printf '| %s | %d |%s | %s |\n' "$label" "$best" "$times" "$summary"
}

echo "| Mode | Best (ms) | Runs (ms) | Result |"
echo "|---|---:|---|---|"
bench uncovered
for mode in ${COVERMODES:-atomic count}; do
	bench "$mode" --cover --coverfile "$profiles/suite-$mode.cov" --covermode="$mode"
	echo "  profile: $(head -1 "$profiles/suite-$mode.cov"), $(wc -l <"$profiles/suite-$mode.cov") lines"
done
