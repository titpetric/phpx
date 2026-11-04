<?php
// Minimal phpunit launcher: skips the entry script's extension gate.
$_SERVER['argv'] = $argv;
exit((new PHPUnit\TextUI\Application)->run($argv));
