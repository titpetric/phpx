name: trait_exists and class_uses
description: >
  trait_exists is true for a trait and false for a class, class_uses lists the traits a class names itself and nothing for its subclass. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
namespace N;
trait T1 { function a() { return 1; } } trait T2 { function b() { return 2; } }
class C { use T1, T2; } class D extends C {} class E {}
var_dump(trait_exists('N\T1'), trait_exists('N\C'), trait_exists('N\Nope'), class_exists('N\T1'));
var_dump(class_uses('N\C'), class_uses(new C), class_uses('N\D'), class_uses('N\E'), class_uses('N\Missing'));
---
bool(true)
bool(false)
bool(false)
bool(false)
array(2) {
  ["N\T1"]=>
  string(4) "N\T1"
  ["N\T2"]=>
  string(4) "N\T2"
}
array(2) {
  ["N\T1"]=>
  string(4) "N\T1"
  ["N\T2"]=>
  string(4) "N\T2"
}
array(0) {
}
array(0) {
}
bool(false)
