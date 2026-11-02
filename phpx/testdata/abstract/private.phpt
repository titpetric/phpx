name: an abstract method cannot be private
description: >
  php refuses a declaration without a body that is also private, and reports
  that before it counts the class abstract. A trait may declare one, because
  the class that uses the trait supplies the body. The expected section is
  what `php` printed, which is nothing. The expected error names the message
  php printed. The bytecode engine declines a program that declares an
  abstract method, so its column skips and the whole program runs on the
  interpreter.
error: 'Abstract function A::f() cannot be declared private'
runner:
  php: false
---
<?php

abstract class A
{
	abstract private function f();
}

echo "unreachable\n";
---
Internal Server Error
