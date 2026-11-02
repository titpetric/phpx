#!/usr/bin/env bash
#
# lint-delta.sh gates splint on growth rather than on zero.
#
# Upstream's `atkins lint` runs `splint ./...` bare, over a tree its own
# documentation says sits at zero. The splint on this machine reports hundreds
# of godoc/outweighs findings against upstream main, and a fork that inherits
# that tree inherits a red gate it did not earn and cannot clear. The rule the
# fork can keep is the one it owns: add no finding.
#
# The baseline comes out of a worktree at upstream/<branch> on every run. A
# checked-in count would go stale on the first pull that moves upstream's own
# number, and a measured one cannot.
#
# A rule that grows fails the run and names the growth. A rule that shrinks
# passes without a word: the fork is free to fix what upstream left.
set -u -o pipefail

branch="${PHPX_UPSTREAM_BRANCH:-main}"
base_ref="upstream/${branch}"

root="$(git rev-parse --show-toplevel)" || exit 1
cd "${root}" || exit 1

if ! git rev-parse --verify --quiet "${base_ref}" >/dev/null; then
	echo "lint-delta: ${base_ref} is missing, run phpx:pull first" >&2
	exit 1
fi

work="$(mktemp -d)"
base_counts="$(mktemp)"
head_counts="$(mktemp)"
trap 'git worktree remove --force "${work}/base" >/dev/null 2>&1; rm -rf "${work}" "${base_counts}" "${head_counts}"' EXIT

# One line per rule, "name count". splint writes a workflow command per finding
# and title= carries the rule that reported it.
rules() {
	grep -o 'title=[a-z/-]*' | sed 's/^title=//' | sort | uniq -c | awk '{print $2, $1}'
}

git worktree add --detach "${work}/base" "${base_ref}" >/dev/null 2>&1 || {
	echo "lint-delta: no worktree at ${base_ref}" >&2
	exit 1
}

(cd "${work}/base" && splint ./... 2>&1) | rules >"${base_counts}"
splint ./... 2>&1 | rules >"${head_counts}"

status=0
printf '%-26s %9s %6s\n' RULE UPSTREAM HEAD
while read -r rule count; do
	base="$(awk -v r="${rule}" '$1 == r {print $2}' "${base_counts}")"
	base="${base:-0}"
	printf '%-26s %9s %6s' "${rule}" "${base}" "${count}"
	if [ "${count}" -gt "${base}" ]; then
		printf '  grew by %s\n' "$((count - base))"
		status=1
	else
		printf '\n'
	fi
done <"${head_counts}"

if [ "${status}" -ne 0 ]; then
	echo "lint-delta: the fork added a splint finding; fix it here, or fix it upstream" >&2
	exit 1
fi

echo "lint-delta: no rule grew over ${base_ref}"
