name: a constant expression may name self
description: >
  php resolves `self::` inside a class constant's expression and a static
  property's default to the declaring class, and an interface constant does the
  same. phpx raised `class constant self::A: unknown class` because the
  expression ran in the scope that asked for the constant. The expected section
  is what `php` printed.
---
<?php
interface I { const X = 2; const Y = self::X * 3; }
class K implements I {
    const A = 1;
    const B = self::A + 1;
    const C = K::B * 2;
    public static $s = self::A;
}
echo K::B, " ", K::C, " ", K::$s, " ", I::Y, " ", K::Y, "\n";
---
2 4 1 6 6
