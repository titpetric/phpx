name: class_alias names an interface
description: >
  `class_alias('Real', 'Legacy_Real')` over an interface lets a class implement
  the old name and `instanceof` answer for both. mustache's compat.php aliases
  every interface it renamed. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
interface Real { const K = 3; function go(); }
class_alias('Real', 'Legacy_Real');
class Impl implements Legacy_Real { function go() { return 'ok'; } }
$i = new Impl;
var_dump($i instanceof Real, $i instanceof Legacy_Real, interface_exists('Legacy_Real'), Legacy_Real::K);
var_dump(class_alias('Real', 'Legacy_Real'));
---
bool(true)
bool(true)
bool(true)
int(3)
bool(false)
