#!/usr/bin/env bash
#
# 001-module-path.sh renames the go module and every import of it.
#
# `go mod edit -module` is the whole of what the go toolchain offers here: it
# writes the module line and leaves the import paths alone. The import pass is
# a text rewrite over the files that name the old path, which is what the go
# documentation says to do for a module path change. splint fix regroups the
# import blocks afterwards, because the new path sorts elsewhere and the lint
# gate reads the grouping.
#
# Nothing but the module path changes. The word "phpscript" stays in help text,
# in a fixture and in every document, and a rewrite of those would edit the
# behaviour this fork inherits.
#
# PHPX_MODULE names the target path. The default matches the git remote.
set -u -o pipefail

old="github.com/titpetric/phpscript"
new="${PHPX_MODULE:-github.com/titpetric/phpx}"

root="$(git rev-parse --show-toplevel)" || exit 1
cd "${root}" || exit 1

if [ "${old}" = "${new}" ]; then
	echo "001: the target path is the source path, nothing to rewrite"
	exit 0
fi

go mod edit -module "${new}" || exit 1

files="$(grep -rlF --include='*.go' "${old}" . || true)"
if [ -z "${files}" ]; then
	echo "001: ${new}, no import of ${old} left"
	exit 0
fi

printf '%s\n' "${files}" | xargs sed -i "s|${old}|${new}|g" || exit 1
echo "001: ${new}, $(printf '%s\n' "${files}" | wc -l | tr -d ' ') files rewritten"

# The import blocks group by origin and the renamed path belongs to a different
# group than the one it sat in. splint fix writes the grouping the gate reads;
# gofmt is the fallback for a tree without splint on the path.
if command -v splint >/dev/null 2>&1; then
	splint fix ./... || exit 1
else
	gofmt -w . || exit 1
fi

go build ./... || exit 1
