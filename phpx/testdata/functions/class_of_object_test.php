<?php
class A { function who() { return $this::class; } }
class B extends A { public $o; }
function make() { return new B; }
$o = new B;
echo $o::class, "\n";
echo make()::class, "\n";
echo $o->who(), "\n";
$o->o = new A;
echo $o->o::class, "\n";
echo A::class, "\n";
$list = [new A, new B];
foreach ($list as $x) { echo $x::class, ","; }
echo "\n";
echo ($o::class === get_class($o)) ? "same" : "diff", "\n";
