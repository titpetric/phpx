name: by-reference parameters alias the caller's variable
description: >
  A parameter declared with & binds the caller's variable, and a property
  assigned by reference from it follows. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
function inc(&$x) { $x++; }
$a = 1; inc($a); echo $a, "\n";
function setp(&$p) { $p = new stdClass; $p->v = 5; }
$o = null; setp($o); echo $o->v ?? 'null', "\n";
class T {
    private $pos;
    function run(&$position) { $this->pos = &$position; $this->pos = 'moved'; }
}
$p = 'start'; $t = new T; $t->run($p); echo $p, "\n";
---
2
5
moved
