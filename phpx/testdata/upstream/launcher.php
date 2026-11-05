<?php
// Minimal phpunit launcher: skips the entry script's extension gate, so it runs
// on a php built without dom, libxml, tokenizer or xmlwriter, and on phpx. It
// drives phpunit 10 and later through Application and phpunit 9 through Command.
require_once __DIR__ . '/shims/RecursionContext.php';
require_once __DIR__ . '/vendor/autoload.php';
$_SERVER['argv'] = $argv;
if (class_exists(PHPUnit\TextUI\Application::class)) {
    exit((new PHPUnit\TextUI\Application)->run($argv));
}
exit(PHPUnit\TextUI\Command::main(false));
