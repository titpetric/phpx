#!/usr/bin/env bash
#
# run.sh runs the commonmark suite on php and on phpx and compares them one
# test at a time (compare.py). It is the fastcov benchmark: turbocov is the
# php driver, phpx cannot load an extension and records statement coverage in
# its own interpreter, so the suite itself is what both stacks run.
#
#   COMMONMARK=/path/to/commonmark ./run.sh [unit|functional|both] [php]
#
# The first argument picks the suite, the second names the php binary. Both
# stacks write a JUnit report; compare.py reads them and writes
# phpx/bench/fastcov/report.md. The exit status is 1 when a test passes on php
# and not on phpx, so this can gate a pipeline.
set -euo pipefail

here=$(cd "$(dirname "$0")" && pwd)
repo=$(cd "$here/../../.." && pwd)
checkout=${COMMONMARK:-}
suite=${1:-unit}
php=${2:-php}
out=${OUT:-$here/report.md}

if [ -z "$checkout" ] || [ ! -f "$checkout/vendor/bin/phpunit" ]; then
	echo "run.sh: set COMMONMARK to a commonmark checkout with its vendor tree" >&2
	exit 2
fi

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

# phpx answers PHP_BINARY with the shim a separate-process test starts, which
# speaks the -d and -c flags phpunit writes.
cat >"$tmp/phpx-shim.sh" <<'SHIM'
#!/bin/sh
while [ $# -gt 0 ]; do
	case "$1" in
	-d|-c) shift 2 ;;
	-d*|-c*) shift ;;
	-n|-q|-E) shift ;;
	-f) shift; break ;;
	--) shift; break ;;
	-*) shift ;;
	*) break ;;
	esac
done
exec "$PHPX_BINARY" run "$@"
SHIM
chmod +x "$tmp/phpx-shim.sh"
(cd "$repo" && go build -tags phpx -o "$tmp/phpx" .)
export PHPX_BINARY="$tmp/phpx"

case "$suite" in
unit) testsuite=(--testsuite 'league/commonmark Unit Tests') ;;
functional) testsuite=(--testsuite 'league/commonmark Functional Tests') ;;
both) testsuite=() ;;
*)
	echo "run.sh: suite is unit, functional or both" >&2
	exit 2
	;;
esac

cd "$checkout"
"$php" vendor/bin/phpunit -c phpunit-plain.xml "${testsuite[@]}" \
	--log-junit "$tmp/php.xml" --do-not-cache-result >/dev/null 2>&1 || true
PHPX_PHP_BINARY="$tmp/phpx-shim.sh" "$tmp/phpx" run vendor/bin/phpunit -c phpunit-plain.xml \
	"${testsuite[@]}" --log-junit "$tmp/phpx.xml" --do-not-cache-result >/dev/null 2>&1 || true

python3 -I "$here/compare.py" --php "$tmp/php.xml" --phpx "$tmp/phpx.xml" --out "$out"
status=$?
echo "report: $out"
exit $status
