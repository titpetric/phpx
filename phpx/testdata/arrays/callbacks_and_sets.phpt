name: array_u*, array_walk, natsort and friends
description: >
  The array_diff and array_intersect family with comparison callbacks, array_merge_recursive, array_pad, array_walk with a by-reference value, natsort and the 8.4 predicates. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
$a = ['a' => 'green', 'b' => 'brown', 'c' => 'blue', 'red', 'Green'];
$b = ['a' => 'GREEN', 'B' => 'brown', 'yellow', 'red'];
$ci = fn($x, $y) => strcasecmp($x, $y);
print_r(array_udiff($a, $b, $ci));
print_r(array_diff_ukey($a, $b, $ci));
print_r(array_diff_uassoc($a, $b, $ci));
print_r(array_udiff_assoc($a, $b, $ci));
print_r(array_udiff_uassoc($a, $b, $ci, $ci));
print_r(array_uintersect($a, $b, $ci));
print_r(array_intersect_assoc(['a' => 'green', 'b' => 'brown'], ['a' => 'green', 'b' => 'yellow']));
print_r(array_uintersect_assoc($a, $b, $ci));
print_r(array_intersect_ukey($a, $b, $ci));
print_r(array_intersect_uassoc($a, $b, $ci));
print_r(array_uintersect_uassoc($a, $b, $ci, $ci));
print_r(array_merge_recursive(['a' => [1], 'b' => 2, 5], ['a' => [3], 'b' => 4, 6], ['c' => 'x']));
print_r(array_pad([1, 2], 4, 0)); print_r(array_pad([1, 2], -4, 0)); print_r(array_pad(['k' => 1], 3, 'p'));
print_r(array_change_key_case(['Ab' => 1, 'cD' => 2, 5 => 3], CASE_UPPER));
print_r(array_count_values(['a', 'b', 'a', 1, '1', 1.5]));
var_dump(array_find([1, 2, 3, 4], fn($v) => $v > 2), array_find_key(['x' => 1, 'y' => 5], fn($v) => $v > 2), array_any([1, 2], fn($v) => $v > 1), array_all([1, 2], fn($v) => $v > 1), array_find([1], fn($v) => $v > 5));
$w = [1, 2, 3]; array_walk($w, function (&$v, $k, $m) { $v = $v * $m + $k; }, 10); print_r($w);
$n = ['a' => [1, 2], 'b' => 3]; array_walk_recursive($n, function (&$v) { $v++; }); print_r($n);
$s = ['img12.png', 'img10.png', 'IMG2.png', 'img1.png']; natsort($s); print_r($s); natcasesort($s); print_r($s);
---
Array
(
    [c] => blue
)
Array
(
    [c] => blue
)
Array
(
    [a] => green
    [c] => blue
    [0] => red
    [1] => Green
)
Array
(
    [b] => brown
    [c] => blue
    [0] => red
    [1] => Green
)
Array
(
    [c] => blue
    [0] => red
    [1] => Green
)
Array
(
    [a] => green
    [b] => brown
    [0] => red
    [1] => Green
)
Array
(
    [a] => green
)
Array
(
    [a] => green
)
Array
(
    [a] => green
    [b] => brown
    [0] => red
    [1] => Green
)
Array
(
    [b] => brown
)
Array
(
    [a] => green
    [b] => brown
)
Array
(
    [a] => Array
        (
            [0] => 1
            [1] => 3
        )

    [b] => Array
        (
            [0] => 2
            [1] => 4
        )

    [0] => 5
    [1] => 6
    [c] => x
)
Array
(
    [0] => 1
    [1] => 2
    [2] => 0
    [3] => 0
)
Array
(
    [0] => 0
    [1] => 0
    [2] => 1
    [3] => 2
)
Array
(
    [k] => 1
    [0] => p
    [1] => p
)
Array
(
    [AB] => 1
    [CD] => 2
    [5] => 3
)
Array
(
    [a] => 2
    [b] => 1
    [1] => 2
)
int(3)
string(1) "y"
bool(true)
bool(false)
NULL
Array
(
    [0] => 10
    [1] => 21
    [2] => 32
)
Array
(
    [a] => Array
        (
            [0] => 2
            [1] => 3
        )

    [b] => 4
)
Array
(
    [2] => IMG2.png
    [3] => img1.png
    [1] => img10.png
    [0] => img12.png
)
Array
(
    [3] => img1.png
    [2] => IMG2.png
    [1] => img10.png
    [0] => img12.png
)
