#!/bin/bash
#
# testcov runs the fixture suite the coverage pipeline runs, once without
# coverage and once per --covermode, and prints a markdown table: the five
# wall times, the best of them, the run's own summary, and what the generated
# profile says. The atomic and count profiles are compared line for line: a
# mode may change the cost and the counts, never which statements ran.
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

# stats reads a profile and prints what it measured of the source the
# fixtures load, leaving out the .phpt entrypoints, which run by definition.
# A profile line is "file:start.col,end.col statements count".
stats() {
	awk '
		NR == 1 { next }
		{
			split($1, at, ":")
			if (at[1] ~ /\.phpt$/) next
			files[at[1]] = 1
			total += $2
			if ($3 > 0) covered += $2
		}
		END {
			pct = total ? 100 * covered / total : 0
			printf "%d files, %.1f%% (%d/%d stmts)", length(files), pct, covered, total
		}
	' "$1"
}

# marks prints each covered statement range, for a line-for-line comparison.
marks() {
	awk 'NR > 1 && $3 > 0 { print $1 }' "$1" | sort
}

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
done

echo
echo "| Mode | Coverage | Profile |"
echo "|---|---|---|"
for mode in ${COVERMODES:-atomic count}; do
	file="$profiles/suite-$mode.cov"
	if [ -f "$file" ]; then
		printf '| %s | %s | %d lines, %d bytes |\n' "$mode" "$(stats "$file")" "$(wc -l <"$file")" "$(wc -c <"$file")"
	else
		printf '| %s | no profile | |\n' "$mode"
	fi
done

# The two modes must agree on which statements ran. count differs only in the
# number each line carries, which a coverage reading ignores.
if [ -f "$profiles/suite-atomic.cov" ] && [ -f "$profiles/suite-count.cov" ]; then
	if diff -q <(marks "$profiles/suite-atomic.cov") <(marks "$profiles/suite-count.cov") >/dev/null; then
		echo
		echo "atomic and count mark the same statements."
	else
		echo
		echo "atomic and count disagree on the statements marked:"
		diff <(marks "$profiles/suite-atomic.cov") <(marks "$profiles/suite-count.cov") | head
	fi
fi
