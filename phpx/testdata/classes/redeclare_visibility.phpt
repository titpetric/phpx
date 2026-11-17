name: a child may not weaken a property's visibility
description: >
  php refuses a redeclaration that narrows the visibility the chain already
  carries for the name, and names the class the child must not go below. The
  expected section is what php printed, which is nothing. The expected error
  names the message php printed. The bytecode engine declines a program that
  declares `extends`, so its column skips.
error: 'Access level to C::$a must be protected (as in class P) or weaker'
runner:
  php: false
---
<?php

class P
{
	protected $a;
}

class C extends P
{
	private $a;
}

echo "unreachable\n";
---
Internal Server Error
