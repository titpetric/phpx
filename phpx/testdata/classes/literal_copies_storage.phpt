name: an array literal copies the arrays it reads from storage
description: >
  `[$this->stack]` holds a value copy, so a later write to the property leaves
  the snapshot alone. A generated parser snapshots its stacks this way and
  restores them on error recovery. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
class P { public $s = [1]; public $stack = [1,2];
 function t() {
  $snap = [$this->s, $this->stack];
  $this->s[] = 2; $this->stack[5] = 9;
  echo count($snap[0]), count($snap[1]), "\n";
  [$x, $y] = $snap; $this->s = $x; 
  $this->s[] = 3; echo count($x), count($snap[0]), "\n";
  $q = $this->stack; $this->stack[] = 1; echo count($q), "\n";
  $r = [$this->stack]; $this->stack[] = 1; echo count($r[0]), "\n";
 }}
(new P)->t();
---
12
11
3
4
