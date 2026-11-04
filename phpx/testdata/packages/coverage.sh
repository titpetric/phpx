#!/usr/bin/env bash
# Statement coverage of the nine libraries under phpx.
# usage: coverage.sh [phpx binary] [profile]    from anywhere; prints the table
# Add `--files <library>` after the profile to list that library's least covered files.
set -u
cd "$(dirname "$0")" || exit 1
PHPX=${1:-phpx}
PROFILE=${2:-${TMPDIR:-/tmp}/phpx-libraries.cov}
shift 2 2>/dev/null
"$PHPX" --cover --coverfile="$PROFILE" run coverage_launcher.php tests/ </dev/null 2>&1 | tail -4
python3 ../../scripts/lib-coverage.py "$PROFILE" "$@"
