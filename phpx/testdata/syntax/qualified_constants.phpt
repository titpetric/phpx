name: fully qualified true, false and null
description: >
  \true, \false and \null are the constants of the same name, in the global namespace and inside one, and a guard on \false skips its block. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
namespace App;
var_dump(\true, \false, \null, \TRUE, \False);
if (\false) { echo "never\n"; }
$a = \null ?? 'fallback';
echo $a, "\n";
function f($x = \null, $y = \false) { var_dump($x, $y); }
f();
echo \strlen('abc'), "\n";
---
bool(true)
bool(false)
NULL
bool(true)
bool(false)
fallback
NULL
bool(false)
3
