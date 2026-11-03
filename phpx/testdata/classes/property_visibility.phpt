name: property visibility is enforced at a read, a write and a listing
description: >
  A protected or private property read, written, unset or counted from outside its
  scope raises php's Error; isset and empty answer false and true; get_object_vars and
  foreach list what the calling body may see. The expected section is what `php` printed.
---
<?php
class A {
    public $pub = 1;
    protected $prot = 2;
    private $priv = 3;
    public function __construct(private int $promoted = 7) {}
    public function all() { return get_object_vars($this); }
    public function bump() { $this->priv++; $this->prot .= "x"; return $this->priv . $this->prot; }
    public function viaClosure() { $f = function () { return $this->priv; }; return $f(); }
    public function peer(A $o) { return $o->priv . $o->prot; }
    public function loop() { $out = []; foreach ($this as $k => $v) { $out[] = $k; } return implode(",", $out); }
}
class B extends A {
    public function readProt() { return $this->prot; }
    public function writeProt() { $this->prot = 20; return $this->prot; }
    public function vars() { return get_object_vars($this); }
}
function t($f) { try { echo $f(), "\n"; } catch (Error $e) { echo get_class($e), ": ", $e->getMessage(), "\n"; } }
$a = new A; $b = new B;
t(function () use ($a, $b) { return $a->pub; });
t(function () use ($a, $b) { return $a->prot; });
t(function () use ($a, $b) { return $a->priv; });
t(function () use ($a) { $a->prot = 5; return "set"; });
t(function () use ($a) { $a->priv = 5; return "set"; });
t(function () use ($a) { unset($a->prot); return "unset"; });
t(function () use ($a) { $a->prot++; return "inc"; });
t(function () use ($a) { $a->pub .= "z"; return $a->pub; });
t(function () use ($a, $b) { return implode(",", array_keys(get_object_vars($a))); });
t(function () use ($a, $b) { return implode(",", array_keys($a->all())); });
t(function () use ($a, $b) { return implode(",", array_keys($b->vars())); });
t(function () use ($a, $b) { return $a->bump(); });
t(function () use ($a, $b) { return $a->viaClosure(); });
t(function () use ($a, $b) { return $a->peer(new A); });
t(function () use ($a, $b) { return $a->loop(); });
t(function () use ($a, $b) { return $b->readProt(); });
t(function () use ($a, $b) { return $b->writeProt(); });
t(function () use ($a) { $n = 0; foreach ($a as $k => $v) { $n++; echo $k, " "; } return $n; });
var_dump(isset($a->pub), isset($a->prot), isset($a->priv), empty($a->prot), empty($a->pub));
var_dump(property_exists($a, 'prot'));
t(function () use ($a, $b) { return $a->promoted; });
---
1
Error: Cannot access protected property A::$prot
Error: Cannot access private property A::$priv
Error: Cannot access protected property A::$prot
Error: Cannot access private property A::$priv
Error: Cannot access protected property A::$prot
Error: Cannot access protected property A::$prot
1z
pub
pub,prot,priv,promoted
pub,prot
42x
4
32
pub,prot,priv,promoted
2
20
pub 1
bool(true)
bool(false)
bool(false)
bool(true)
bool(false)
bool(true)
Error: Cannot access private property A::$promoted
