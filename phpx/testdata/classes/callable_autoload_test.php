<?php
spl_autoload_register(function ($c) { if ($c === 'Late') { eval('class Late { public static function twice($x) { return $x * 2; } public function inst($x) { return $x; } }'); } });
var_dump(is_callable(['Late', 'twice']), is_callable(['Late', 'nope']));
echo implode(',', array_map(['Late', 'twice'], [1, 2])), "\n";
echo call_user_func('Late::twice', 4), "\n";
