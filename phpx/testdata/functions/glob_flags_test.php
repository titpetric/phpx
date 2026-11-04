<?php
$d = sys_get_temp_dir() . '/glob_flags_' . getmypid();
mkdir($d . '/sub', 0777, true);
foreach (['a.txt', 'b.php', 'c.md'] as $f) { file_put_contents("$d/$f", 'x'); }
$strip = fn(array $l) => array_map(fn($p) => substr($p, strlen($d)), $l);
print_r($strip(glob("$d/*.txt")));
print_r($strip(glob("$d/*.{txt,php}", GLOB_BRACE)));
print_r($strip(glob("$d/*", GLOB_ONLYDIR)));
print_r($strip(glob("$d/*", GLOB_MARK | GLOB_ONLYDIR)));
print_r(count(glob("$d/none*")));
print_r($strip(glob("$d/none*", GLOB_NOCHECK)));
foreach (['a.txt', 'b.php', 'c.md'] as $f) { unlink("$d/$f"); }
rmdir($d . '/sub'); rmdir($d);
