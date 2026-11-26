name: an array literal holds the variable a reference element names
description: >
  `[1, 2, &$x]` stores the cell, so a write to $x reads back through the
  element and a write to the element reads back through $x. A spread of such
  an array carries the reference into a php function's by-reference parameter
  and into a builtin's output parameter, which is how nette's Strings::match
  fills $m through `preg_match(...$args)`. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
$a = [1, 2, &$x, 4];
$x = 'changed';
echo $a[2], "\n";
$a[2] = 'again';
echo $x, "\n";
$keyed = ['slot' => &$x];
echo $keyed['slot'], "\n";

function fill($prefix, $suffix, &$out) { $out = $prefix . $suffix; }
$o = '';
fill(...['a', 'b', &$o]);
echo $o, "\n";

$m = [];
preg_match(...['/(b)/', 'abc', &$m]);
echo json_encode($m), "\n";
---
changed
again
again
ab
["b","b"]
