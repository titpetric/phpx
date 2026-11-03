name: method visibility is enforced at a call
description: >
  A private or protected method, static method or constructor called from outside
  its scope raises php's Error. The expected section is what `php` printed.
---
<?php
class A {
    public function pub() { return "pub"; }
    protected function prot() { return "prot"; }
    private function priv() { return "priv"; }
    public function viaSelf() { return $this->prot() . $this->priv(); }
    public function viaClosure() { $f = function () { return $this->priv(); }; return $f(); }
    public function other(A $o) { return $o->priv(); }
    protected static function sprot() { return "sprot"; }
    private static function spriv() { return "spriv"; }
    public static function make() { return static::sprot() . self::spriv(); }
}
class B extends A {
    public function callProt() { return $this->prot(); }
    public function callPriv() { return $this->priv(); }
    public function callSProt() { return parent::sprot() . static::sprot(); }
    public function onOther(A $o) { return $o->prot(); }
}
class C { private function __construct() {} public static function get() { return new C; } }
class D { protected function __construct() {} }
function t($f) { try { echo $f(), "\n"; } catch (Error $e) { echo "Error: ", $e->getMessage(), "\n"; } }
$a = new A; $b = new B;
t(function () use ($a, $b) { return $a->pub(); });
t(function () use ($a, $b) { return $a->prot(); });
t(function () use ($a, $b) { return $a->priv(); });
t(function () use ($a, $b) { return $a->viaSelf(); });
t(function () use ($a, $b) { return $a->viaClosure(); });
t(function () use ($a, $b) { return $a->other(new A); });
t(function () use ($a, $b) { return $b->callProt(); });
t(function () use ($a, $b) { return $b->callPriv(); });
t(function () use ($a, $b) { return $b->prot(); });
t(function () use ($a, $b) { return $b->priv(); });
t(function () use ($a, $b) { return $b->viaSelf(); });
t(function () use ($a, $b) { return $b->onOther($a); });
t(function () use ($a, $b) { return A::make(); });
t(function () use ($a, $b) { return B::make(); });
t(function () use ($a, $b) { return A::sprot(); });
t(function () use ($a, $b) { return A::spriv(); });
t(function () use ($a, $b) { return $b->callSProt(); });
t(function () use ($a, $b) { return C::get() instanceof C ? "ok" : "no"; });
t(function () use ($a, $b) { return new C; });
t(function () use ($a, $b) { return new D; });

---
pub
Error: Call to protected method A::prot() from global scope
Error: Call to private method A::priv() from global scope
protpriv
priv
priv
prot
Error: Call to private method A::priv() from scope B
Error: Call to protected method A::prot() from global scope
Error: Call to private method A::priv() from global scope
protpriv
prot
sprotspriv
sprotspriv
Error: Call to protected method A::sprot() from global scope
Error: Call to private method A::spriv() from global scope
sprotsprot
ok
Error: Call to private C::__construct() from global scope
Error: Call to protected D::__construct() from global scope
