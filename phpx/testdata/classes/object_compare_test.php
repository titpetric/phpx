<?php
class P { public $a; function __construct($a) { $this->a = $a; } }
class Q extends P {}
$x = new P(1); $y = new P(1); $z = new P(2);
var_dump($x == $y, $x == $z, $x === $y, $x == $x, $x != $z, $x == new Q(1), $x < $z, $z > $x, [$x] == [$y], [$x] == [$z]);
$n1 = new P(new P(1)); $n2 = new P(new P(1)); $n3 = new P(new P(2));
var_dump($n1 == $n2, $n1 == $n3);
var_dump(in_array(new P(2), [$x, $y, $z]), in_array(new P(3), [$x, $y, $z]), in_array($z, [$x, $y], true));
