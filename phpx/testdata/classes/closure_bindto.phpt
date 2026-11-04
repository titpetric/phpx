---
name: Closure::bindTo and Closure::call rebind $this and the class scope
description: >
  bindTo() takes the same arguments as Closure::bind and call() runs the
  closure once against another object, in that object's class scope. The
  expected section is what `php` printed.
---
<?php
class A { private $v = 7; protected static $s = 'st'; function get() { return function () { return $this->v; }; } }
class B { private $v = 100; }
$f = function ($x = 1) { return $this->v + $x; };
$a = new A;
$g = $f->bindTo($a, A::class);
echo $g(), "\n";
echo $g(10), "\n";
echo $f->call(new B, 5), "\n";
$h = (new A)->get();
$k = $h->bindTo(new B, B::class);
echo $k(), "\n";
$bare = $f->bindTo($a);
try { $bare(); } catch (Error $e) { echo get_class($e), ": ", $e->getMessage(), "\n"; }
$sf = static function () { return static::$s; };
echo $sf->bindTo(null, A::class)(), "\n";
try { $f->call(5); } catch (TypeError $e) { echo "TypeError\n"; }
---
8
17
105
100
Error: Cannot access private property A::$v
st
TypeError
