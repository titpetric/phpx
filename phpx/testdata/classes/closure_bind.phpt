---
name: Closure::bind carries the class scope that visibility reads
description: >
  A closure bound with a class scope reads and writes that class's private and
  protected members, as composer's autoloader does. Bound with no scope it is
  refused. The expected section is what `php` printed.
---
<?php
class Loader {
    private $map = ['a' => 1];
    protected static $n = 5;
}
$l = new Loader;
$read = \Closure::bind(function () { return $this->map['a']; }, $l, Loader::class);
echo $read(), "\n";
$init = \Closure::bind(static function () { return Loader::$n; }, null, Loader::class);
echo $init(), "\n";
$set = \Closure::bind(function ($v) { $this->map['b'] = $v; return count($this->map); }, $l, Loader::class);
echo $set(9), "\n";
$bare = \Closure::bind(function () { return $this->map; }, $l);
try { $bare(); } catch (Error $e) { echo get_class($e), ": ", $e->getMessage(), "\n"; }
---
1
5
2
Error: Cannot access private property Loader::$map
