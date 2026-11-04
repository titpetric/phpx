<?php
namespace App;
class Base { const KIND = 'base'; public static $count = 3; public static function make($n = 1) { return static::class . $n; } public function who() { return 'who'; } }
class Child extends Base { const KIND = 'child'; public static function twice(int $a, int $b) { return $a * $b; } }
$cls = Child::class;
echo $cls::KIND, "\n";
echo $cls::make(2), "\n";
echo $cls::twice(3, 4), "\n";
$m = 'twice';
echo $cls::$m(5, 6), "\n";
echo $cls::$count, "\n";
$obj = new Child;
echo $obj::KIND, "\n";
echo $obj::make(), "\n";
$name = 'make';
echo $obj::$name(9), "\n";
$args = [2, 7];
echo $cls::twice(...$args), "\n";
foreach ([Base::class, Child::class] as $c) { echo $c::KIND, ' '; }
echo "\n";
