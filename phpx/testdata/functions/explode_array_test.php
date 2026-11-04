<?php
$e = explode(",", "x,y,z");
unset($e[1]);
echo implode(",", array_keys($e)), "\n";
echo json_encode($e), "\n";
var_dump(isset($e[2]));
$p = explode("/", "a/b/c");
echo array_pop($p), array_shift($p), count($p), "\n";
array_push($p, "z");
echo implode(",", $p), "\n";
$chars = str_split("abc");
array_unshift($chars, "_");
echo implode("", $chars), "\n";
$parts = preg_split('/[\s,]+/', "a, b  c");
sort($parts);
echo implode("|", $parts), "\n";
$last = array_pop(explode(".", "file.tar.gz"));
echo $last, "\n";
