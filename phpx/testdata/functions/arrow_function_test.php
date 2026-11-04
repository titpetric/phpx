<?php
$add = fn($a, $b) => $a + $b;
echo $add(2, 3), "\n";

$n = 10;
$addn = fn($x) => $x + $n;
echo $addn(5), "\n";

$n = 100; // by value: the closure keeps 10
echo $addn(5), "\n";

echo implode(",", array_map(fn($x) => $x * $n, [1, 2, 3])), "\n";

$nested = fn($x) => fn($y) => $x + $y + $n;
echo $nested(1)(2), "\n";

$typed = fn(int $x): int => $x * 2;
echo $typed(21), "\n";

class Counter {
    private $step = 3;
    public function make() { return fn($x) => $x + $this->step; }
    public static function sq() { return static fn($x) => $x * $x; }
}
echo (new Counter)->make()(4), "\n";
echo Counter::sq()(7), "\n";

$ternary = fn($x) => $x > 1 ? "big" : "small";
echo $ternary(2), $ternary(0), "\n";

$byParam = fn($n) => $n; // a parameter shadows the captured name
echo $byParam(1), "\n";

$cb = fn() => $missing ?? "unset";
echo $cb(), "\n";

$rows = [3, 1, 2];
usort($rows, fn($a, $b) => $b - $a);
echo implode(",", $rows), "\n";
