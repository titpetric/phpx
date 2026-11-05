name: alternative control syntax and static property writes through a variable
description: >
  if/while/for/foreach/switch with a colon and an end keyword, and $cls::$prop assigned. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
$n = 3;
if ($n > 5): echo "big\n"; elseif ($n > 2): echo "mid\n"; else: echo "small\n"; endif;
while ($n > 0): echo $n--, " "; endwhile;
echo "\n";
for ($i = 0; $i < 3; $i++): echo $i; endfor;
echo "\n";
foreach ([1, 2, 3] as $k => $v): echo "$k=$v,"; endforeach;
echo "\n";
switch ($i): case 3: echo "three\n"; break; default: echo "other\n"; endswitch;
if ($i): echo "html $i\n"; endif;
class A { public static $v = 1; public static $s = 'a'; }
$d = 'A';
$d::$v = 5;
$d::$s .= 'b';
echo A::$v, A::$s, "\n";
---
mid
3 2 1 
012
0=1,1=2,2=3,
three
html 3
5ab
