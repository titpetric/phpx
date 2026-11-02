name: an abstract method has no body to call
description: >
  A child that implements an abstract method may reach the parent's version
  through parent::, and php refuses the call because there is no body behind
  it. The expected section is what `php` printed, which is nothing. The
  expected error names the message php printed. The bytecode engine declines a
  program that declares an abstract method, so its column skips and the whole
  program runs on the interpreter.
error: 'Cannot call abstract method A::f()'
runner:
  php: false
---
<?php

abstract class A
{
	abstract public function f();
}

class B extends A
{
	public function f()
	{
		return parent::f();
	}
}

$b = new B();
echo $b->f(), "\n";
---
Internal Server Error
