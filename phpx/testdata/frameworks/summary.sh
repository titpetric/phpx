#!/usr/bin/env bash
# Runs each framework's phpunit shim under php and under phpx and prints the
# tests that pass in each and the first wall phpx meets.
# Usage: summary.sh [phpx binary]    after `composer install` in each folder
cd "$(dirname "$0")" || exit 1
PHPX=${1:-phpx}
echo "| Framework | Tests | php passes | phpx passes | first phpx wall |"
echo "|---|--:|--:|--:|---|"
for d in slim latte laravel filament twig commonmark yaml carbon phpparser mustache uuid; do
	[ -d "$d/vendor" ] || { echo "| $d | | | | run composer install in $d |"; continue; }
	php_out=$(cd "$d" && php launcher.php --testdox tests/ 2>&1)
	x_out=$(cd "$d" && timeout 300 "$PHPX" run launcher.php --testdox tests/ </dev/null 2>&1)
	total=$(grep -oE 'OK \([0-9]+ test|Tests: [0-9]+' <<<"$php_out" | grep -oE '[0-9]+' | head -1)
	p=$(grep -c ' ✔ ' <<<"$php_out")
	x=$(grep -c ' ✔ ' <<<"$x_out")
	wall=$(grep -m1 -E 'Unexpected error|errorString|Error:|Exception:|line [0-9]+:' <<<"$x_out" | sed -E 's/^[ │]*//; s/^[0-9/: ]*Unexpected error: //' | cut -c1-110 | tr '|' '/')
	echo "| \`$d\` | $total | $p | $x | $wall |"
done
