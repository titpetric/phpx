#!/usr/bin/env bash
# Lints the source of each package under phpx/testdata/packages/vendor and
# prints one markdown row per package plus the tally of what blocked it.
# Usage: PHPX=bin/phpx phpx/scripts/package-survey.sh > phpx/testdata/packages/survey.md
set -u
cd "$(dirname "$0")/../testdata/packages"
PHPX=${PHPX:-phpx}
[ -d vendor ] || composer install --no-interaction >&2
out=$(mktemp -d)
echo "| Package | Source | Passing | Warnings | Failing |"
echo "|---------|--------|--------:|---------:|--------:|"
while read -r pkg dir; do
	f="$out/${pkg//\//_}.txt"
	"$PHPX" lint --include vendor/autoload.php "./vendor/$pkg/$dir/..." >"$f" 2>&1
	pass=$(grep -c '^| PASS' "$f"); warn=$(grep -c '^| WARN' "$f"); fail=$(grep -c '^| FAIL' "$f")
	echo "| \`$pkg\` | \`$dir\` | $pass | $warn | $fail |"
done <<'L'
psr/log src
psr/container src
psr/http-message src
guzzlehttp/psr7 src
guzzlehttp/guzzle src
symfony/console .
symfony/finder .
monolog/monolog src
doctrine/inflector src
L
echo
echo "Parse errors, by cause:"
echo
cat "$out"/*.txt | grep '^| FAIL' | sed 's/.*| *1 | *//; s/line [0-9]*: //; s/@[0-9]*//; s/  *|$//; s/^\(.\{110\}\).*/\1.../' | sort | uniq -c | sort -rn | head -12 | sed 's/^ *\([0-9]*\) \(.*\)/- \1 files: \2/'
echo
echo "Undefined functions (the binding gap), by calls:"
echo
cat "$out"/*.txt | grep -o 'call to undefined function [A-Za-z_\\0-9]*()' | sed 's/.*\\//; s/call to undefined function //' | sort | uniq -c | sort -rn | sed 's/^ *\([0-9]*\) \(.*\)/- `\2` x\1/'
rm -rf "$out"
