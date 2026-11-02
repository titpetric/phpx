#!/usr/bin/env bash
#
# splint-fix.sh runs `splint fix ./...` for the import blocks and puts the
# markdown back the way it found it.
#
# splint fix rewrites markdown as well as Go and takes no flag to leave it
# alone: it pads every table it walks past and reflows prose onto one line.
# Upstream's generated help documents are unpadded, and the fork keeps
# upstream's prose, so a run of it on its own hands the tree a thirteen-file
# diff and the vale gate a finding upstream does not have.
#
# Two callers, for the same reason. migrations/001-module-path.sh runs it
# because the renamed import sorts into a different group than the one it sat
# in, and the lint gate reads the grouping. The fmt job runs it because that is
# what formatting Go means here. A pipeline run has to end on a clean tree:
# the next follow refuses a dirty one.
#
# gofmt is the fallback for a tree with no splint on the path.
set -u -o pipefail

root="$(git rev-parse --show-toplevel)" || exit 1
cd "${root}" || exit 1

if ! command -v splint >/dev/null 2>&1; then
	gofmt -w . || exit 1
	echo "splint-fix: no splint on the path, gofmt wrote the tree"
	exit 0
fi

docs="$(mktemp -d)"
trap 'rm -rf "${docs}"' EXIT

while read -r doc; do
	mkdir -p "${docs}/$(dirname "${doc}")"
	cp "${doc}" "${docs}/${doc}"
done < <(find . -name '*.md' -not -path './.git/*')

splint fix ./... || exit 1

restored=0
while read -r doc; do
	[ -f "${docs}/${doc}" ] || continue
	cmp -s "${doc}" "${docs}/${doc}" && continue
	cp "${docs}/${doc}" "${doc}" || exit 1
	restored=$((restored + 1))
done < <(find . -name '*.md' -not -path './.git/*')

echo "splint-fix: ${restored} markdown files put back the way splint found them"
