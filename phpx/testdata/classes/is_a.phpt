name: is_a answers instanceof and the class-name form
description: >
  php's is_a() takes an object, or a class name string when the third argument
  allows it. phpx left it unregistered while is_subclass_of() existed, so
  framework code guarding an object with it raised an undefined function. The
  expected section is what `php` printed.
---
<?php
interface I {}
class A implements I {}
class B extends A {}
$b = new B;
$s = new stdClass;
var_dump(is_a($b, 'A'), is_a($b, 'B'), is_a($b, 'I'), is_a($b, 'a'), is_a($b, '\\A'));
var_dump(is_a($s, 'stdClass'), is_a($b, 'X'), is_a($b, 'stdClass'));
var_dump(is_a('B', 'A'), is_a('B', 'A', true), is_a('A', 'B', true), is_a('Nope', 'A', true));
var_dump(is_a(1, 'A'), is_a(null, 'A'), is_a([], 'A'));
---
bool(true)
bool(true)
bool(true)
bool(true)
bool(true)
bool(true)
bool(false)
bool(false)
bool(false)
bool(true)
bool(false)
bool(false)
bool(false)
bool(false)
bool(false)
