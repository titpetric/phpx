<?php
$a = ['b' => 2, 'a' => 3, 'c' => 1];
ksort($a, SORT_STRING); echo json_encode($a), "\n";
krsort($a, SORT_REGULAR); echo json_encode($a), "\n";
asort($a, SORT_NUMERIC); echo json_encode($a), "\n";
arsort($a); echo json_encode($a), "\n";
$l = [3, 1, 2]; sort($l, SORT_NUMERIC); echo json_encode($l), "\n";
rsort($l, SORT_REGULAR); echo json_encode($l), "\n";
