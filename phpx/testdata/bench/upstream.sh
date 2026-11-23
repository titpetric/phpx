#!/bin/sh
# Usage: upstream.sh [phpx binary] [phpscript binary]
#
# Times one workload written two ways across three runtimes: php, upstream
# phpscript and this fork. The phpscript binary is built from the upstream/main
# ref of the checkout when the second argument is empty, so the comparison is
# against the tree the fork last followed.
#
# flat.php is phpscript's style and oop.php is phpx's; upstream refuses the
# second at its first interface, which is the point of the fork.
set -u
here=$(cd "$(dirname "$0")" && pwd)
phpx=${1:-phpx}
phpscript=${2:-}

tmp=
cleanup() {
	[ -n "$tmp" ] || return 0
	git -C "$tmp/root" worktree remove --force "$tmp/src" >/dev/null 2>&1
	rm -rf "$tmp"
}
trap cleanup EXIT INT TERM

if [ -z "$phpscript" ]; then
	root=$(git -C "$here" rev-parse --show-toplevel) || exit 1
	tmp=$(mktemp -d)
	mkdir -p "$tmp/root"
	ln -sfn "$root/.git" "$tmp/root/.git"
	git -C "$root" worktree add --detach "$tmp/src" upstream/main >/dev/null 2>&1 || exit 1
	(cd "$tmp/src" && GOFLAGS= go build -o "$tmp/phpscript" .) || exit 1
	# The binary resolves a sibling bin/phpscript when it runs a program.
	mkdir -p "$tmp/src/bin"
	ln -sf "$tmp/phpscript" "$tmp/src/bin/phpscript"
	phpscript=$tmp/src/bin/phpscript
fi

best() {
	b=999999
	for _ in 1 2 3; do
		s=$(date +%s%N)
		"$@" >/dev/null 2>&1
		e=$(( ($(date +%s%N) - s) / 1000000 ))
		if [ "$e" -lt "$b" ]; then b=$e; fi
	done
	echo "$b"
}

cd "$here" || exit 1
printf '| Program | php | phpscript | phpx |\n'
printf '|---|---:|---:|---:|\n'
for f in flat oop; do
	if "$phpscript" "$f.php" >/dev/null 2>&1; then
		up="$(best "$phpscript" "$f.php") ms"
	else
		up="refused"
	fi
	printf '| `%s.php` | %s ms | %s | %s ms |\n' "$f" "$(best php "$f.php")" "$up" "$(best "$phpx" "$f.php")"
done
