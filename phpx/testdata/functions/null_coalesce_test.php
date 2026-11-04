<?php
$a = null;
echo $a ?? "null", "\n";
echo $undefined ?? "undef", "\n";
$f = false; $z = 0; $e = "";
var_dump($f ?? "x", $z ?? "x", $e ?? "x");
$m = ['k' => null, 'j' => 5, 'n' => ['d' => 7]];
echo $m['k'] ?? "nullk", $m['j'] ?? "none", $m['zz'] ?? "missing", "\n";
echo $m['n']['d'] ?? "no", $m['n']['q']['r'] ?? "deep", "\n";
echo $nothing ?? $m['k'] ?? "chain", "\n";
class P { public $x = null; public $y = 3; }
$o = new P;
echo $o->x ?? "px", $o->y ?? "py", $o->w ?? "pw", "\n";
$a ??= 5;
$a ??= 9;
echo $a, "\n";
$m['k'] ??= "filled";
$m['new'] ??= "created";
$m['j'] ??= "kept";
echo $m['k'], $m['new'], $m['j'], "\n";
$o->x ??= "set";
echo $o->x, "\n";
echo ($a ?? 0) + 1, "\n";
echo $undefined2 ?? 1 ? "t" : "f", "\n";
