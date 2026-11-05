name: a reference to an array element writes back
description: >
  '$r = &$arr[0]; $r = 9' changes the element, and '&$arr[]' appends one. The other direction, a write to the element read through $r, is not kept. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
$arr = [1, 2, 3];
$r = &$arr[0]; $r = 9; echo $arr[0], "\n";
$c = &$arr[]; $c = 'x'; echo count($arr), $arr[3], "\n";
$m = ['k' => ['n' => 1]]; $p = &$m['k']['n']; $p = 5; echo $m['k']['n'], "\n";
---
9
4x
5
