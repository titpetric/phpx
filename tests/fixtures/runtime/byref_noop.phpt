name: reference markers alias a variable
phpx: true
description: >
  phpx reverses the won't-implement in docs/design.md for `$a = &$b`: the two
  names share one slot, as in php. Returning a reference from `function &f()`
  stays a value copy, so that half is not in this fixture.
runner:
  flatstack: false
---
<?php

$b = 2;
$a = &$b;
$b = 9;
echo $a, "\n";
$a = 4;
echo $b, "\n";
---
9
4
