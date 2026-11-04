<?php
echo json_encode(array_map('strtoupper', ['a' => 'x', 'b' => 'y'])), "\n";
echo json_encode(array_map(fn($v) => $v * 2, [3 => 1, 7 => 2])), "\n";
echo json_encode(array_map(fn($a, $b) => $a . $b, ['a', 'b'], ['c'])), "\n";
echo json_encode(array_map(null, [1, 2], ['x', 'y'])), "\n";
echo json_encode(array_map(fn($x) => $x, [])), "\n";
