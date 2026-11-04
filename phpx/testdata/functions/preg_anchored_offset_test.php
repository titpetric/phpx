<?php
$in = "sum 1 'quoted arg'";
var_dump(preg_match('/\s+/A', $in, $m, 0, 3), $m);
var_dump(preg_match('/\s+/A', $in, $m, 0, 4), $m);
var_dump(preg_match('/x/A', 'ax'));
var_dump(preg_match('/a/A', 'ab'));
var_dump(preg_match_all('/\d/A', '123a4', $m), $m);
var_dump(defined('PCRE_VERSION'), defined('Foo::BAR'));
enum Lvl: int { case Low = 1; }
var_dump(defined(Lvl::class . '::Low'), defined(Lvl::class . '::High'));
