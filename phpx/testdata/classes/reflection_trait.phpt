name: ReflectionClass reads a trait
description: >
  isTrait, the trait's methods and short name, and the trait names a class uses. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
namespace N;
trait T1 { public $p = 1; function a() { return 1; } static function s() {} }
$r = new \ReflectionClass('N\T1'); var_dump($r->isTrait(), $r->getName(), $r->isInterface(), $r->hasMethod('a'), count($r->getMethods()), $r->getShortName());
class C { use T1; } $c = new \ReflectionClass(C::class); var_dump($c->isTrait(), $c->getTraitNames(), array_keys($c->getTraits()));
---
bool(true)
string(4) "N\T1"
bool(false)
bool(true)
int(2)
string(2) "T1"
bool(false)
array(1) {
  [0]=>
  string(4) "N\T1"
}
array(1) {
  [0]=>
  string(4) "N\T1"
}
