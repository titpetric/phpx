#!/usr/bin/env bash
#
# 005-phpunit-cache.sh keeps phpunit's result cache out of the tree.
#
# `phpx/testdata/phpunit/phpunit.xml` names a cache directory, and phpunit
# writes its result cache and its test-result cache there on every run. git
# sees an untracked folder, phpx:follow refuses a dirty tree, and a commit that
# stages it carries a build artefact into the repository. The entry names the
# directory and leaves the fixtures alone.
set -u -o pipefail

root="$(git rev-parse --show-toplevel)" || exit 1
cd "${root}" || exit 1

pattern='phpx/testdata/phpunit/.phpunit.cache/'

if [ ! -f .gitignore ]; then
	echo "005: .gitignore is missing" >&2
	exit 1
fi

if grep -qxF "${pattern}" .gitignore; then
	echo "005: .gitignore already ignores ${pattern}"
	exit 0
fi

{
	printf '\n'
	printf '# phpunit writes its result cache beside the configuration that names\n'
	printf '# it, and the fixtures are the tree content. The cache is not.\n'
	printf '%s\n' "${pattern}"
} >>.gitignore || exit 1

echo "005: .gitignore ignores ${pattern}"
