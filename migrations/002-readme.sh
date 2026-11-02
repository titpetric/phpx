#!/usr/bin/env bash
#
# 002-readme.sh writes README.md from README.phpx.md.
#
# Both projects edit README.md, and the fork publishes the phpx one. A copy
# resolves the conflict before git reports it: the pull throws the upstream
# version away and this step writes the fork's version over it.
set -u -o pipefail

root="$(git rev-parse --show-toplevel)" || exit 1
cd "${root}" || exit 1

if [ ! -f README.phpx.md ]; then
	echo "002: README.phpx.md is missing, the fork prompt is tracked" >&2
	exit 1
fi

if cmp -s README.phpx.md README.md; then
	echo "002: README.md is the copy already"
	exit 0
fi

cp README.phpx.md README.md
echo "002: README.md written from README.phpx.md"
