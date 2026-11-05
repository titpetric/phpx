name: new takes a parenthesized class expression
description: >
  `new ($this->nodeClass)($a, $b)` is how twig builds a node from a class name
  held in a property. Expected output is php 8.4's.
---
<?php
class Pair { function __construct(public $a, public $b) {} }
class Holder {
    public $cls = 'Pair';
    function make() { return new ($this->cls)(1, 2); }
}
$p = (new Holder)->make();
echo get_class($p), $p->a, $p->b, "\n";
$name = 'Pair';
$q = new ($name)('x', 'y');
echo $q->a, $q->b, "\n";
$r = new ('P' . 'air')(3, 4);
echo $r->a + $r->b, "\n";
---
Pair12
xy
7
