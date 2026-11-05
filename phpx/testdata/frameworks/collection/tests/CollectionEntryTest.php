<?php
use PHPUnit\Framework\TestCase;
use Ramsey\Collection\Collection;
use Ramsey\Collection\Map\TypedMap;

final class CollectionEntryTest extends TestCase
{
    public function testTyped(): void
    {
        $c = new Collection('int', [1, 2, 3]);
        $c->add(4);
        $this->assertSame([1, 2, 3, 4], $c->toArray());
        $this->assertSame(4, count($c));
        $this->assertTrue($c->contains(3));
    }
    public function testTypeError(): void
    {
        $c = new Collection('string');
        $this->expectException(\Ramsey\Collection\Exception\InvalidArgumentException::class);
        $c->add(5);
    }
    public function testMap(): void
    {
        $m = new TypedMap('string', 'int', ['a' => 1]);
        $m['b'] = 2;
        $this->assertSame(['a' => 1, 'b' => 2], $m->toArray());
    }
    public function testSortAndFilter(): void
    {
        $c = new Collection('int', [3, 1, 2]);
        $this->assertSame([3], array_values($c->filter(fn($x) => $x > 2)->toArray()));
    }
}
