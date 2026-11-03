name: __CLASS__, __TRAIT__ and get_class() inside a trait method
description: >
  php compiles __CLASS__ in a trait body to the class that used the trait and
  __TRAIT__ to the trait, and get_class() with no argument names the class of
  the running method. In an inherited ordinary method both name the class that
  wrote the body. The expected section is what `php` printed.
---
<?php

trait Names
{
	public function viaClass() { return __CLASS__; }
	public function viaTrait() { return __TRAIT__; }
	public function viaGetClass() { return get_class(); }
	public static function viaStatic() { return __CLASS__ . "/" . __TRAIT__; }
	public function method() { return __METHOD__; }
}

class First { use Names; }
class Second { use Names; }
class Plain { public function name() { return __CLASS__ . "/" . get_class(); } }
class Child extends Plain {}

$a = new First();
$b = new Second();
echo $a->viaClass(), " ", $a->viaTrait(), " ", $a->viaGetClass(), "\n";
echo $b->viaClass(), " ", $b->viaTrait(), " ", $b->viaGetClass(), "\n";
echo First::viaStatic(), " ", Second::viaStatic(), "\n";
echo $a->method(), "\n";
echo (new Child())->name(), "\n";
echo var_export(__TRAIT__, true), "\n";
---
First Names First
Second Names Second
First/Names Second/Names
Names::method
Plain/Plain
''
