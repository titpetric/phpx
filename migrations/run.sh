#!/usr/bin/env bash
#
# run.sh runs every migration step in order and stops at the first one that
# fails. The steps rewrite the tree and commit nothing; `phpx/scripts/pull.sh`
# commits what they produce.
set -u -o pipefail

root="$(git rev-parse --show-toplevel)" || exit 1
cd "${root}" || exit 1

steps="$(find migrations -maxdepth 1 -name '[0-9]*.sh' | sort)"
if [ -z "${steps}" ]; then
	echo "run.sh: no step under migrations/" >&2
	exit 1
fi

for step in ${steps}; do
	echo "run.sh: ${step}"
	bash "${step}" || {
		echo "run.sh: ${step} failed" >&2
		exit 1
	}
done
