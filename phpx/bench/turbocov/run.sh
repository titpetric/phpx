#!/bin/sh
# run.sh measures the commonmark suite turbocov benchmarks, four ways: php
# without coverage, php with turbocov, php with xdebug's coverage mode, and
# phpx. Every leg runs the same checkout in the image the Dockerfile builds.
#
# turbocov runs on php alone and writes Clover. phpx writes its own statement
# profile; the two reports are the comparison, and nothing of turbocov is
# loaded into phpx.
#
#   ./run.sh [repeats]      repeats is 3 by default, and a leg reports its best.
#   PHPX_COVER=1 ./run.sh   also collects phpx's own profile, which is compared
#                           against the Clover report turbocov wrote.
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
cp "$here/Dockerfile" "$here"/phpunit-*.xml "$here/phpx-shim.sh" "$here/phpx.yml" "$ctx/"
(cd "$repo" && go build -tags phpx -o "$ctx/phpx" .)

docker build --quiet -t "$image" "$ctx"

container=phpx-turbocov-run
docker rm -f "$container" >/dev/null 2>&1 || true
docker run --name "$container" -e REPEATS="$repeats" -e PHPX_COVER="${PHPX_COVER:-0}" "$image" sh -c '
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
cp build/logs/clover.xml /tmp/clover-php.xml
run "$REPEATS" xdebug   php -d memory_limit=-1 -d zend_extension=xdebug.so -d xdebug.mode=coverage vendor/bin/phpunit -c phpunit-xdebug.xml
# phpx interprets the suite and takes minutes, so it runs once whatever the
# repeat count is. PHPX=0 skips it; PHPX_COVER=1 adds its coverage run, which
# writes the statement profile compare.py reads beside the Clover report.
if [ "${PHPX:-1}" = 1 ]; then
	run 1 phpx /usr/local/bin/phpx -f phpx.yml run vendor/bin/phpunit -c phpunit-plain.xml
	if [ "$PHPX_COVER" = 1 ]; then
		run 1 phpx+cov /usr/local/bin/phpx -f phpx.yml --coverage --coverfile=/tmp/phpx.cov run vendor/bin/phpunit -c phpunit-plain.xml
	fi
fi
' | sort | while IFS="$(printf '\t')" read -r name ms; do
	awk -v n="$name" -v ms="$ms" 'BEGIN { printf "| %-8s | %6.2fs |\n", n, ms / 1000 }'
done

# The comparison: turbocov's Clover from the php leg, and phpx's own profile
# from its coverage leg. compare.py reads a Clover report beside a profile.
docker cp "$container":/tmp/clover-php.xml "$here/clover-php.xml" 2>/dev/null || true
docker cp "$container":/tmp/phpx.cov "$here/phpx.cov" 2>/dev/null || true
docker rm -f "$container" >/dev/null 2>&1 || true

if [ -f "$here/clover-php.xml" ] && [ -f "$here/phpx.cov" ]; then
	python3 "$here/compare.py" "$here/clover-php.xml" "$here/phpx.cov"
fi
