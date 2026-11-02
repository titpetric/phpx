#!/usr/bin/env bash
#
# 003-overlay.sh copies the fork's version of an upstream file over the tree.
#
# phpx/testdata/patches/ mirrors the source tree:
# phpx/testdata/patches/runner/runner.go is the fork's runner/runner.go. A
# follow lands upstream's version of that file and this step writes the fork's
# over it, which is how a local change to an upstream file survives a follow.
#
# Under testdata because the go tool ignores any path with a testdata element.
# Without it `go build ./...` compiles the copies as packages of their own and
# splint reads them as source: the first run of this mechanism reported a
# 1218-line file and three missing test files, all of them in the overlay.
#
# A whole file rather than a diff. A diff fails on context drift and leaves the
# tree half-migrated; a copy always applies, and the cost is that an upstream
# edit to an overlaid file goes unnoticed. `phpx:overlay:report` prints what
# each overlay is dropping, and reading it after a follow is the fork's job.
#
# phpx/testdata/patches/REMOVED lists the upstream paths the fork drops. A path there
# cannot also be overlaid: the removal runs first and the copy would bring it
# back.
#
# An overlay whose destination is missing is an addition, which is how the fork
# adds a file inside an upstream directory: follow.sh drops a tracked path
# upstream does not carry, and this step writes it again.
set -u -o pipefail

overlay="phpx/testdata/patches"
removed="${overlay}/REMOVED"

root="$(git rev-parse --show-toplevel)" || exit 1
cd "${root}" || exit 1

if [ ! -d "${overlay}" ]; then
	echo "003: no ${overlay}, nothing to copy"
	exit 0
fi

if [ -f "${removed}" ]; then
	while read -r path; do
		case "${path}" in '' | '#'*) continue ;; esac
		if [ -e "${path}" ]; then
			rm -f "${path}" || exit 1
			echo "003: removed ${path}"
		fi
	done <"${removed}"
fi

copied=0
added=0
while read -r src; do
	dest="${src#"${overlay}"/}"
	if [ ! -e "${dest}" ]; then
		added=$((added + 1))
	fi
	mkdir -p "$(dirname "${dest}")" || exit 1
	cp "${src}" "${dest}" || exit 1
	copied=$((copied + 1))
done < <(find "${overlay}" -type f ! -name REMOVED | sort)

echo "003: ${copied} files copied over the tree, ${added} of them added"
