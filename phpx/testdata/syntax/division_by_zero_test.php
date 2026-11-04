<?php
function t(callable $f) { try { var_dump($f()); } catch (DivisionByZeroError $e) { echo get_class($e), ': ', $e->getMessage(), "\n"; } }
$z = 0;
t(fn() => 1 / $z);
t(fn() => 1 % $z);
t(fn() => 1.5 / 0.0);
t(fn() => 5 % 0);
t(fn() => 7 / "0");
t(fn() => intdiv(1, $z));
t(fn() => 10 % 3);
t(fn() => 1 % 0.5);
$a = 4;
try { $a /= $z; } catch (DivisionByZeroError $e) { echo $e->getMessage(), "\n"; }
try { $a %= $z; } catch (DivisionByZeroError $e) { echo $e->getMessage(), "\n"; }
echo $a, "\n";
