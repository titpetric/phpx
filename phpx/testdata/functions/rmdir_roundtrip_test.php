<?php
$d = sys_get_temp_dir() . '/rmdir_probe_' . getmypid();
var_dump(mkdir($d . '/sub/deep', 0777, true));
file_put_contents($d . '/sub/c.txt', 'c');
var_dump(@rmdir($d . '/sub'));
var_dump(unlink($d . '/sub/c.txt'));
var_dump(rmdir($d . '/sub/deep'), rmdir($d . '/sub'), rmdir($d));
var_dump(is_dir($d));
