#!/usr/bin/env bash
# Test 3, the class-load test. Usage: load.sh <phpx> <app-dir>
# One process per class under php and under phpx. Prints a markdown row and
# leaves the per-class results in $OUT (default /tmp/load-<app>).
set -u
PHPX=$(realpath "$1"); APP=$(realpath "$2"); here=$(cd "$(dirname "$0")" && pwd)
name=$(basename "$APP"); OUT=${OUT:-/tmp/load-$name}; mkdir -p "$OUT"
cp "$here/load.php" "$APP/phpx_load.php"; cd "$APP" || exit 1
php phpx_load.php > "$OUT/classes.txt" 2>/dev/null
one() { # runtime script class
	r=$($1 phpx_load.php "$2" 2>&1 | grep -m1 -E '^(OK|ERR)|rror|undefined|Undefined' | cut -c1-200)
	echo "${r:-ERR no output}"
}
run() { # runtime label
	while read -r c; do printf '%s\t%s\n' "$c" "$(timeout 60 bash -c "$(declare -f one); one '$1' '$c'")"; done < "$OUT/classes.txt" > "$OUT/$2.tsv"
}
run php php & run "$PHPX run" phpx & wait
total=$(wc -l < "$OUT/classes.txt")
pok=$(grep -c $'\tOK$' "$OUT/php.tsv"); xok=$(grep -c $'\tOK$' "$OUT/phpx.tsv")
both=$(join -t $'\t' <(sort "$OUT/php.tsv") <(sort "$OUT/phpx.tsv") | awk -F'\t' '$2=="OK"&&$3=="OK"' | wc -l)
echo "| \`$name\` | $total | $pok | $xok | $both | $(awk "BEGIN{printf \"%.1f%%\", $both*100/($pok>0?$pok:1)}") |"
rm -f phpx_load.php
