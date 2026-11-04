<?php
declare(strict_types=1);

use PHPUnit\Framework\Attributes\CoversNothing;
use PHPUnit\Framework\Attributes\DataProvider;
use PHPUnit\Framework\Attributes\Test;
use PHPUnit\Framework\Attributes\TestWith;
use PHPUnit\Framework\TestCase;

#[CoversNothing]
final class ProviderTest extends TestCase
{
    public static function additions(): array
    {
        return [
            'zeros'    => [0, 0, 0],
            'ones'     => [1, 1, 2],
            'negative' => [-1, 1, 0],
        ];
    }

    #[DataProvider('additions')]
    public function testAdd(int $a, int $b, int $expected): void
    {
        $this->assertSame($expected, $a + $b);
    }

    #[Test]
    public function addsWithoutThePrefix(): void
    {
        $this->assertTrue(true);
    }

    #[TestWith(['abc', 3])]
    #[TestWith(['', 0])]
    public function testLength(string $s, int $n): void
    {
        $this->assertSame($n, strlen($s));
    }

    public function testDependsOnNothing(): void
    {
        $this->assertSame('a', 'a');
    }
}
