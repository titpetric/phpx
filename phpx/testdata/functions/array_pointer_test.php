<?php
$a = ['x' => 1, 'y' => 2, 'z' => 3];
var_dump(current($a), key($a));
var_dump(next($a), key($a));
var_dump(next($a), next($a), key($a), current($a));
var_dump(reset($a), key($a));
var_dump(end($a), key($a));
var_dump(prev($a), key($a));
$e = [];
var_dump(current($e), key($e), reset($e), end($e));
$list = [10, 20, 30];
while (($v = current($list)) !== false) {
    echo key($list), '=', $v, ' ';
    next($list);
}
echo "\n";
$args = ['--a', '--b'];
$arg = current($args);
while ($arg !== false) { echo $arg, ';'; $arg = next($args); }
echo "\n";
