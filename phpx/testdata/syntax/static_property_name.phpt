name: a static property named by an expression
description: >
  `Class::${expr}` and `Class::$$name` name the static property at run time,
  where `Class::$name` names the literal `name`. The `$` is the property
  sigil, so `$$name` names the property `$name` holds. Symfony's
  DebugClassLoader writes `self::${$property}[$class]`.
  Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
class A {
  public static $arr = ['k' => 1];
  public static $list = [];
  public static $made = 0;
  public static $s = 'v';
  public static $arr2 = ['k' => 9];
  public static function read($p) { return self::${$p}; }
  public static function write($p, $v) { self::${$p} = $v; }
  public static function append($p, $v) { self::${$p}[] = $v; }
  public static function compound($p) { self::${$p} .= '!'; }
  public static function has($p, $k) { return isset(self::${$p}[$k]); }
  public static function expr($o) { return self::${'ar' . $o}; }
  public static function late($p) { return static::${$p}; }
  public static function dbl($n) { return self::$$n; }
  public static function inc($p) { self::${$p}++; }
}
class B extends A { public static $arr = ['k' => 2]; }
echo A::read('arr')['k'], "\n";
A::write('made', 5); echo A::$made, "\n";
A::append('list', 'x'); A::append('list', 'y'); echo implode(',', A::$list), "\n";
A::compound('s'); echo A::$s, "\n";
var_dump(A::has('list', '0'), A::has('list', '9'));
echo A::expr('r')['k'], "\n";
echo B::late('arr')['k'], "\n";
$cls = 'A'; echo $cls::${'arr'}['k'], "\n";
$obj = new B; echo $obj::${'arr'}['k'], "\n";
$n = 'made'; echo A::dbl($n), "\n";
A::inc('made'); echo A::$made, "\n";
echo A::$arr2['k'], "\n";
---
1
5
x,y
v!
bool(true)
bool(false)
1
2
1
2
5
6
9
