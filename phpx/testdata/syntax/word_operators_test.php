<?php
var_dump(true and false);
var_dump(false or true);
var_dump(true xor true);
var_dump(true xor false);
$x = true and false;
var_dump($x);
$y = false or true;
var_dump($y);
$z = (true xor true);
var_dump($z);
$t = 1 and 0 or 1;
var_dump($t);
function loud($v) { echo "eval $v\n"; return $v; }
var_dump(loud(1) xor loud(1));
var_dump(loud(0) and loud(1));
var_dump(loud(1) or loud(0));
if (loud(true) and !loud(false)) { echo "both\n"; }
for ($i = 0, $c = 3; $i < $c; $i++) { echo $i; }
echo "\n";
for ($i = 0, $j = 10; $i < $j; $i += 4, $j -= 4) { echo "$i-$j "; }
echo "\n";
$n = 0;
for (;;) { if (++$n > 3) { break; } echo $n; }
echo "\n";
for ($k = 0; ; $k++) { if ($k == 2) { break; } echo $k; }
echo "\n";
class Reg {
    private static $inst;
    public $id = 0;
    public static function get() { return self::$inst ?? self::$inst = new self; }
}
Reg::get()->id = 7;
echo Reg::get()->id, "\n";
$arr = [];
$arr['a'] ?? $arr['a'] = 'set';
echo $arr['a'], "\n";
