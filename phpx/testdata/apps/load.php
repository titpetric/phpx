<?php
// Test 3: autoload one class, interface, trait or enum of the app. Run from the
// app root with the runtime under test:
//   php phpx_load.php            lists the names under app/
//   php phpx_load.php 'A\B\C'    loads one: prints OK, or ERR and the first line of the error
// load.sh starts one process per name, so a fatal in one class costs one class.
require 'vendor/autoload.php';
if (!isset($argv[1])) {
	$it = new RecursiveIteratorIterator(new RecursiveDirectoryIterator('app', FilesystemIterator::SKIP_DOTS));
	$files = [];
	foreach ($it as $f) {
		if (substr($f->getPathname(), -4) === '.php') {
			$files[] = $f->getPathname();
		}
	}
	sort($files);
	foreach ($files as $path) {
		$src = file_get_contents($path);
		if (preg_match('/^namespace\s+([^;\s]+);/m', $src, $ns) && preg_match('/^(?:final\s+|abstract\s+|readonly\s+)*(class|interface|trait|enum)\s+(\w+)/m', $src, $m)) {
			echo $ns[1] . '\\' . $m[2], "\n";
		}
	}
	exit(0);
}
$name = $argv[1];
try {
	if (class_exists($name) || interface_exists($name) || trait_exists($name) || enum_exists($name)) {
		echo "OK\n";
	} else {
		echo "ERR not found\n";
	}
} catch (Throwable $e) {
	echo 'ERR ', preg_replace('/\s+/', ' ', get_class($e) . ': ' . $e->getMessage()), "\n";
}
