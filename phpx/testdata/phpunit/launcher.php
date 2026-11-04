<?php
// Minimal phpunit launcher: skips the entry script's extension gate, so it runs
// on a php built without dom, libxml, tokenizer or xmlwriter, and on phpx.
require_once __DIR__ . '/vendor/autoload.php';
$_SERVER['argv'] = $argv;
exit((new PHPUnit\TextUI\Application)->run($argv));
