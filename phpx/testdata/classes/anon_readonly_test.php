<?php
interface I { function r(); }
$f = fn() => 1;
$o = new readonly class($f, 5) implements I {
    public function __construct(private Closure $c, public int $n) {}
    function r() { return ($this->c)() + $this->n; }
};
echo $o->r(), "\n";
try { $o->n = 9; } catch (Error $e) { echo str_contains($e->getMessage(), "Cannot modify readonly property") ? "readonly\n" : "other\n"; }
