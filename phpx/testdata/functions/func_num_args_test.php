<?php
function f() { return func_num_args() . ':' . json_encode(func_get_args()) . ':' . (func_num_args() > 1 ? func_get_arg(1) : '-'); }
echo f(), f(1), f(1, 'b', 3), "\n";
class K { public function m($a = 0) { return func_num_args(); } }
echo (new K)->m(), (new K)->m(1), (new K)->m(1, 2), "\n";
