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
stage() { grep -E '^[0-9] ' "$1" | tail -1 | cut -c1-40; }
reason() { grep -m1 -E 'Unexpected error|Fatal error' "$1" | sed -E 's/^[0-9\/: ]+//; s/PHP Fatal error: +//' | cut -c1-110; }
read -r total hit < <(python3 - "$OUT/boot.cov" "$APP/app/" <<'PY'
import re,sys
seen={}
for raw in open(sys.argv[1]):
    m=re.match(r"^(.*):\d+\.\d+,\d+\.\d+ (\d+) (\d+)$",raw.strip())
    if m and m.group(1).startswith(sys.argv[2]) or (m and m.group(1).startswith("app/")):
        k=raw.split(" ")[0]; seen[k]=(int(m.group(2)),max(int(m.group(3)),seen.get(k,(0,0))[1]))
print(sum(s for s,c in seen.values()), sum(s for s,c in seen.values() if c>0))
PY
)
echo "| \`$name\` | $(stage "$OUT/php.txt") | $(stage "$OUT/phpx.txt") | $(reason "$OUT/phpx.txt") | ${hit:-0} / ${total:-0} |"
