<?php
use PHPUnit\Framework\TestCase;

interface Animal { public function speak(): string; }
trait Named { protected string $name = 'anon'; public function name(): string { return $this->name; } public static function make(string $n): static { $o = new static; $o->name = $n; return $o; } }
abstract class Base implements Animal { use Named; abstract public function legs(): int; public function describe(): string { return static::class . ':' . $this->name() . ':' . $this->legs() . ':' . $this->speak(); } }
final class Dog extends Base { public function speak(): string { return 'woof'; } public function legs(): int { return 4; } }
class Bird extends Base { public function speak(): string { return 'tweet'; } public function legs(): int { return 2; } }
final class Parrot extends Bird { public function speak(): string { return parent::speak() . '!'; } }
final class Counter { private static int $n = 0; public static function next(): int { return ++self::$n; } public function __construct(public readonly int $start = 5) {} }
final class Money implements Stringable { public function __construct(private int $cents) {} public function __toString(): string { return sprintf('%d.%02d', intdiv($this->cents, 100), $this->cents % 100); } }

final class OopTest extends TestCase
{
    public function testInheritanceAndTraits(): void
    {
        $this->assertSame('Dog:rex:4:woof', Dog::make('rex')->describe());
        $this->assertSame('Parrot:polly:2:tweet!', Parrot::make('polly')->describe());
        $this->assertInstanceOf(Animal::class, new Parrot);
        $this->assertInstanceOf(Bird::class, new Parrot);
        $this->assertSame('anon', (new Dog)->name());
    }
    public function testStaticAndReadonly(): void
    {
        $this->assertSame(1, Counter::next());
        $this->assertSame(2, Counter::next());
        $c = new Counter(9);
        $this->assertSame(9, $c->start);
        $this->expectException(Error::class);
        $c->start = 1;
    }
    public function testStringable(): void
    {
        $this->assertSame('12.05', (string) new Money(1205));
        $this->assertStringContainsString('3.00', 'cost ' . new Money(300));
    }
    public function testAbstractCannotInstantiate(): void
    {
        $this->expectException(Error::class);
        new Base;
    }
    public function testReflectionOnHierarchy(): void
    {
        $r = new ReflectionClass(Parrot::class);
        $this->assertSame('Bird', $r->getParentClass()->getName());
        $this->assertTrue($r->isFinal());
        $this->assertTrue($r->implementsInterface(Animal::class));
        $this->assertTrue($r->hasMethod('legs'));
    }
    public function testClosuresBindAndStatic(): void
    {
        $f = function () { return $this->start; };
        $this->assertSame(7, Closure::bind($f, new Counter(7), Counter::class)());
        $mul = fn(int $k) => fn(int $x) => $x * $k;
        $this->assertSame(21, $mul(3)(7));
    }
}
