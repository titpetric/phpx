#!/bin/sh
# run.sh measures the commonmark suite turbocov benchmarks, four ways: php
# without coverage, php with turbocov, php with xdebug's coverage mode, and
# phpx. Every leg runs the same checkout in the image the Dockerfile builds.
#
#   ./run.sh [repeats]      repeats is 3 by default, and a leg reports its best.
#
# A checkout of turbocov and the phpx binary are staged into a build context,
# because the docker daemon here cannot see the host's paths. The image carries
# commonmark 2.10.7 and phpunit with turbocov's driver installed.
set -eu

here=$(cd "$(dirname "$0")" && pwd)
repo=$(cd "$here/../../.." && pwd)
repeats=${1:-3}
image=${IMAGE:-phpx-turbocov}
ctx=$(mktemp -d)
trap 'rm -rf "$ctx"' EXIT

git clone --quiet --depth 1 https://github.com/shipmonk-rnd/turbocov "$ctx/turbocov"
cp "$here/Dockerfile" "$here"/phpunit-*.xml "$here/phpx-shim.sh" "$ctx/"
(cd "$repo" && go build -tags phpx -o "$ctx/phpx" .)

docker build --quiet -t "$image" "$ctx"

docker run --rm -e REPEATS="$repeats" "$image" sh -c '
cd /app
run() {
	count=$1
	name=$2
	shift 2
	best=
	for _ in $(seq "$count"); do
		s=$(date +%s%N)
		if ! "$@" >/tmp/"$name".log 2>&1; then
			echo "$name failed:" >&2
			tail -5 /tmp/"$name".log >&2
		fi
		e=$(date +%s%N)
		ms=$(( (e - s) / 1000000 ))
		if [ -z "$best" ] || [ "$ms" -lt "$best" ]; then best=$ms; fi
	done
	printf "%s\t%s\n" "$name" "$best"
}
export PHPX_PHP_BINARY=/app/phpx-shim.sh
run "$REPEATS" plain    php -d memory_limit=-1 vendor/bin/phpunit -c phpunit-plain.xml
run "$REPEATS" turbocov php -d memory_limit=-1 -d turbocov.collect_paths=src vendor/bin/phpunit -c phpunit-turbocov.xml
run "$REPEATS" xdebug   php -d memory_limit=-1 -d zend_extension=xdebug.so -d xdebug.mode=coverage vendor/bin/phpunit -c phpunit-xdebug.xml
# phpx interprets the suite and takes minutes, so it runs once whatever the
# repeat count is. PHPX=0 skips it; PHPX_COVER=1 adds its coverage run, which
# writes the statement profile the README compares against Clover.
if [ "${PHPX:-1}" = 1 ]; then
	run 1 phpx /usr/local/bin/phpx run vendor/bin/phpunit -c phpunit-plain.xml
	if [ "$PHPX_COVER" = 1 ]; then
		run 1 phpx+cov /usr/local/bin/phpx --coverfile=/tmp/phpx.cov run vendor/bin/phpunit -c phpunit-plain.xml
	fi
fi
' | sort | while IFS="$(printf '\t')" read -r name ms; do
	awk -v n="$name" -v ms="$ms" 'BEGIN { printf "| %-8s | %6.2fs |\n", n, ms / 1000 }'
done
