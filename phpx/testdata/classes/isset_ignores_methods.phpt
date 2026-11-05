name: isset on a method name is false
description: >
  isset($o->name) and empty() ask whether a property exists; a method of that name is not one. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
class A { private function priv() { return 1; } public function pub() { return 2; } public $p = 3; public $n = null; }
$a = new A;
var_dump(isset($a->priv), isset($a->pub), isset($a->p), isset($a->n), isset($a->nope), empty($a->pub), property_exists($a, 'priv'), property_exists($a, 'p'));
---
bool(false)
bool(false)
bool(true)
bool(false)
bool(false)
bool(true)
bool(false)
bool(true)
