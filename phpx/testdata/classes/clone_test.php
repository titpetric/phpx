<?php
class Inner { public $n = 1; }
class A {
    public $v = 1;
    public array $list = [1, 2];
    public Inner $inner;
    public static int $clones = 0;
    public function __construct() { $this->inner = new Inner; }
    public function __clone() { self::$clones++; $this->v = $this->v + 100; }
    public function inc() { $this->v++; return $this; }
}
$a = new A;
$b = clone $a;
$b->v = 5;
$b->list[] = 3;
$b->inner->n = 9;
echo $a->v, ' ', $b->v, ' ', count($a->list), ' ', count($b->list), ' ', $a->inner->n, "\n";
echo (clone $a)->inc()->v, ' ', $a->v, "\n";
echo A::$clones, "\n";
var_dump($a == clone $a, $a === clone $a);
$o = new stdClass; $o->x = 1;
$p = clone $o; $p->x = 2;
echo $o->x, $p->x, "\n";
final class Imm {
    public function __construct(private int $n) {}
    public function with(int $n): static { $c = clone $this; $c->n = $n; return $c; }
    public function n(): int { return $this->n; }
}
$i = new Imm(1);
echo $i->with(2)->n(), $i->n(), "\n";
