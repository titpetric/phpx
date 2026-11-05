#!/usr/bin/env bash
# Lists the tests that fail under phpx and pass under php.
# usage: delta.sh <name> <test path>     after run.sh has set .cache/<name> up
set -u
cd "$(dirname "$0")/.cache/$1" || exit 1
PHPX=${PHPX:-phpx}
args=(--no-configuration --bootstrap vendor/autoload.php --do-not-cache-result "$2")
names() { grep -E '^[0-9]+\) ' | sed -E 's/^[0-9]+\) //; s/ with data set.*//' | sort -u; }
php launcher.php "${args[@]}" 2>&1 </dev/null | names > /tmp/delta-php.txt
here=$(pwd)
timeout "${TIMEOUT:-300}" "$PHPX" -w / run "$here/launcher.php" --no-configuration --bootstrap "$here/vendor/autoload.php" --do-not-cache-result "$here/$2" 2>&1 </dev/null > /tmp/delta-phpx.out
names < /tmp/delta-phpx.out > /tmp/delta-phpx.txt
comm -13 /tmp/delta-php.txt /tmp/delta-phpx.txt
