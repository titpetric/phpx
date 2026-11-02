#!/usr/bin/env bash
#
# pull.sh rebases the fork onto upstream phpscript and writes the generated
# tree again.
#
# The fork carries two kinds of commit. One holds the files upstream never
# writes - README.phpx.md, atkins.phpx.yml, lefthook.yml, migrations/, phpx/ -
# and a rebase replays it without a conflict to resolve. The other holds what
# migrations/ rewrites, the module path above all, and that one touches most of
# the tree. Rebasing it would be one resolution per file, and this script drops
# it before the rebase and writes it again after.
#
# A conflict means an upstream commit reached a path the fork owns. The run
# stops there for an agent to read, because the resolution is a decision.
#
# `atkins -f atkins.phpx.yml phpx:pull` runs this. PHPX_UPSTREAM and
# PHPX_UPSTREAM_BRANCH name the source; the defaults read phpscript main over
# https, without a key. PHPX_FORCE=1 runs the steps again when upstream has not
# moved, which is what a change to a migration step needs.
set -u -o pipefail

upstream_url="${PHPX_UPSTREAM:-https://github.com/titpetric/phpscript.git}"
branch="${PHPX_UPSTREAM_BRANCH:-main}"
marker="chore(phpx): generated"

root="$(git rev-parse --show-toplevel)" || exit 1
cd "${root}" || exit 1

# A rebase over uncommitted work either refuses to start or carries it into a
# conflict resolution that drops it. The run ends here instead.
if [ -n "$(git status --porcelain)" ]; then
	echo "pull.sh: the tree is dirty, commit or stash before a rebase" >&2
	exit 1
fi

git remote get-url upstream >/dev/null 2>&1 || git remote add upstream "${upstream_url}"
git fetch upstream "${branch}" || exit 1

onto="$(git rev-parse "upstream/${branch}")"
short="$(git rev-parse --short "${onto}")"

if git merge-base --is-ancestor "${onto}" HEAD && [ "${PHPX_FORCE:-}" != "1" ]; then
	echo "pull.sh: upstream ${short} is already in, PHPX_FORCE=1 runs the steps again"
	exit 0
fi

# Every generated commit at the tip comes off. What is left is the fork's own
# files and the tree upstream published, which is the input the steps expect.
for _ in $(seq 1 10); do
	case "$(git log -1 --format=%s)" in
	"${marker}"*) ;;
	*) break ;;
	esac
	echo "pull.sh: dropping $(git log -1 --format='%h %s')"
	git reset --hard HEAD~1 || exit 1
done

before="$(git rev-parse HEAD)"

git rebase "upstream/${branch}" || {
	unmerged="$(git diff --name-only --diff-filter=U)"
	echo "pull.sh: the rebase stopped, resolve it by hand:" >&2
	echo "${unmerged:-see git status}" >&2
	exit 1
}

bash migrations/run.sh || exit 1

count="$(git rev-list --count "${before}..${onto}" 2>/dev/null || echo 0)"
stamp="$(git log -1 --format=%cs "${onto}")"
heading="## ${stamp} - upstream ${short}"

if [ ! -f CHANGELOG.md ]; then
	echo "pull.sh: CHANGELOG.md is missing, the fork scratchpad is tracked" >&2
	exit 1
fi

# Newest entry first, under the title and above the one before it. A re-run
# over the same upstream tip finds its own heading and adds nothing. The entry
# commits on its own: it is the fork's record and a rebase has to replay it,
# which the generated commit below never is.
if ! grep -qF "${heading}" CHANGELOG.md; then
	tmp="$(mktemp)"
	trap 'rm -f "${tmp}"' EXIT
	{
		head -n 1 CHANGELOG.md
		printf '\n%s\n\n' "${heading}"
		printf -- '- rebase onto `%s`, %s commits from upstream\n' "${short}" "${count}"
		printf -- '- `migrations/run.sh` wrote the generated tree again\n'
		tail -n +2 CHANGELOG.md
	} >"${tmp}"
	mv "${tmp}" CHANGELOG.md
	git add CHANGELOG.md || exit 1
	git commit --message "chore(phpx): rebase onto upstream ${short}" \
		--message "Upstream main moved and the fork follows it. The entry names the range; the generated tree is the commit above this one." || exit 1
fi

if [ -z "$(git status --porcelain)" ]; then
	echo "pull.sh: upstream ${short} changed nothing the steps rewrite"
	exit 0
fi

git add --all || exit 1
# PHPX_README lets the lefthook guard through: migrations/002-readme.sh wrote
# that copy. Every other route to a staged README.md is an edit by hand, and
# the next pull throws one away.
PHPX_README=1 git commit --message "${marker} tree from migrations" \
	--message "The module path and README.md come out of migrations/, not out of an edit. The next pull drops this commit and writes it again." || exit 1

echo "pull.sh: on upstream ${short}, $(git log -1 --format=%h) generated"
