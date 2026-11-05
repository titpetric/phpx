#!/usr/bin/env bash
# Test 1 and 2 of the consumer set: lints an app's own code and its vendor tree,
# one markdown row per tree. Usage: survey.sh <phpx> <apps-dir> [app...]
set -u
PHPX=$(realpath "$1"); APPS=$2; shift 2
apps=${*:-$(ls "$APPS")}
tmp=$(mktemp -d)
echo "| App | Tree | Files | Passing | Warnings | Failing | Parse rate |"
echo "|-----|------|------:|--------:|---------:|--------:|-----------:|"
for app in $apps; do
	cd "$APPS/$app" || continue
	for tree in app routes database config tests resources vendor/laravel/framework/src vendor/filament vendor/livewire vendor/symfony; do
		[ -d $tree ] || continue
		f="$tmp/$app-${tree////_}.txt"
		timeout 900 "$PHPX" lint --include vendor/autoload.php "./$tree/..." >"$f" 2>&1
		read -r p w x < <(grep -E '^Passing [0-9]+, with' "$f" | tail -1 | sed -E 's/[^0-9]+/ /g')
		p=${p:-0}; w=${w:-0}; x=${x:-0}
		n=$((p+w+x)); [ $n -gt 0 ] || continue
		echo "| \`$app\` | \`$tree/\` | $n | $p | $w | $x | $(awk "BEGIN{printf \"%.1f%%\", ($p+$w)*100/$n}") |"
	done
done
echo
echo "Parse errors, by cause:"
echo
cat "$tmp"/*.txt | grep '^| FAIL' | sed 's/.*| *1 | *//; s/line [0-9]*: //; s/@[0-9]*//; s/  *|$//; s/^\(.\{110\}\).*/\1.../' | sort | uniq -c | sort -rn | head -15 | sed 's/^ *\([0-9]*\) \(.*\)/- \1 files: \2/'
cp -r "$tmp" "${KEEP:-/tmp/app-survey}" 2>/dev/null
