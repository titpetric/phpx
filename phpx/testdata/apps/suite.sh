#!/usr/bin/env bash
# Runs an app's own phpunit test files through the phpunit in its vendor tree.
# The XML configuration loader needs DOMDocument::loadXML and DOMXPath, which
# phpx does not bind, so --no-configuration stands in for phpunit.xml and the
# env that file names is passed here.
# Usage: suite.sh <phpx> <app-dir> <test path> [phpunit args...]
set -u
PHPX=$(realpath "$1"); APP=$(realpath "$2"); shift 2
here=$(cd "$(dirname "$0")" && pwd)
cp "$here/../phpunit/launcher.php" "$APP/phpx_launcher.php" || exit 1
cd "$APP" || exit 1
APP_ENV=${APP_ENV:-testing} \
APP_KEY=${APP_KEY:-base64:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=} \
DB_CONNECTION=${DB_CONNECTION:-sqlite} \
DB_DATABASE=${DB_DATABASE:-:memory:} \
SESSION_DRIVER=${SESSION_DRIVER:-array} \
CACHE_STORE=${CACHE_STORE:-array} \
QUEUE_CONNECTION=${QUEUE_CONNECTION:-sync} \
MAIL_MAILER=${MAIL_MAILER:-array} \
BCRYPT_ROUNDS=${BCRYPT_ROUNDS:-4} \
	"$PHPX" run phpx_launcher.php --no-configuration "$@"
status=$?
rm -f "$APP/phpx_launcher.php"
exit $status
