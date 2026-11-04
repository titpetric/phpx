<?php
echo serialize(null), serialize(true), serialize(-5), serialize(1.5), serialize("héllo"), "\n";
echo serialize([1, 'a' => [true, null], 2 => 'x']), "\n";
$d = unserialize(serialize(['k' => [1, 2.5, 'v', false, null]]));
echo json_encode($d), "\n";
class P { public $a = 1; public $b = 'two'; }
echo serialize(new P), "\n";
var_dump(unserialize('garbage'));
