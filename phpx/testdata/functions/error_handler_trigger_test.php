<?php
$log = [];
$prev = set_error_handler(function ($no, $str) use (&$log) { $log[] = "$no:$str"; return true; });
var_dump($prev);
trigger_error("warn", E_USER_WARNING);
trigger_error("dep", E_USER_DEPRECATED);
trigger_error("note");
$second = set_error_handler(function ($no, $str) use (&$log) { $log[] = "second $no:$str"; return true; }, E_USER_NOTICE);
var_dump($second instanceof Closure);
trigger_error("n2");
restore_error_handler();
trigger_error("back", E_USER_WARNING);
restore_error_handler();
echo implode(" | ", $log), "\n";
