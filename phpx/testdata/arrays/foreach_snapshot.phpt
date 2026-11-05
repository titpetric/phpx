name: foreach walks the array as it was when the loop began
description: >
  A body that unsets from the array it walks, or calls a method that appends
  to it, leaves the iteration alone: php walks a copy. commonmark's
  extension registration loop does both. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
class E { public $u = []; function add($x) { $this->u[] = $x; } function run() {
  $n = 0;
  while (count($this->u) > 0) { foreach ($this->u as $i => $x) { $n++; if ($n < 5 && $x == 'a') { $this->add('b'); } unset($this->u[$i]); } }
  return $n; } }
$e = new E; $e->add('a'); $e->add('c'); echo $e->run(), "\n";
$arr = [1,2,3]; foreach ($arr as $k => $v) { if ($k == 0) { unset($arr[1]); $arr[] = 9; } echo $v; } echo count($arr), "\n";
---
3
1233
