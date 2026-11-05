name: settype converts a variable in place
description: >
  `settype($v, "integer")` rewrites the variable and answers true; an unknown
  type answers false. Expected output is php 8.4's.
runner:
  flatstack: false
---
<?php
function f() { $a = "12abc"; settype($a, "integer"); var_dump($a); $b = 0; settype($b, "bool"); var_dump($b); $c = 3; settype($c, "array"); var_dump($c); $d = 1.5; settype($d, "string"); var_dump($d); var_dump(settype($d, "null"), $d); }
f();
$g = "7"; settype($g, "float"); var_dump($g);
---
int(12)
bool(false)
array(1) {
  [0]=>
  int(3)
}
string(3) "1.5"
bool(true)
NULL
float(7)
