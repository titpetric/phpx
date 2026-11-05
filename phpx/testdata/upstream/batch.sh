#!/usr/bin/env bash
# Runs run.sh for each library of the manifest and appends the results to a log.
# usage: PHPX=<binary> batch.sh <log> [name ...]
cd "$(dirname "$0")" || exit 1
log=$1; shift
want=" $* "
while read -r name url tests extras; do
	case "$name" in ''|'#'*) continue;; esac
	if [ -n "$*" ] && [[ "$want" != *" $name "* ]]; then continue; fi
	./run.sh "$name" "$url" "$tests" $extras >> "$log" 2>&1
done < libraries.txt
echo "done" >> "$log"
