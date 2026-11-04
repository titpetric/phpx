<?php
eval('class Gen { public function hi() { return "hi"; } }');
echo (new Gen)->hi(), "\n";
echo eval('return 6 * 7;'), "\n";
eval('function made($x) { return $x + 1; }');
echo made(1), "\n";
$n = 'Dyn' . 1;
eval("namespace Ev; final class $n extends \\Gen { }");
echo get_class(new Ev\Dyn1), ' ', (new Ev\Dyn1)->hi(), "\n";
try { eval('class { '); } catch (ParseError $e) { echo "parse error\n"; }
