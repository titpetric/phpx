name: a child may not change a property's type
description: >
  A property's type is invariant across a redeclaration, and php names the
  type the chain carries. The expected section is what php printed, which is
  nothing. The expected error names the message php printed. The bytecode
  engine declines a program that declares `extends`, so its column skips.
error: 'Type of C::$a must be array (as in class P)'
runner:
  php: false
---
<?php

class P
{
	public array $a = [];
}

class C extends P
{
	public $a;
}

echo "unreachable\n";
---
Internal Server Error
