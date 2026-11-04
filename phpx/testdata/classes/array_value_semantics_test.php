<?php
// The class keeps the program on the interpreter: the bytecode engine shares the handle.
function f(array $a, array $b, bool $canon) { if ($canon) { sort($a); sort($b); } return $a == $b; }
var_dump(f([3, 1, 2], [1, 2, 3], true));
$x = [3, 1, 2]; $y = $x; sort($y); echo json_encode([$x, $y]), "\n";
function g(array $a) { sort($a); return $a; } $z = [3, 1, 2]; echo json_encode([g($z), $z]), "\n";
class Bag { private array $items = [3, 1, 2]; public function items(): array { return $this->items; } public function first() { return $this->items[0]; } }
$bag = new Bag; $list = $bag->items(); sort($list); echo json_encode([$list, $bag->items()]), "\n";
function nested(array $a) { $a['x']['y'] = 9; return $a; } $n = ['x' => ['y' => 1]]; $m = nested($n); echo json_encode([$n, $m]), "\n";
function readOnly(array $a) { return count($a); } $big = range(1, 5); echo readOnly($big), "\n";
$p = ['k' => [1]]; $q = $p; $q['k'][] = 2; echo json_encode([$p, $q]), "\n";
