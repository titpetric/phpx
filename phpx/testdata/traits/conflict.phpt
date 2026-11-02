name: two traits declaring one method are a collision
runner:
  php: false
description: >
  php composes a class from every trait it names, and two traits declaring one
  method collide whatever the bodies say. The fatal names the later trait, the
  class it could not be applied to, and the earlier trait. The expected section
  is what `php` printed, which is nothing: the fatal fires at class declaration
  and the script never reaches an echo. The expected error names the message
  php printed. The bytecode engine declines a program that declares or uses a
  trait, so its column skips and the whole program runs on the interpreter.
error: 'Trait method B::foo has not been applied as C::foo, because of collision with A::foo'
---
<?php

trait A
{
	public function foo()
	{
		return "A";
	}
}

trait B
{
	public function foo()
	{
		return "B";
	}
}

class C
{
	use A, B;
}

echo "unreachable\n";
---
Internal Server Error
