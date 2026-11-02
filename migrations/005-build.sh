#!/usr/bin/env bash
#
# 005-build.sh compiles the assembled tree.
#
# The build cannot live in 001: the fork's own files in upstream packages
# (parser/phpx_traits.go, runner/phpx_*.go) read the model fields the overlay
# adds, and 003-overlay.sh has not copied those yet when 001 runs. The check
# belongs where the tree is whole, which is after every step has rewritten it.
set -u -o pipefail

root="$(git rev-parse --show-toplevel)" || exit 1
cd "${root}" || exit 1

go build ./... || exit 1
