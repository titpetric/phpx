<?php
namespace Foo\Bar;

use Other\Thing;
use Other\Thing as Aliased;

interface Marker {}
class K {
    public ?Marker $m = null;
    public function a(Marker $x, ?Thing $y, Aliased|int $z, \DateTime $d, self $s, string $str): ?Marker { return null; }
    public function b(): static { return $this; }
    public function c(): \Closure|null { return null; }
}
$m = new \ReflectionMethod(K::class, 'a');
foreach ($m->getParameters() as $p) { echo $p->getName(), ': ', $p->getType(), "\n"; }
echo $m->getReturnType(), "\n";
echo (new \ReflectionMethod(K::class, 'b'))->getReturnType(), "\n";
echo (new \ReflectionMethod(K::class, 'c'))->getReturnType(), "\n";
echo (new \ReflectionProperty(K::class, 'm'))->getType(), "\n";
