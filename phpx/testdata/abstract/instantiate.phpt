name: an abstract class cannot be instantiated
description: >
  php refuses `new` on a class that declares the abstract modifier, whatever
  its methods look like. The expected section is what `php` printed, which is
  nothing: the fatal fires before the script reaches an echo. The expected
  error names the message php printed. The bytecode engine declines a program
  that declares an abstract class, so its column skips and the whole program
  runs on the interpreter.
error: 'Cannot instantiate abstract class A'
runner:
  php: false
---
<?php

abstract class A
{
	public function f()
	{
		return 1;
	}
}

new A();

echo "unreachable\n";
---
Internal Server Error
