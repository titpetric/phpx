#!/usr/bin/env bash
#
# follow.sh takes upstream phpscript's tree and writes the fork over it. The
# integration runs one way: phpscript leads, phpx follows, and nothing travels
# back.
#
# It neither pulls nor merges. A pull asks git to reconcile two edits of one
# file, and the fork's edit of an upstream file is a migration, applied to the
# whole tree at once: git would report a conflict on every file the rename
# touched and the resolution would be "run the step again" every time. So the
# step runs instead. Upstream's version of an upstream file lands verbatim, the
# steps rewrite what they rewrite, and one commit records the result.
#
# Three consequences worth knowing:
#
#   The history is append-only. Every follow is one commit on top, every push
#   is a fast-forward, and a published commit stays published.
#
#   A conflict cannot happen. No file has two authors: upstream owns its files
#   and the fork owns the ones fork_owns names.
#
#   An edit the fork wants inside an upstream file belongs in a migration step.
#   A committed edit to an upstream file disappears on the next follow, because
#   upstream's version of that file lands verbatim.
#
# `atkins -f atkins.phpx.yml phpx:follow` runs this. PHPX_UPSTREAM and
# PHPX_UPSTREAM_BRANCH name the source; the defaults read phpscript main over
# https, without a key. PHPX_FORCE=1 runs the steps again when upstream has not
# moved, which is what a change to a migration step needs.
set -u -o pipefail

upstream_url="${PHPX_UPSTREAM:-https://github.com/titpetric/phpscript.git}"
branch="${PHPX_UPSTREAM_BRANCH:-main}"

root="$(git rev-parse --show-toplevel)" || exit 1
cd "${root}" || exit 1

# A follow over uncommitted work overwrites it with upstream's version. The run
# ends here instead.
if [ -n "$(git status --porcelain)" ]; then
	echo "follow.sh: the tree is dirty, commit or stash before a follow" >&2
	exit 1
fi

git remote get-url upstream >/dev/null 2>&1 || git remote add upstream "${upstream_url}"
git fetch upstream "${branch}" || exit 1

onto="$(git rev-parse "upstream/${branch}")"
short="$(git rev-parse --short "${onto}")"
count="$(git rev-list --count "HEAD..${onto}" 2>/dev/null || echo 0)"

# fork_owns names every file the fork writes by hand. Everything else belongs
# to upstream and arrives from upstream's tree on every follow.
#
# The phpx.go rule is the policy from README.phpx.md read back as code: a
# change the fork needs inside an upstream package goes in a file named for
# the fork, and a file named for the fork survives a follow.
fork_owns() {
	case "$1" in
	README.phpx.md | ACCEPTANCE.md | CHANGELOG.md | atkins.phpx.yml | lefthook.yml) return 0 ;;
	phpx/* | migrations/* | stdlib/phpx/*) return 0 ;;
	phpx.go | phpx_*.go | */phpx.go | */phpx_*.go) return 0 ;;
	# A directory named for the fork holds data the Go files embed (the PHP prelude).
	*/phpx_*/*) return 0 ;;
	esac
	return 1
}

# What upstream dropped, the fork drops. git checkout below writes the paths
# upstream carries and says nothing about the ones it stopped carrying.
paths="$(mktemp)"
trap 'rm -f "${paths}"' EXIT
git ls-tree -r --name-only "upstream/${branch}" >"${paths}"

removed=0
while read -r path; do
	fork_owns "${path}" && continue
	grep -qxF "${path}" "${paths}" && continue
	git rm -q -- "${path}" || exit 1
	echo "follow.sh: dropped ${path}"
	removed=$((removed + 1))
done < <(git ls-files)

# Upstream's version of every path upstream carries, verbatim. A fork-owned
# path is not in that tree, so this reaches none of them.
git checkout "upstream/${branch}" -- . || exit 1

bash migrations/run.sh || exit 1

if [ ! -f CHANGELOG.md ]; then
	echo "follow.sh: CHANGELOG.md is missing, the fork scratchpad is tracked" >&2
	exit 1
fi

stamp="$(git log -1 --format=%cs "${onto}")"
heading="## ${stamp} - upstream ${short}"

# Newest entry first, under the title and above the one before it. A re-run
# over the same upstream tip finds its own heading and adds nothing.
if ! grep -qF "${heading}" CHANGELOG.md; then
	tmp="$(mktemp)"
	{
		head -n 1 CHANGELOG.md
		printf '\n%s\n\n' "${heading}"
		printf -- '- follow `%s`, %s commits ahead of the last one\n' "${short}" "${count}"
		printf -- '- `migrations/run.sh` wrote the generated tree again, %s files dropped\n' "${removed}"
		tail -n +2 CHANGELOG.md
	} >"${tmp}"
	mv "${tmp}" CHANGELOG.md
fi

git add --all || exit 1

# The index decides, rather than `git status`: the checkout above staged
# upstream's version of every file and the steps rewrote the worktree over it,
# so the tree reads as changed until the index is restaged. A follow that
# reproduces the tree it started from has nothing to record.
if git diff --cached --quiet HEAD; then
	echo "follow.sh: upstream ${short} is already what this tree carries"
	exit 0
fi
# PHPX_README lets the lefthook guard through: migrations/002-readme.sh wrote
# that copy. Every other route to a staged README.md is an edit by hand, and
# the next follow throws one away.
PHPX_README=1 git commit --message "chore(phpx): follow upstream ${short}" \
	--message "Upstream's tree landed verbatim and migrations/ rewrote the fork's half over it: the module path, and README.md from README.phpx.md." || exit 1

echo "follow.sh: on upstream ${short}, $(git log -1 --format=%h) written"
