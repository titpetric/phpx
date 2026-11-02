name: a parent's body cannot become abstract in the child
description: >
  A class may not redeclare a method its parent implements as a declaration
  without a body: php calls that making a non abstract method abstract. The
  expected section is what `php` printed, which is nothing. The expected error
  names the message php printed. The bytecode engine declines a program that
  declares an abstract method, so its column skips and the whole program runs
  on the interpreter.
error: 'Cannot make non abstract method P::f() abstract in class C'
runner:
  php: false
---
<?php

class P
{
	public function f()
	{
		return "p";
	}
}

abstract class C extends P
{
	abstract public function f();
}

echo "unreachable\n";
---
Internal Server Error
