name: nested and non-array destructuring
description: >
  A pattern inside a pattern destructures its element, and a value that is no
  array gives every target null without an error. Expected output is php 8.4's.
---
<?php
[[$a, $b], $c] = [[1, 2], 3];
echo "$a$b$c\n";
foreach ([[1, [2, 3]], [4, [5, 6]]] as [$x, [$y, $z]]) echo "$x$y$z,";
echo "\n";
list($p, list($q, $r)) = [7, [8, 9]]; echo "$p$q$r\n";
[, [, $k]] = [0, [1, 2]]; echo $k, "\n";
[[$u, $v], $w] = [null, 3]; var_dump($u, $v, $w);
[$m, $n] = null; var_dump($m);
[$m2] = 5; var_dump($m2);
---
123
123,456,
789
2
NULL
NULL
int(3)
NULL
NULL
