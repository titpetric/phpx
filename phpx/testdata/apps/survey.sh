#!/usr/bin/env bash
# Test 1 and 2 of the consumer set: lints an app's own code and its vendor tree,
# one markdown row per tree. Usage: survey.sh <phpx> <apps-dir> [app...]
set -u
PHPX=$(realpath "$1"); APPS=$2; shift 2
apps=${*:-$(ls "$APPS")}
tmp=$(mktemp -d)
echo "| App | Tree | Files | Parses | Warning findings | Failing | Parse rate |"
echo "|-----|------|------:|-------:|-----------------:|--------:|-----------:|"
for app in $apps; do
	cd "$APPS/$app" || continue
	for tree in app routes database config tests resources vendor/laravel/framework/src vendor/filament vendor/livewire vendor/symfony; do
		[ -d $tree ] || continue
		f="$tmp/$app-${tree////_}.txt"
		timeout 900 "$PHPX" lint --include vendor/autoload.php "./$tree/..." >"$f" 2>&1
		# A file parses when its rows are PASS or WARN; one FAIL row is a parse error.
		# The lint summary's "Passing N, with W warnings" counts files and findings
		# in the same line, so the rows are read instead: a warning names a file once
		# per finding, and a file with several warnings is still one file.
		read -r n p w x < <(awk -F'|' '
			/^## / { sec=$0; sub(/^## /, "", sec); gsub(/ +$/, "", sec); next }
			/^\| *(PASS|WARN|FAIL) *\|/ {
				st=$2; gsub(/^ +| +$/, "", st)
				name=$3; gsub(/^ +| +$/, "", name)
				key=sec "/" name
				if (st == "FAIL") { if (!(key in fail)) { fail[key]=1; x++ } }
				else { if (!(key in ok)) { ok[key]=1; p++ } }
				if (st == "WARN") w++
			}
			END { print p+x, p, w, x }' "$f")
		n=${n:-0}; p=${p:-0}; w=${w:-0}; x=${x:-0}
		[ "$n" -gt 0 ] || continue
		echo "| \`$app\` | \`$tree/\` | $n | $p | $w | $x | $(awk "BEGIN{printf \"%.1f%%\", $p*100/$n}") |"
	done
done
echo
echo "Parse errors, by cause:"
echo
cat "$tmp"/*.txt | grep '^| FAIL' | sed 's/.*| *1 | *//; s/line [0-9]*: //; s/@[0-9]*//; s/  *|$//; s/^\(.\{110\}\).*/\1.../' | sort | uniq -c | sort -rn | head -15 | sed 's/^ *\([0-9]*\) \(.*\)/- \1 files: \2/'
cp -r "$tmp" "${KEEP:-/tmp/app-survey}" 2>/dev/null
