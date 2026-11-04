<?php
// phpunit launcher for a coverage run: after the suites end it includes every
// source file of the nine libraries, so a file no test reached still counts in
// the denominator, as php-code-coverage's whitelist does.
require_once __DIR__ . '/vendor/autoload.php';
$_SERVER['argv'] = $argv;
register_shutdown_function(function () {
    foreach (require __DIR__ . '/coverage_libraries.php' as $dirs) {
        foreach ($dirs as $dir) {
            $it = new RecursiveIteratorIterator(new RecursiveDirectoryIterator(__DIR__ . '/' . $dir, FilesystemIterator::SKIP_DOTS));
            foreach ($it as $f) {
                if (substr($f->getFilename(), -4) !== '.php' || strpos($f->getPathname(), '/Resources/') !== false) { continue; }
                try { include_once $f->getPathname(); } catch (\Throwable $e) { }
            }
        }
    }
});
exit((new PHPUnit\TextUI\Application)->run($argv));
