<?php

// An assignment as a prefix operator's operand. php applies the operator to
// the assignment, not the other way round: the target stays a variable, and a
// write to the operator's result is what php refuses.

class Thing
{
    public $v = 1;
    public $cache = null;
}

$ch = 'abc';
$id = (int) $ch = '42z';
var_dump($id, $ch);

$n = 1;
$neg = -$n = 3;
var_dump($n, $neg);

$s = 'a';
$at = @$s = 'b';
var_dump($s, $at);

$arr = [1, 2];
$cast = (array) $arr = [];
var_dump($arr, $cast);

$a = null;
$copy = clone $a = new Thing();
var_dump($a === $copy, get_class($a));

$cached = null;
$held = clone $cached ??= new Thing();
var_dump($held === $cached, $cached instanceof Thing);

function flag($x)
{
    return $x > 0;
}

$ok = 0;
$not = !$ok = flag(5);
var_dump($ok, $not);

// The assignment's value is the operand's, so a compound one reads as php's.
$total = 0;
$sum = (int) $total += 7;
var_dump($total, $sum);
