name: an inherited static property shares its storage with the declaring class
description: >
  php keeps one storage for a static property per declaration. A child that
  inherits it reads and writes the parent's value, and a child that redeclares
  it gets its own. The expected section is what `php` printed.
---
<?php
class A { static $n = 0; static $r = 'a'; static function inc() { static::$n++; } }
class B extends A { static $r = 'b'; }
class C extends B {}
B::inc(); A::$n++; C::$n += 5; echo A::$n, B::$n, C::$n, "\n";
C::$r = 'c'; echo A::$r, B::$r, C::$r, "\n";
class Reg { protected static $items = []; static function add($x) { static::$items[] = $x; } static function all() { return static::$items; } }
class Sub extends Reg {}
Sub::add(1); Reg::add(2); echo count(Sub::all()), count(Reg::all()), "\n";
---
777
acc
22
