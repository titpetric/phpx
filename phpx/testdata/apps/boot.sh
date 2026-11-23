#!/usr/bin/env bash
# Tests 4 and 5: boot the app under php and phpx (boot.php), then measure the
# statements of the app's own code that phpx executed while it ran.
# Usage: boot.sh <phpx> <app-dir>      prints one markdown row
set -u
PHPX=$(realpath "$1"); APP=$(realpath "$2"); here=$(cd "$(dirname "$0")" && pwd)
name=$(basename "$APP"); OUT=${OUT:-/tmp/boot-$name}; mkdir -p "$OUT"
cp "$here/boot.php" "$APP/phpx_boot.php"; cd "$APP" || exit 1
php phpx_boot.php >"$OUT/php.txt" 2>&1
timeout 300 "$PHPX" --cover --coverfile="$OUT/boot.cov" run phpx_boot.php >"$OUT/phpx.txt" 2>&1
rm -f phpx_boot.php
stage() {
	local n; n=$(grep -E '^[0-9] ' "$1" | tail -1 | cut -d' ' -f1)
	case "$n" in
	9) echo "9 ($(sed -n 's/^9 response: //p' "$1" | head -1))" ;;
	"") echo "n/a" ;;
	*) echo "$n" ;;
	esac
}
reason() { grep -m1 -E 'Unexpected error|Fatal error|^9a exception' "$1" | sed -E 's/^9a exception: //; s/^[0-9]{4}\/[0-9]{2}\/[0-9]{2} [0-9:]{8} //; s/^PHP Fatal error: +//' | cut -c1-110; }
classes() { local c; c=$(grep -m1 '^classes: ' "$1" | cut -d' ' -f2); echo "${c:-n/a}"; }
read -r ahit atot fhit ftot vhit vtot < <(python3 - "$OUT/boot.cov" <<'PY'
import re,sys
seen={}
def tally(prefix):
    return (sum(s for k,(s,c) in seen.items() if k.startswith(prefix)),
            sum(s for k,(s,c) in seen.items() if k.startswith(prefix) and c>0))
for raw in open(sys.argv[1]):
    m=re.match(r"^(.*):\d+\.\d+,\d+\.\d+ (\d+) (\d+)$",raw.strip())
    if not m:
        continue
    k=raw.split(" ")[0]
    seen[k]=(int(m.group(2)),max(int(m.group(3)),seen.get(k,(0,0))[1]))
out=[]
for prefix in ("app/","vendor/laravel/framework/src/","vendor/"):
    total,hit=tally(prefix)
    out += [hit,total]
print(*out)
PY
)
echo "| \`$name\` | $(stage "$OUT/php.txt") | $(stage "$OUT/phpx.txt") | $(reason "$OUT/phpx.txt") | ${ahit:-0} / ${atot:-0} | ${fhit:-0} / ${ftot:-0} | ${vhit:-0} / ${vtot:-0} | $(classes "$OUT/php.txt") / $(classes "$OUT/phpx.txt") |"
