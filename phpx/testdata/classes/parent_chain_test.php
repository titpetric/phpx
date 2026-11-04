<?php
class A { public static function make($x) { return 'A' . $x; } public function who() { return 'A'; } public static function create() { return new static; } }
class B extends A {
    public static function make($x) { return 'B' . parent::make($x); }
    public function who() { return 'B' . parent::who(); }
}
class C extends B {
    public static function make($x) { return 'C' . parent::make($x); }
    public function who() { return 'C' . parent::who(); }
}
echo C::make(1), (new C)->who(), get_class(C::create()), "\n";
abstract class Base { public static function build(...$a) { return static::class . count($a); } }
final class Impl extends Base { public static function build(...$a) { return 'I' . parent::build(...$a); } }
echo Impl::build(1, 2, 3), "\n";
class A2 { public function who() { return 'A'; } protected function tag() { return 'a'; } public function both() { return $this->tag(); } }
class B2 extends A2 { public function who() { return 'B' . parent::who(); } protected function tag() { return 'b' . parent::tag(); } }
class C2 extends B2 { public function who() { return 'C' . parent::who(); } protected function tag() { return 'c' . parent::tag(); } }
$c = new C2;
echo $c->who(), $c->both(), "\n";
class K1 { public function __construct(public string $log = '') { $this->log .= 'k1;'; } }
class K2 extends K1 { public function __construct() { parent::__construct(); $this->log .= 'k2;'; } }
class K3 extends K2 { public function __construct() { parent::__construct(); $this->log .= 'k3;'; } }
echo (new K3)->log, "\n";
