name: a copy is deferred until a write splits it
description: >
  php copies an array at assignment but shares the storage until one side
  writes; every first write splits it. The copy holds at the top level and one
  level down, through a by-reference loop, through a reference to an element,
  through a builtin that writes its array argument in place, and through a
  result a binding builds from another array's values. Expected output is php
  8.4's.
runner:
  flatstack: false
---
<?php
// A copy at assignment is deferred: the two names read one storage until a
// write splits it. Every write below has to leave the other name alone.
function show($a, $b) { echo json_encode([$a, $b]), "\n"; }

$a = [1, 2]; $b = $a; $b[] = 3; show($a, $b);

$p = ['k' => [1]]; $q = $p; $q['k'][] = 2; show($p, $q);

$p = ['k' => [3, 1, 2]]; $q = $p; sort($q['k']); show($p, $q);

$p = ['k' => [1]]; $q = $p; unset($q['k']); show($p, $q);

$p = ['k' => [1, 2]]; $q = $p; foreach ($q as &$v) { $v[] = 9; } unset($v); show($p, $q);

$p = ['k' => [1, 2]]; $q = $p; $r = &$q['k']; $r[] = 9; show($p, $q);

// A result a binding builds holds its own handle for an array element.
$a = ['k' => [1]]; $m = array_merge($a, []); $m['k'][0] = 9; show($a, $m);

$a = ['k' => [1]]; $s = array_slice($a, 0); $s['k'][] = 2; show($a, $s);

$v = [1]; $f = array_fill(0, 2, $v); $f[0][] = 9; show($v, $f);

// A function argument is a copy, and so is a return of storage.
function grow(array $x) { $x[] = 'x'; return $x; }
$g = [1]; $h = grow($g); show($g, $h);

class Bag { public $items = ['k' => [1]]; }
$bag = new Bag; $list = $bag->items; $list['k'][] = 2; show($bag->items, $list);

// Both sides write: each splits into its own storage once.
$x = ['k' => [1]]; $y = $x;
$y['k'][] = 2; $x['k'][] = 3; show($x, $y);

// A callback that writes through an element a builtin hands it.
$w = ['k' => [['x' => 1]]]; $v = $w;
array_walk($v['k'], function (&$e) { $e['x'] = 9; });
show($w, $v);

// The other array a variadic in-place builtin writes.
$m1 = ['k' => [2, 1]]; $m2 = ['k' => [1, 2]]; $n1 = $m1; $n2 = $m2;
array_multisort($n1['k'], $n2['k']); show($m1, $n1); show($m2, $n2);

// Two more results a binding builds from values.
$rows = [['id' => 1, 'v' => [1]], ['id' => 2, 'v' => [2]]];
$col = array_column($rows, 'v', 'id'); $col[1][] = 9; show($rows, $col);

$keys = ['a']; $vals = [[1]]; $comb = array_combine($keys, $vals); $comb['a'][] = 2; show($vals, $comb);

// A list() element is a value of its own.
$src = [['k' => 1], 2]; [$one, $two] = $src; $one['k'] = 9; show($src, $one);
---
[[1,2],[1,2,3]]
[{"k":[1]},{"k":[1,2]}]
[{"k":[3,1,2]},{"k":[1,2,3]}]
[{"k":[1]},[]]
[{"k":[1,2]},{"k":[1,2,9]}]
[{"k":[1,2]},{"k":[1,2,9]}]
[{"k":[1]},{"k":[9]}]
[{"k":[1]},{"k":[1,2]}]
[[1],[[1,9],[1]]]
[[1],[1,"x"]]
[{"k":[1]},{"k":[1,2]}]
[{"k":[1,3]},{"k":[1,2]}]
[{"k":[{"x":1}]},{"k":[{"x":9}]}]
[{"k":[2,1]},{"k":[1,2]}]
[{"k":[1,2]},{"k":[2,1]}]
[[{"id":1,"v":[1]},{"id":2,"v":[2]}],{"1":[1,9],"2":[2]}]
[[[1]],{"a":[1,2]}]
[[{"k":1},2],{"k":9}]
