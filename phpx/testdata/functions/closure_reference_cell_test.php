<?php
function make() {
    $promise = null;
    $fn = function () use (&$promise) { return $promise === null ? 'null' : $promise; };
    $promise = 'assigned after';
    return $fn;
}
echo make()(), "\n";
function counter() { $n = 0; return [function () use (&$n) { return ++$n; }, function () use (&$n) { return $n; }]; }
[$inc, $get] = counter();
$inc(); $inc();
echo $get(), "\n";
$total = 0;
array_map(function ($x) use (&$total) { $total += $x; }, [1, 2, 3]);
echo $total, "\n";
$self = null;
$self = function ($n) use (&$self) { return $n <= 1 ? 1 : $n * $self($n - 1); };
echo $self(5), "\n";
$byValue = 1;
$f = function () use ($byValue) { return $byValue; };
$byValue = 2;
echo $f(), "\n";
class K { public $log = []; function run() { $l = &$this->log; $add = function ($v) use (&$l) { $l[] = $v; }; $add(1); $add(2); return count($this->log); } }
echo (new K)->run(), "\n";
