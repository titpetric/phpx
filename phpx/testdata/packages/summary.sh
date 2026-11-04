#!/usr/bin/env bash
# Runs every tests/*Test.php under php and under phpx and prints, per package, the
# tests that pass in each (testdox check marks) and whether phpunit ran to the end.
# Usage: summary.sh [phpx binary]    from phpx/testdata/packages, after composer install
cd "$(dirname "$0")" || exit 1
PHPX=${1:-phpx}
echo "| Package | Test file | php passes | phpx passes | phpx ran to the end | first phpx wall |"
echo "|---|---|--:|--:|:-:|---|"
tp=0; tx=0
for t in tests/*Test.php; do
	n=$(basename "$t" .php)
	php launcher.php --testdox "$t" > /tmp/phpx-sum-$n.php 2>&1
	"$PHPX" run launcher.php --testdox "$t" > /tmp/phpx-sum-$n.phpx 2>&1
	total=$(grep -oE 'Tests: [0-9]+|OK \([0-9]+ test' /tmp/phpx-sum-$n.php | grep -oE '[0-9]+' | head -1)
	p=$(grep -c ' ✔ ' /tmp/phpx-sum-$n.php)
	x=$(grep -c ' ✔ ' /tmp/phpx-sum-$n.phpx)
	end=no; grep -qE '^(OK|FAILURES|ERRORS|Tests:)' /tmp/phpx-sum-$n.phpx && end=yes
	wall=$(grep -m1 -E 'errorString|Error|error' /tmp/phpx-sum-$n.phpx | sed -E 's/^[ │]*//' | cut -c1-90 | tr '|' '/')
	echo "| ${n%Test} | $total tests | $p | $x | $end | $wall |"
	tp=$((tp+p)); tx=$((tx+x))
done
echo "| **total** | | **$tp** | **$tx** | | |"
