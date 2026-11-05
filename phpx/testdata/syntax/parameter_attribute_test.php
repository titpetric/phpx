<?php
#[Attribute]
class Marker { public function __construct(public string $v = '') {} }
function f(#[Marker('a]b')] $x, #[Marker] int $y = 2, #[Marker, Marker('x')] ...$rest) { return $x . $y . count($rest); }
class C {
    public function __construct(#[Marker] private string $s = 's', #[SensitiveParameter] public ?int $n = null) {}
    public function s() { return $this->s; }
}
echo f('p'), f('p', 5, 1, 2), (new C)->s(), "\n";
$g = function (#[Marker] $a) { return $a * 2; };
echo $g(4), "\n";
