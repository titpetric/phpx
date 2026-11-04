<?php
declare(strict_types=1);

use PHPUnit\Framework\Attributes\Depends;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

interface Shape { public function area(): float; }
abstract class Base implements Shape, JsonSerializable, Countable, IteratorAggregate
{
    protected array $items = [];
    public function count(): int { return count($this->items); }
    public function getIterator(): Iterator { return new ArrayIterator($this->items); }
    public function jsonSerialize(): mixed { return ['area' => $this->area()]; }
}
final class Rect extends Base
{
    public function __construct(private float $w, private float $h) { $this->items = [$w, $h]; }
    public function area(): float { return $this->w * $this->h; }
}
enum Level: int { case Low = 1; case High = 2; }
trait Greets { public function greet(): string { return 'hi ' . $this->name; } }
final class Person { use Greets; public function __construct(public string $name = 'x') {} }

final class AssertionsTest extends TestCase
{
    private array $log = [];

    protected function setUp(): void { $this->log[] = 'setUp'; }
    protected function tearDown(): void { $this->log[] = 'tearDown'; }

    public function testObjects(): void
    {
        $r = new Rect(2.0, 3.5);
        $this->assertSame(7.0, $r->area());
        $this->assertInstanceOf(Shape::class, $r);
        $this->assertCount(2, $r);
        $this->assertSame([2.0, 3.5], iterator_to_array($r));
        $this->assertJsonStringEqualsJsonString('{"area":7}', json_encode($r));
        $this->assertSame('hi bob', (new Person('bob'))->greet());
        $this->assertEquals(new Rect(1.0, 1.0), new Rect(1.0, 1.0));
        $this->assertNotEquals(new Rect(1.0, 1.0), new Rect(2.0, 1.0));
        $this->assertNotSame(new Rect(1.0, 1.0), new Rect(1.0, 1.0));
    }

    public function testEnums(): void
    {
        $this->assertSame(Level::High, Level::from(2));
        $this->assertSame('Low', Level::Low->name);
        $this->assertContains(Level::Low, Level::cases());
    }

    public function testStrings(): void
    {
        $this->assertSame('abc', 'a' . 'bc');
        $this->assertStringEndsWith('lo', 'hello');
        $this->assertStringNotContainsString('z', 'hello');
        $this->assertSame(5, strlen('hello'));
        $this->assertStringMatchesFormat('%d items', '12 items');
        $this->assertEqualsIgnoringCase('ABC', 'abc');
        $this->assertSame("a\nb", <<<EOT
a
b
EOT);
    }

    public function testArrays(): void
    {
        $this->assertSame(['a' => 1, 'b' => 2], ['a' => 1, 'b' => 2]);
        $this->assertEquals([1, 2, 3], [1, 2, 3]);
        $this->assertArrayHasKey('k', ['k' => null]);
        $this->assertArrayNotHasKey('z', ['k' => 1]);
        $this->assertContainsOnly('int', [1, 2, 3]);
        $this->assertContainsOnlyInstancesOf(Rect::class, [new Rect(1.0, 2.0)]);
        $this->assertSameSize([1, 2], ['a', 'b']);
        $this->assertIsList([1, 2]);
        $this->assertGreaterThanOrEqual(2, count([1, 2]));
        [$a, [$b, $c]] = [1, [2, 3]];
        $this->assertSame(6, $a + $b + $c);
    }

    public function testNumbers(): void
    {
        $this->assertEqualsWithDelta(0.3, 0.1 + 0.2, 0.0001);
        $this->assertLessThan(3, 2);
        $this->assertIsInt(3);
        $this->assertIsFloat(1.5);
        $this->assertIsString('s');
        $this->assertIsArray([]);
        $this->assertIsBool(true);
        $this->assertIsCallable('strlen');
        $this->assertIsNumeric('12.5');
        $this->assertNan(NAN);
        $this->assertInfinite(INF);
        $this->assertSame(3, intdiv(7, 2));
        $this->assertSame(1, 7 <=> 3);
    }

    public function testExceptions(): void
    {
        $this->expectException(DomainException::class);
        $this->expectExceptionCode(42);
        $this->expectExceptionMessageMatches('/^bad \d+$/');
        throw new DomainException('bad 7', 42);
    }

    public function testMatchAndClosures(): void
    {
        $f = fn(int $x): string => match (true) { $x < 0 => 'neg', $x === 0 => 'zero', default => 'pos' };
        $this->assertSame(['neg', 'zero', 'pos'], array_map($f, [-1, 0, 1]));
        $counter = function () { static $n = 0; return ++$n; };
        $this->assertSame(1, $counter());
        $this->assertSame(2, $counter());
        $this->assertSame(['x' => 1], (fn(...$args) => $args)(x: 1));
    }

    public function testFirst(): string
    {
        $this->assertSame(['setUp'], $this->log);
        return 'token';
    }

    #[Depends('testFirst')]
    public function testSecond(string $token): void
    {
        $this->assertSame('token', $token);
    }

    public static function pairs(): iterable
    {
        return [[1, 'one'], [2, 'two']];
    }

    #[DataProvider('pairs')]
    public function testPairs(int $n, string $word): void
    {
        $this->assertSame($word, [1 => 'one', 2 => 'two'][$n]);
    }

    public function testOutput(): void
    {
        $this->expectOutputString('printed');
        print 'printed';
    }

    public function testFailsOnPurpose(): void
    {
        $this->assertEquals(['a' => 1, 'b' => [1, 2]], ['a' => 1, 'b' => [1, 3]]);
    }

    public function testErrors(): void
    {
        throw new LogicException('boom');
    }
}
