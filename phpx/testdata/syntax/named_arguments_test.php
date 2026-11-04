<?php
function f($a, $b = 2, $c = 3) { return "$a-$b-$c"; }
echo f(1), "\n";
echo f(b: 5, a: 1), "\n";
echo f(1, c: 9), "\n";
echo f(c: 7, a: 0), "\n";
class P {
    public function __construct(public int $x = 1, public int $y = 2, public string $label = 'p') {}
    public function m(int $n = 1, string $s = 's') { return $n . $s; }
    public static function make(int $x = 0, int $y = 0) { return new self(y: $y, x: $x); }
}
$p = new P(y: 20);
echo $p->x, ' ', $p->y, ' ', $p->label, "\n";
echo $p->m(s: 'z'), "\n";
$q = P::make(y: 4);
echo $q->x, $q->y, "\n";
$cl = function ($u, $v = 'v') { return $u . $v; };
echo $cl(v: 'b', u: 'a'), "\n";
echo $cl(u: 'only'), "\n";
echo (fn($m, $n) => $m - $n)(n: 1, m: 10), "\n";
try {
    f(1, nope: 2);
} catch (Error $e) {
    echo get_class($e), ': ', $e->getMessage(), "\n";
}
$e = new Exception('boom', previous: new Exception('inner'));
echo $e->getMessage(), '/', $e->getPrevious()->getMessage(), "\n";
$e2 = new RuntimeException(code: 4, message: 'm');
echo $e2->getMessage(), $e2->getCode(), "\n";
echo str_pad(string: 'x', length: 3, pad_string: '-'), "\n";
echo implode(separator: ',', array: [1, 2]), "\n";
$a = [1, 2];
$b = ['k' => 'v'];
print_r([0, ...$a, ...$b, 9]);
function g(...$args) { return count($args); }
echo g(...$a, ...[3, 4]), "\n";
$collect = fn(...$args) => $args;
print_r($collect(1, 2, x: 3, y: 4));
function tail($first, ...$rest) { return $first . json_encode($rest); }
echo tail('a', 'b', k: 'v'), "\n";
var_dump(is_numeric('12.5'), is_numeric('1e3'), is_numeric(' 5'), is_numeric('abc'), is_numeric('0x1A'), is_numeric(null));
