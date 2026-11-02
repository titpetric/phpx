name: a concrete class must implement every abstract method
description: >
  A class that does not declare the abstract modifier may not leave a method
  without a body, whether it declared the method itself or inherited it. php
  names the class, counts the methods still missing and lists them as
  Class::method in the order its table holds them: the class's own
  declarations first, then the parent's. The expected section is what `php`
  printed, which is nothing. The expected error names the message php printed.
  The bytecode engine declines a program that declares an abstract method, so
  its column skips and the whole program runs on the interpreter.
error: 'Class B contains 1 abstract method and must therefore be declared abstract or implement the remaining methods (A::g)'
runner:
  php: false
---
<?php

abstract class A
{
	abstract public function f();
	abstract public function g();
}

class B extends A
{
	public function f()
	{
		return "f";
	}
}

echo "unreachable\n";
---
Internal Server Error
