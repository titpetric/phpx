<?php
$name = "World";
$arr = ['k' => 'v', 'n' => [1, 2]];
$obj = new stdClass;
$obj->p = 'prop';
echo <<<EOT
Hello $name!
  Indented "quoted" \$escaped \\ tab:\t|
Arr: $arr[k] {$arr['n'][1]} Obj: $obj->p {$obj->p}
EOT;
echo "\n";
echo <<<'EOT'
Raw $name {$x} \n "q"
EOT;
echo "\n";
echo <<<"EOT"
Dq $name
EOT . "|\n";
$s = <<<EOT
    indented body
      keeps relative
    EOT;
var_dump($s);
function f($a, $b) { return $a . $b; }
echo f(<<<A
first
A, <<<B
second
B), "\n";
class C {
    const T = <<<'X'
const body
X;
    public $p = 1;
}
echo C::T, "\n";
$empty = <<<E
E;
var_dump($empty);
echo <<<EOT
Line with \"escaped quote\" and "bare"
EOT, "\n";
