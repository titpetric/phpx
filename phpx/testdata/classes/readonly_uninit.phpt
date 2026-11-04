---
name: a readonly property read before its first write raises
description: >
  Reading an uninitialized readonly property raises Error, isset() answers
  false, and the read works once the constructor or a method has written it.
  The expected section is what `php` printed.
---
<?php
class A {
    public readonly int $x;
    public function init() { $this->x = 4; }
    public function peek() { return $this->x; }
}
class B extends A {}
function t(callable $f) {
    try { $f(); } catch (Error $e) { echo get_class($e), ": ", $e->getMessage(), "\n"; }
}
$a = new B;
t(function () use ($a) { echo $a->x; });
t(function () use ($a) { echo $a->peek(); });
echo isset($a->x) ? "set" : "unset", "\n";
$a->init();
echo $a->x, $a->peek(), "\n";
---
Error: Typed property A::$x must not be accessed before initialization
Error: Typed property A::$x must not be accessed before initialization
unset
44
