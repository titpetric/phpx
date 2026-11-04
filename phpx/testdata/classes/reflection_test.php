<?php
namespace App;
interface Shape { public function area(): float; }
abstract class Base implements Shape {
    const SIDES = 4;
    protected int $id = 0;
    public static $count = 0;
    abstract public function name(): string;
    public function area(): float { return 0.0; }
}
final class Sq extends Base {
    private string $secret = 'hidden';
    public function __construct(public float $w = 2.0, private ?string $label = null) {}
    public function name(): string { return 'sq'; }
    public static function make(float $w, string ...$rest): self { return new self($w); }
    private function hush(int $a, $b = 5) { return $a + $b; }
    public function testOne() {}
    public function testTwo() {}
}
$rc = new \ReflectionClass(Sq::class);
echo $rc->getName(), '|', $rc->getShortName(), '|', $rc->getNamespaceName(), "\n";
var_dump($rc->isFinal(), $rc->isAbstract(), $rc->isInstantiable(), $rc->isInterface());
echo $rc->getParentClass()->getName(), "\n";
echo implode(',', $rc->getInterfaceNames()), "\n";
foreach ($rc->getMethods() as $m) { echo $m->getName(), ':', $m->class, ' '; }
echo "\n";
$m = new \ReflectionMethod(Sq::class, 'hush');
var_dump($m->isPrivate(), $m->getNumberOfParameters(), $m->getNumberOfRequiredParameters());
foreach ($m->getParameters() as $p) { echo $p->getName(), $p->isOptional() ? '?' : '', ' '; }
echo "\n";
$o = new Sq(3.0);
echo $m->invoke($o, 1), "\n";
$p = new \ReflectionProperty(Sq::class, 'secret');
$p->setAccessible(true);
echo $p->getValue($o), "\n";
$p->setValue($o, 'changed');
echo $p->getValue($o), "\n";
$n = $rc->newInstanceWithoutConstructor();
var_dump($n instanceof Sq);
$i = $rc->newInstanceArgs([5.0]);
echo $i->w, "\n";
echo $rc->getMethod('make')->getReturnType()->getName(), "\n";
var_dump($rc->hasMethod('name'), $rc->hasMethod('nope'), $rc->hasProperty('secret'));
echo (new \ReflectionClass('App\Shape'))->isInterface() ? "iface\n" : "class\n";
try { new \ReflectionClass('Nope\Missing'); } catch (\ReflectionException $e) { echo $e->getMessage(), "\n"; }
$f = new \ReflectionFunction(fn(int $x, $y = 2) => $x + $y);
echo $f->getNumberOfParameters(), $f->invoke(1), "\n";
echo $rc->getConstant('SIDES'), "\n";
$rm = new \ReflectionMethod('App\Sq::testTwo');
echo $rm->name, "\n";
