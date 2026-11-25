#!/usr/bin/env bash
# Statement coverage of the nine libraries under phpx.
# usage: coverage.sh [phpx binary] [profile] [--covermode count|atomic] [--files <library>]
# Prints the table; --covermode goes to phpx like the other coverage flags, and
# --files lists that library's least covered files.
set -u
cd "$(dirname "$0")" || exit 1
PHPX=${1:-phpx}
PROFILE=${2:-${TMPDIR:-/tmp}/phpx-libraries.cov}
shift 2 2>/dev/null
mode=""
rest=()
while [ $# -gt 0 ]; do
	case "$1" in
	--covermode)
		mode=${2:-}
		shift 2
		;;
	--covermode=*)
		mode=${1#*=}
		shift
		;;
	*)
		rest+=("$1")
		shift
		;;
	esac
done
"$PHPX" --cover --coverfile="$PROFILE" ${mode:+--covermode="$mode"} run coverage_launcher.php tests/ </dev/null 2>&1 | tail -4
python3 ../../scripts/lib-coverage.py "$PROFILE" "${rest[@]}"
