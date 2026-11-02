#!/usr/bin/env bash
#
# overlay.sh is the other half of migrations/003-overlay.sh: `capture` reads a
# file back out of the tree into phpx/testdata/patches/, and `report` says what each
# overlay carries against the version upstream published.
#
# capture runs at the end of the pipeline. A job in the run edits an overlaid
# file - gofmt regroups an import, mdox rewraps a paragraph, a generator writes
# a document - and without the read-back phpx/testdata/patches/ drifts and the next
# follow installs the stale version over the fixed one.
#
# report takes the upstream version from `upstream/<branch>` and diffs it
# against the overlay, which is the review a copy cannot give on its own: an
# upstream edit to an overlaid file shows up here as a line the fork is
# dropping.
#
#	bash phpx/scripts/overlay.sh capture
#	bash phpx/scripts/overlay.sh report
set -u -o pipefail

overlay="phpx/testdata/patches"
branch="${PHPX_UPSTREAM_BRANCH:-main}"
mode="${1:-capture}"

root="$(git rev-parse --show-toplevel)" || exit 1
cd "${root}" || exit 1

if [ ! -d "${overlay}" ]; then
	echo "overlay.sh: no ${overlay}, nothing to ${mode}"
	exit 0
fi

case "${mode}" in
capture)
	captured=0
	while read -r src; do
		dest="${src#"${overlay}"/}"
		if [ ! -e "${dest}" ]; then
			echo "overlay.sh: ${dest} is missing from the tree, overlay kept" >&2
			continue
		fi
		cmp -s "${dest}" "${src}" && continue
		cp "${dest}" "${src}" || exit 1
		echo "overlay.sh: captured ${dest}"
		captured=$((captured + 1))
	done < <(find "${overlay}" -type f ! -name REMOVED | sort)
	echo "overlay.sh: ${captured} files read back into ${overlay}"
	;;
report)
	if ! git rev-parse --verify --quiet "upstream/${branch}" >/dev/null; then
		echo "overlay.sh: upstream/${branch} is missing, run phpx:follow first" >&2
		exit 1
	fi
	while read -r src; do
		dest="${src#"${overlay}"/}"
		echo "=== ${dest}"
		if ! git cat-file -e "upstream/${branch}:${dest}" 2>/dev/null; then
			echo "    added by the fork, upstream carries no such path"
			continue
		fi
		git diff --no-index --stat -- <(git show "upstream/${branch}:${dest}") "${src}" 2>/dev/null |
			tail -n 1 | sed 's/^/    /'
	done < <(find "${overlay}" -type f ! -name REMOVED | sort)
	if [ -f "${overlay}/REMOVED" ]; then
		echo "=== removed"
		grep -v '^[[:space:]]*#' "${overlay}/REMOVED" | grep -v '^[[:space:]]*$' | sed 's/^/    /'
	fi
	;;
*)
	echo "overlay.sh: unknown mode ${mode}, expected capture or report" >&2
	exit 1
	;;
esac
