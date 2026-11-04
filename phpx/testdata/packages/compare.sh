#!/usr/bin/env bash
# Runs each tests/*Test.php under php and under phpx and diffs what phpunit
# prints, with the lines that differ by machine (time, memory, paths) removed.
# Usage: compare.sh [phpx binary]    from phpx/testdata/phpunit, after composer install
set -u
cd "$(dirname "$0")"
PHPX=${1:-phpx}
norm() { grep -v -E '^Time: |^Runtime: |^PHPUnit |^/|\.php:[0-9]+$' | sed -E 's#/[^ ]*/tests/#tests/#g' | sed -E '/^ *│ *$/d' | sed -E '/^$/d'; }
status=0
for t in tests/*Test.php; do
	name=$(basename "$t" .php)
	php launcher.php --testdox "$t" 2>&1 | norm > "/tmp/phpx-cmp-$name.php.txt"
	"$PHPX" run launcher.php --testdox "$t" 2>&1 | norm > "/tmp/phpx-cmp-$name.phpx.txt"
	if diff -q "/tmp/phpx-cmp-$name.php.txt" "/tmp/phpx-cmp-$name.phpx.txt" > /dev/null; then
		echo "same  $name: $(tail -1 "/tmp/phpx-cmp-$name.phpx.txt")"
	else
		echo "DIFF  $name"; diff "/tmp/phpx-cmp-$name.php.txt" "/tmp/phpx-cmp-$name.phpx.txt" | head -20; status=1
	fi
done
exit $status
