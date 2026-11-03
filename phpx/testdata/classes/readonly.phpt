name: readonly properties are set once, from inside the declaring class
description: >
  A readonly property, declared or promoted, or any property of a readonly
  class, refuses a second write, an unset and an indexed write, and refuses
  its first write from outside the declaring class. The expected section is
  what `php` printed.
---
<?php
class P {
    public readonly int $x;
    protected readonly array $list;
    public function __construct(int $x, public readonly string $n = "a") { $this->x = $x; $this->list = [1]; }
    public function set($v) { $this->x = $v; }
    public function push() { $this->list[] = 2; }
    public function count() { return count($this->list); }
}
class Q extends P {
    public function reset() { $this->x = 0; }
}
class Lazy {
    public readonly int $v;
    public function init($v) { $this->v = $v; }
}
final readonly class Pt {
    public int $a;
    public function __construct(public int $x, public int $y) { $this->a = $x + $y; }
}
function t(callable $f) {
    try { $f(); } catch (Error $e) { echo get_class($e), ": ", $e->getMessage(), "\n"; }
}
$p = new P(3);
echo $p->x, $p->n, "\n";
t(function () use ($p) { $p->x = 5; });
t(function () use ($p) { $p->set(9); });
t(function () use ($p) { $p->n .= "z"; });
t(function () use ($p) { unset($p->n); });
t(function () use ($p) { $p->push(); });
echo $p->count(), "\n";
$q = new Q(4);
t(function () use ($q) { $q->reset(); });
echo $q->x, "\n";
$l = new Lazy;
t(function () use ($l) { $l->v = 1; });
$l->init(7);
t(function () use ($l) { $l->init(8); });
echo $l->v, "\n";
$pt = new Pt(1, 2);
echo $pt->x, $pt->y, $pt->a, "\n";
t(function () use ($pt) { $pt->y = 5; });
echo isset($l->v) ? "set" : "unset", "\n";
$m = new Lazy;
echo isset($m->v) ? "set" : "unset", "\n";
print_r($p);
---
3a
Error: Cannot modify readonly property P::$x
Error: Cannot modify readonly property P::$x
Error: Cannot modify readonly property P::$n
Error: Cannot unset readonly property P::$n
Error: Cannot indirectly modify readonly property P::$list
1
Error: Cannot modify readonly property P::$x
4
Error: Cannot modify protected(set) readonly property Lazy::$v from global scope
Error: Cannot modify readonly property Lazy::$v
7
123
Error: Cannot modify readonly property Pt::$y
set
unset
P Object
(
    [x] => 3
    [list:protected] => Array
        (
            [0] => 1
        )

    [n] => a
)
