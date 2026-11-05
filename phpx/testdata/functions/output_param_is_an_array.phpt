name: an output parameter holds an array a script can change
description: >
  preg_match_all and preg_match leave an array in the variable that array_shift, array_pop and array_unshift change in place, including the while ([...] = array_shift($parts)) loop. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
preg_match_all('/(\d+)\s*(\w+)/', '1 day 2 hours', $parts, PREG_SET_ORDER);
var_dump(array_shift($parts)[1], count($parts));
$n = 0; while ([$a, $b, $c] = array_shift($parts)) { $n++; if ($n > 4) break; } echo "n=$n count=", count($parts), "\n";
preg_match('/(\d+)/', 'x 12', $m); array_shift($m); var_dump(count($m));
$s = str_split('abc'); array_shift($s); var_dump(count($s)); $e = explode(',', 'a,b,c'); array_pop($e); array_unshift($e, 'z'); var_dump($e);
$k = array_keys(['a' => 1, 'b' => 2]); array_shift($k); var_dump($k);
---
string(1) "1"
int(1)
n=1 count=0
int(1)
int(2)
array(3) {
  [0]=>
  string(1) "z"
  [1]=>
  string(1) "a"
  [2]=>
  string(1) "b"
}
array(1) {
  [0]=>
  string(1) "b"
}
