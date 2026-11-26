#!/bin/sh
# phpx-shim.sh speaks the part of php's command line that phpunit uses when it
# starts a test in its own process: -d and -c are accepted and dropped, -f
# names the script, and everything after it is the script's. phpx answers
# PHP_BINARY with this path, so a child process starts phpx the same way.
while [ $# -gt 0 ]; do
	case "$1" in
	-d|-c) shift 2 ;;
	-d*|-c*) shift ;;
	-n|-q|-E) shift ;;
	-f) shift; break ;;
	--) shift; break ;;
	-*) shift ;;
	*) break ;;
	esac
done
exec /usr/local/bin/phpx run "$@"
