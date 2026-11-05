<?php
function f() {
    $a = 'orig';
    echo extract(['a' => 'x', 'b' => 'y', '1bad' => 'z', 'c d' => 'w']), $a, $b, "\n";
    echo extract(['a' => 'skipped', 'n' => 'new'], EXTR_SKIP), $a, $n, "\n";
    echo extract(['a' => 'p', 'q' => 'r'], EXTR_PREFIX_SAME, 'dup'), $a, $dup_a, $q, "\n";
    echo extract(['k' => 1], EXTR_PREFIX_ALL, 'pre'), $pre_k, "\n";
    echo extract(['a' => 'only', 'zz' => 1], EXTR_IF_EXISTS), $a, isset($zz) ? 'zz' : 'nozz', "\n";
    $g = function () use (&$a) { $a = 'changed'; };
    $g();
    echo $a, "\n";
}
f();
