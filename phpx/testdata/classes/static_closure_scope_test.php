<?php
class A { const X = 'x'; public static function make() { return static fn() => self::X . static::class . self::helper(); }
  private static function helper() { return '!'; } }
class B extends A {}
echo (A::make())(), (B::make())(), "\n";
