#!/usr/bin/env bash
#
# Runs an app written in the OOP baseline under php, phpscript and phpx, and
# compares what each answers and what each costs.
#
# phpscript here is upstream's binary, not bin/phpscript: in this tree that is
# a link to phpx, and comparing phpx with itself would read as parity. Build
# upstream once, outside this tree, with the same toolchain:
#
#   git clone --depth 1 https://github.com/titpetric/phpscript /tmp/phpscript
#   (cd /tmp/phpscript && GOFLAGS= CGO_ENABLED=0 go build -o /tmp/phpscript .)
#   phpx/bench/oop/run.sh --phpscript /tmp/phpscript --phpx bin/phpx
#
# One CPU is pinned and the run takes the measurement lock, because another
# agent's build on this box takes the cores a sample runs on.
set -u -o pipefail

here="$(cd "$(dirname "$0")" && pwd)"
root="$(cd "${here}/../../.." && pwd)"
runs=15
php="php"
phpscript=""
phpx="${root}/bin/phpx"

while [ $# -gt 0 ]; do
	case "$1" in
	--runs) runs="$2"; shift 2 ;;
	--php) php="$2"; shift 2 ;;
	--phpscript) phpscript="$2"; shift 2 ;;
	--phpx) phpx="$2"; shift 2 ;;
	*) echo "run.sh: unknown argument $1" >&2; exit 2 ;;
	esac
done

if [ ! -x "${phpx}" ]; then
	echo "run.sh: build ${phpx} first (atkins -f atkins.phpx.yml build)" >&2
	exit 1
fi
if [ -n "${phpscript}" ] && [ "$(realpath "${phpscript}")" = "$(realpath "${phpx}")" ]; then
	echo "run.sh: --phpscript is phpx itself; leave it out or name upstream's build" >&2
	phpscript=""
fi

args=(--php "${php}" --phpx "${phpx}" --runs "${runs}")
[ -n "${phpscript}" ] && args+=(--phpscript "${phpscript}")
exec flock -w 3600 /tmp/phpscript-measure.lock python3 -I "${here}/measure.py" "${args[@]}"
