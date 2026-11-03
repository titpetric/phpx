name: constructor property promotion declares and fills the property
description: >
  A constructor parameter with a visibility or readonly modifier is also a
  property, assigned from the argument before the body runs, with a default
  and a parent constructor chain. The expected section is what `php` printed.
---
<?php
class Point {
    public $tag = "pt";
    public function __construct(public int $x = 1, protected int $y = 2, private string $label = "p", public readonly array $meta = []) {}
    public function label() { return $this->label . ":" . $this->x . "," . $this->y; }
}
$p = new Point(5);
echo $p->label(), "\n";
echo $p->x, " ", $p->tag, " ", count($p->meta), "\n";
$q = new Point(1, 9, "q");
echo $q->label(), "\n";
var_dump($p);
class Base { public function __construct(protected string $name = "base") {} public function name() { return $this->name; } }
class Child extends Base {
    public function __construct(string $name, public int $age = 3) { parent::__construct(strtoupper($name)); }
    public function describe() { return $this->name . "/" . $this->age; }
}
$c = new Child("ann", 7);
echo $c->describe(), " ", $c->name(), "\n";
class Plain extends Base {}
echo (new Plain("kept"))->name(), "\n";
class Body {
    public function __construct(public int $n) { $this->n = $n * 2; }
}
echo (new Body(4))->n, "\n";
$o = new Point(2, 3);
$o->x = 10;
echo $o->x, "\n";
---
p:5,2
5 pt 0
q:1,9
object(Point)#1 (5) {
  ["tag"]=>
  string(2) "pt"
  ["x"]=>
  int(5)
  ["y":protected]=>
  int(2)
  ["label":"Point":private]=>
  string(1) "p"
  ["meta"]=>
  array(0) {
  }
}
ANN/7 ANN
kept
8
10
