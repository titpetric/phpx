<?php
class T {
    public function run() {
        $x = 1; $inc = function () use (&$x) { $x++; }; $inc(); $inc();
        $acc = []; $add = function ($v) use (&$acc) { $acc[] = $v; }; $add('a'); $add('b');
        $n = 0; $byVal = function () use ($n) { $n++; return $n; }; $byVal(); $byVal();
        return [$x, $acc, $n];
    }
}
echo json_encode((new T)->run()), "\n";
$total = 0; array_map(function ($v) use (&$total) { $total += $v; }, [1, 2, 3]); echo $total, "\n";
