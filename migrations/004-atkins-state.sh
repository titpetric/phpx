#!/usr/bin/env bash
#
# 004-atkins-state.sh keeps atkins' own run database out of the tree.
#
# atkins opens .atkins/atkins.db beside the tracked skills it reads, and SQLite
# writes -shm and -wal files there while a run holds one. git sees three
# untracked paths, phpx:follow refuses a dirty tree, and a commit that stages
# them carries atkins' run state into the repository. The entry names the
# database and leaves the tracked skills alone.
set -u -o pipefail

root="$(git rev-parse --show-toplevel)" || exit 1
cd "${root}" || exit 1

pattern='.atkins/atkins.db*'

if [ ! -f .gitignore ]; then
	echo "004: .gitignore is missing" >&2
	exit 1
fi

if grep -qxF "${pattern}" .gitignore; then
	echo "004: .gitignore already ignores ${pattern}"
	exit 0
fi

{
	printf '\n'
	printf '# atkins keeps its run database beside the tracked skills under\n'
	printf '# .atkins/, and SQLite writes -shm and -wal files beside it while a\n'
	printf '# run holds one. They are run state and never tree content.\n'
	printf '%s\n' "${pattern}"
} >>.gitignore || exit 1

echo "004: .gitignore ignores ${pattern}"
