#!/bin/sh
# Usage: run.sh [phpx binary]. Times flat.php and oop.php on php and phpx, best of 3.
bin=${1:-phpx}; cd "$(dirname "$0")"
best() { b=999999; for i in 1 2 3; do s=$(date +%s%N); "$@" >/dev/null; e=$(( ($(date +%s%N)-s)/1000000 )); [ $e -lt $b ] && b=$e; done; echo $b; }
for f in flat oop; do
  echo "$f: php $(php $f.php) | phpx $($bin $f.php)"
  echo "  php ${f}: $(best php $f.php) ms, phpx: $(best $bin $f.php) ms"
done
