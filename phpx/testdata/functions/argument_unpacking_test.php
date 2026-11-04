<?php
function sum(int ...$n) { return array_sum($n); }
function three($a, $b, $c) { return "$a-$b-$c"; }
class A { public $v; function __construct($x, $y) { $this->v = $x . $y; }
  function m($a, $b) { return $a + $b; } static function s(...$r) { return count($r); } }
$args = [1, 2, 3];
echo sum(...$args), "\n";
echo sum(0, ...$args), "\n";
echo three(...[4, 5, 6]), "\n";
$a = new A(...['x', 'y']);
echo $a->v, "\n";
echo $a->m(...[1, 2]), "\n";
echo A::s(...$args, ...$args), "\n";
$f = function ($p, $q) { return $p * $q; };
echo $f(...[3, 4]), "\n";
echo max(...$args), "\n";
echo sum(...[]), "\n";
