<?php
function count3() { for ($i = 1; $i <= 3; $i++) { yield $i; } return 'done'; }
foreach (count3() as $k => $v) { echo "$k=$v "; }
echo "\n";
function kv() { yield 'a' => 1; yield 'b' => 2; }
echo json_encode(iterator_to_array(kv())), "\n";
function inner() { yield 1; yield 2; return 'r'; }
function outer() { $r = yield from inner(); yield 3; yield from [10, 20]; return $r; }
$g = outer(); foreach ($g as $v) { echo $v, ' '; } echo $g->getReturn(), "\n";
function echoer() { while (true) { $x = yield; if ($x === null) { return; } echo "got $x\n"; } }
$e = echoer(); $e->current(); $e->send('A'); $e->send('B');
class C { private $items = [1, 2]; public function each() { foreach ($this->items as $i) { yield $i * 2; } } }
foreach ((new C)->each() as $v) { echo $v, ','; } echo "\n";
$gen = (function () { yield 1; yield 2; })(); echo $gen->current(); $gen->next(); echo $gen->current(), var_export($gen->valid(), true), "\n";
function thrower() { yield 1; throw new RuntimeException('boom'); }
try { foreach (thrower() as $v) { echo $v; } } catch (RuntimeException $e) { echo ' caught ', $e->getMessage(), "\n"; }
var_dump(count3() instanceof Traversable, count3() instanceof Generator);
$rows = [[1, 2], [3]]; $k = 10;
$g = (function () use ($rows, $k) { foreach ($rows as $r) { yield array_sum($r) + $k; } })();
echo json_encode(iterator_to_array($g)), "\n";
