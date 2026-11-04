<?php
echo 2 <=> 1, 1 <=> 2, 1 <=> 1, "\n";
echo "a" <=> "b", "b" <=> "a", "a" <=> "a", "\n";
echo "10" <=> "9", "10" <=> "9a", "\n";
echo 1.5 <=> 1.5, 1 <=> 1.5, "\n";
var_dump([1, 2] <=> [1, 3], [1, 2, 3] <=> [1, 2], [1, 2] <=> [1, 2]);
var_dump(null <=> false, true <=> false, null <=> 0);
var_dump(NAN <=> 1, 1 <=> NAN);
$n = 0;
function next_n() { global $n; return ++$n; }
var_dump(next_n() <=> next_n());
$rows = [[3, 'c'], [1, 'a'], [2, 'b']];
usort($rows, fn($x, $y) => $x[0] <=> $y[0]);
echo implode(",", array_column($rows, 1)), "\n";
usort($rows, fn($x, $y) => [$y[1]] <=> [$x[1]]);
echo implode(",", array_column($rows, 1)), "\n";
var_dump(1 + 1 <=> 2, 1 <=> 2 + 3);
