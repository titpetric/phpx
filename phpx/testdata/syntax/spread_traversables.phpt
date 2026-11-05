name: spread unpacks generators, iterators and aggregates
description: >
  A call and an array literal unpack any Traversable; a string key names the parameter. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
function gen() { yield 1; yield 2; yield 'k' => 3; }
function f(...$a) { return json_encode($a); }
class I { public static function g() { yield 'a'; yield 'b'; } }
class Agg implements IteratorAggregate { function getIterator(): Iterator { return new ArrayIterator(['x', 'y']); } }
echo f(...gen()), f(...I::g()), f(...[1, 2]), f(...new Agg), f(...new ArrayIterator([5, 6])), "\n";
echo json_encode([...gen(), ...I::g(), ...[7], ...new Agg, ...new ArrayObject(['m' => 1, 2])]), "\n";
function named(int $a, int $b) { return "$a-$b"; }
echo named(...['b' => 2, 'a' => 1]), named(...(function () { yield 'a' => 5; yield 'b' => 6; })()), "\n";
---
{"0":1,"1":2,"k":3}["a","b"][1,2]["x","y"][5,6]
{"0":1,"1":2,"k":3,"2":"a","3":"b","4":7,"5":"x","6":"y","m":1,"7":2}
1-25-6
