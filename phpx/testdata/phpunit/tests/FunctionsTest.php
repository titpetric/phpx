<?php
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\TestCase;

final class FunctionsTest extends TestCase
{
    public static function slugs(): array
    {
        return ['simple' => ['Hello World', 'hello-world'], 'symbols' => ['A & B!', 'a-b'], 'spaces' => ['  x  y ', 'x-y'], 'empty' => ['', '']];
    }
    #[DataProvider('slugs')]
    public function testSlug(string $in, string $out): void
    {
        $s = trim(preg_replace('/[^a-z0-9]+/', '-', strtolower($in)), '-');
        $this->assertSame($out, $s);
    }
    public function testStrings(): void
    {
        $this->assertSame('abc', substr('xabcx', 1, 3));
        $this->assertSame(['a', 'b', 'c'], explode(',', 'a,b,c'));
        $this->assertSame('a-b-c', implode('-', ['a', 'b', 'c']));
        $this->assertSame('Hello', ucfirst('hello'));
        $this->assertSame('  x', str_pad('x', 3, ' ', STR_PAD_LEFT));
        $this->assertSame(3, substr_count('banana', 'an'));
        $this->assertSame('b', str_replace('a', 'b', 'a'));
        $this->assertSame('1,234.50', number_format(1234.5, 2));
        $this->assertSame('x=1', sprintf('%s=%d', 'x', 1));
        $this->assertTrue(str_starts_with('hello', 'he'));
        $this->assertSame(['x' => 1], json_decode('{"x":1}', true));
        $this->assertSame('{"a":[1,2]}', json_encode(['a' => [1, 2]]));
    }
    public function testArrays(): void
    {
        $this->assertSame([2, 4], array_values(array_filter([1, 2, 3, 4], fn($x) => $x % 2 === 0)));
        $this->assertSame([1, 4, 9], array_map(fn($x) => $x ** 2, [1, 2, 3]));
        $this->assertSame(10, array_sum([1, 2, 3, 4]));
        $this->assertSame(6, array_reduce([1, 2, 3], fn($c, $x) => $c + $x, 0));
        $this->assertSame(['a' => 1, 'b' => 2], array_merge(['a' => 1], ['b' => 2]));
        $this->assertSame([3, 2, 1], array_reverse([1, 2, 3]));
        $this->assertSame(['x', 'y'], array_keys(['x' => 1, 'y' => 2]));
        $this->assertSame(1, array_search('b', ['a', 'b']));
        $this->assertSame([[1, 2], [3]], array_chunk([1, 2, 3], 2));
        $a = [3, 1, 2]; usort($a, fn($x, $y) => $y <=> $x);
        $this->assertSame([3, 2, 1], $a);
        $this->assertSame(['b' => 1, 'a' => 2], array_flip([ 1 => 'b', 2 => 'a']));
        $this->assertSame([0 => 'a', 2 => 'c'], array_diff_key(['a', 'b', 'c'], [1 => 0]));
    }
    public function testMath(): void
    {
        $this->assertSame(3, (int) 3.9);
        $this->assertSame(4.0, round(3.6));
        $this->assertSame(8.0, floor(8.9));
        $this->assertSame(1, max(1, 0));
        $this->assertSame(2, abs(-2));
        $this->assertSame(1, 7 % 3);
        $this->assertEqualsWithDelta(1.4142, sqrt(2), 0.0001);
    }
    public function testReferencesAndStatics(): void
    {
        $f = function () { static $n = 0; return ++$n; };
        $f(); $this->assertSame(2, $f());
        $x = 1; $inc = function () use (&$x) { $x++; }; $inc(); $inc();
        $this->assertSame(3, $x);
    }
    public function testExceptionsFlow(): void
    {
        $log = [];
        try { try { throw new RuntimeException('a'); } finally { $log[] = 'inner'; } } catch (Exception $e) { $log[] = $e->getMessage(); } finally { $log[] = 'outer'; }
        $this->assertSame(['inner', 'a', 'outer'], $log);
        $this->expectExceptionMessage('deep');
        (function () { throw new LogicException('deep'); })();
    }
}
