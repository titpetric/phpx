#!/bin/sh
# bench.sh times the destructor's reachability walk: three scripts, each of
# which drops references to instances whose class declares __destruct, so the
# runtime walks what a drop may have freed. php runs the same scripts as the
# reference.
#
#   ./bench.sh [phpx-binary]
#
# A run prints milliseconds, best of three.
set -eu

here=$(cd "$(dirname "$0")" && pwd)
phpx=${1:-phpx}
[ -x "$phpx" ] || phpx=$(command -v "$phpx")

time_it() {
	best=
	for _ in 1 2 3; do
		s=$(date +%s%N)
		"$@" >/dev/null 2>&1 || true
		e=$(date +%s%N)
		ms=$(( (e - s) / 1000000 ))
		if [ -z "$best" ] || [ "$ms" -lt "$best" ]; then best=$ms; fi
	done
	echo "$best"
}

for script in graph static drop; do
	for n in 1000 2000 4000; do
		phpm=$(time_it php "$here/$script.php" "$n")
		phpxm=$(time_it "$phpx" run "$here/$script.php" "$n")
		printf '%s\t%s\tphp=%sms\tphpx=%sms\n' "$script" "$n" "$phpm" "$phpxm"
	done
done
