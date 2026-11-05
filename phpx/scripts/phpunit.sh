#!/usr/bin/env bash
# Runs the test files a phpunit.yml globs through phpunit's source tree, under
# the phpx binary. Usage: phpunit.sh [phpx binary] [phpunit.yml] [phpunit args]
# Run from the folder that holds launcher.php and vendor/.
set -eu
PHPX=${1:-phpx}
CFG=${2:-phpunit.yml}
shift $(( $# > 2 ? 2 : $# ))
status=0
n=0
while read -r glob; do
	for f in $glob; do
		[ -f "$f" ] || continue
		n=$((n + 1))
		"$PHPX" run launcher.php "$@" "$f" || status=1
	done
done < <(sed -n -E 's/^[[:space:]]*-[[:space:]]+(.+)$/\1/p' "$CFG")
[ "$n" -gt 0 ] || { echo "no test file matches $CFG" >&2; exit 2; }
exit $status
