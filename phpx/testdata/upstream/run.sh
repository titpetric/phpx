#!/usr/bin/env bash
# Runs a library's own phpunit suite under php and under phpx and compares the totals.
# usage: run.sh <name> <git url> <test path> [extra dev package ...]   (phpx binary: $PHPX, default phpx)
# The checkout, its vendor tree and the results live in .cache/<name>, which git ignores.
set -u
cd "$(dirname "$0")" || exit 1
name=$1; url=$2; tests=$3; shift 3; extras=("$@"); major=^11.5
PHPX=${PHPX:-phpx}
dir=.cache/$name
mkdir -p .cache
if [ ! -d "$dir/.git" ]; then
	git clone -q --depth 1 "$url" "$dir" 2>&1 | tail -1
fi
cd "$dir" || exit 1
if [ ! -f .phpx-ready ]; then
	composer install --ignore-platform-reqs -q --no-interaction 2>&1 | tail -2
	if [ ! -d vendor/phpunit/phpunit ] || [ "${#extras[@]}" -gt 0 ]; then
		composer require --dev "phpunit/phpunit:$major" "${extras[@]}" --ignore-platform-reqs -q --no-interaction -W 2>&1 | tail -2
	fi
	mkdir -p shims
	cp ../../launcher.php launcher.php
	cp ../../../frameworks/latte/shims/RecursionContext.php shims/
	touch .phpx-ready
fi
args=(--no-configuration --bootstrap vendor/autoload.php --do-not-cache-result "$tests")
pout=$(timeout ${TIMEOUT:-300} php launcher.php "${args[@]}" 2>&1 </dev/null | tail -40)
here=$(pwd)
# The runtime roots the file functions at its working directory; the root is / here, as php has it.
xargs=(--no-configuration --bootstrap "$here/vendor/autoload.php" --do-not-cache-result "$here/$tests")
xout=$(timeout ${TIMEOUT:-300} "$PHPX" -w / run "$here/launcher.php" "${xargs[@]}" 2>&1 </dev/null | tail -40)
summary() { grep -E '^(OK|Tests:|FAILURES|ERRORS|No tests)' <<<"$1" | tr '\n' ' '; }
echo "$name"
echo "  php : $(summary "$pout")"
echo "  phpx: $(summary "$xout")"
if [ "$(summary "$pout")" != "$(summary "$xout")" ]; then
	echo "  first phpx lines:"; grep -v '^$' <<<"$xout" | head -8 | sed 's/^/    /' | cut -c1-200
fi
