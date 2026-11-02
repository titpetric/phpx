name: a class's own abstract declaration is the one php names
description: >
  A class that declares an abstract method itself and does not declare the
  class abstract hears about that method before any trait binds or a parent's
  members arrive, which is php's compile-time order. The expected section is
  what `php` printed, which is nothing. The expected error names the message
  php printed. The bytecode engine declines a program that declares an
  abstract method, so its column skips and the whole program runs on the
  interpreter.
error: 'Class C contains 1 abstract method and must therefore be declared abstract or implement the remaining methods (C::z)'
runner:
  php: false
---
<?php

class C
{
	abstract public function z();
}

echo "unreachable\n";
---
Internal Server Error
