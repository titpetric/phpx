name: a Go slice a binding answers is a script array once a variable holds it
description: >
  explode() and its kin answer a Go slice upstream, which reads and iterates but
  cannot take array_pop or unset. The interpreter turns it into a script array
  where a variable or a parameter takes it, so php's numbering and mutators
  hold. The bytecode engine keeps the slice, so the fixture opts it out.
runner:
  flatstack: false
---
<?php
$e = explode(",", "x,y,z");
unset($e[1]);
echo implode(",", array_keys($e)), "\n";
echo json_encode($e), "\n";
var_dump(isset($e[2]));
$p = explode("/", "a/b/c");
echo array_pop($p), array_shift($p), count($p), "\n";
array_push($p, "z");
echo implode(",", $p), "\n";
$chars = str_split("abc");
array_unshift($chars, "_");
echo implode("", $chars), "\n";
$parts = preg_split('/[\s,]+/', "a, b  c");
sort($parts);
echo implode("|", $parts), "\n";
$last = array_pop(explode(".", "file.tar.gz"));
echo $last, "\n";
?>
---
0,2
{"0":"x","2":"z"}
bool(true)
ca1
b,z
_abc
a|b|c
gz
