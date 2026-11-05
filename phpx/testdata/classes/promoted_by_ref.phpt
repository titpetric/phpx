name: a promoted by-reference parameter shares the caller's variable
description: >
  `__construct(private array &$names)` stores a reference, so a method that
  appends to the property appends to the caller's array. php-parser's visitors
  collect their findings this way. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
class V { public function __construct(private array &$names) {} function add($x) { $this->names[] = $x; } }
$names = []; $v = new V($names); $v->add('a'); $v->add('b'); echo count($names), "\n";
$o = new class($names) { public function __construct(private array &$n) {} function push($x) { $this->n[] = $x; } };
$o->push('c'); echo count($names), "\n";
class W { public $ref; function __construct(&$r) { $this->ref = &$r; } function set($v) { $this->ref = $v; } }
$z = 1; $w = new W($z); $w->set(5); echo $z, "\n";
---
2
3
5
