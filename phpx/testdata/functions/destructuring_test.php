<?php
[$a, $b] = [1, 2];
echo $a, $b, "\n";
[$a, $b] = [$b, $a];
echo $a, $b, "\n";
['x' => $x, 'y' => $y] = ['x' => 10, 'y' => 20];
echo $x + $y, "\n";
[$p, [$q, $r]] = [1, [2, 3]];
echo $p, $q, $r, "\n";
['k' => [$m, $n], 'z' => $z] = ['k' => [4, 5], 'z' => 6];
echo $m, $n, $z, "\n";
foreach ([[1, 2], [3, 4]] as [$u, $v]) {
    echo $u + $v, "\n";
}
foreach ([['id' => 1, 'pt' => [7, 8]], ['id' => 2, 'pt' => [9, 10]]] as ['id' => $id, 'pt' => [$px, $py]]) {
    echo $id, ':', $px, ',', $py, "\n";
}
foreach (['a' => [1, 2]] as $key => [$f, $g]) {
    echo $key, $f, $g, "\n";
}
class Box { public $v; public array $l = []; }
$o = new Box;
[$o->v, $o->l[0]] = [5, 6];
echo $o->v, $o->l[0], "\n";
function pair() { return [7, 8]; }
[$s, $t] = pair();
echo $s . $t, "\n";
$res = [$w, [$w2]] = [1, [2]];
echo count($res), $w, $w2, "\n";
function src() { echo "src\n"; return ['a' => [1, 2]]; }
['a' => [$h, $i]] = src();
echo $h, $i, "\n";
[$one, , $three] = [1, 2, 3];
echo $one, $three, "\n";
