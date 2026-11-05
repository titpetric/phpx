#!/usr/bin/env bash
# The five consumer tests over a set of real apps. Usage: run.sh <phpx> <apps-dir> <apps-prod-dir> [app...]
#   apps-dir       checkouts with `composer install` (dev dependencies): the parse survey
#   apps-prod-dir  the same checkouts with `composer install --no-dev`: load, boot, request, measure
# See README.md for the clone and install commands.
set -u
here=$(cd "$(dirname "$0")" && pwd); PHPX=$1; DEV=$(realpath "$2"); PROD=$(realpath "$3"); shift 3
apps=${*:-$(ls "$PROD")}
echo "## 1. Parse"; echo; "$here/survey.sh" "$PHPX" "$DEV" $apps; echo
echo "## 2. Class load"; echo
echo "| App | Classes | php loads | phpx loads | both load | phpx / php |"; echo "|---|--:|--:|--:|--:|--:|"
for a in $apps; do (cd "$PROD" && "$here/load.sh" "$PHPX" "$a"); done; echo
echo "## 3-5. Boot, first request, measure"; echo
echo "| App | php reached | phpx reached | phpx stopped at | app statements run |"; echo "|---|---|---|---|--:|"
for a in $apps; do (cd "$PROD" && "$here/boot.sh" "$PHPX" "$a"); done
