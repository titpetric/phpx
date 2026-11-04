<?php
declare(strict_types=1);

use App\Calc;
use PHPUnit\Framework\TestCase;

require_once __DIR__ . '/Calc.php';

final class CalcTest extends TestCase
{
    private Calc $calc;

    protected function setUp(): void
    {
        $this->calc = new Calc(1);
    }

    public function testAddition(): void
    {
        $this->assertSame(4, $this->calc->add(1, 2));
    }

    public function testDivision(): void
    {
        $this->assertEquals(2.5, $this->calc->div(5, 2));
        $this->assertEqualsWithDelta(0.333, $this->calc->div(1, 3), 0.001);
    }

    public function testDivisionByZeroThrows(): void
    {
        $this->expectException(InvalidArgumentException::class);
        $this->expectExceptionMessage('division by zero');
        $this->calc->div(1, 0);
    }

    public function testLog(): void
    {
        $this->calc->add(1, 1);
        $this->calc->add(2, 2);
        $this->assertCount(2, $this->calc->log());
        $this->assertContains('add', $this->calc->log());
        $this->assertSame(['add', 'add'], $this->calc->log());
    }

    public function testStringsAndArrays(): void
    {
        $this->assertStringContainsString('ell', 'hello');
        $this->assertStringStartsWith('he', 'hello');
        $this->assertArrayHasKey('a', ['a' => 1]);
        $this->assertNull(null);
        $this->assertTrue(true);
        $this->assertFalse(false);
        $this->assertEmpty([]);
        $this->assertNotEmpty([1]);
        $this->assertInstanceOf(Calc::class, $this->calc);
        $this->assertGreaterThan(1, 2);
        $this->assertMatchesRegularExpression('/^h/', 'hello');
        $this->assertEqualsCanonicalizing([3, 1, 2], [1, 2, 3]);
    }

    public function testFailure(): void
    {
        $this->assertSame(5, $this->calc->add(1, 1), 'deliberate failure');
    }

    public function testSkipped(): void
    {
        $this->markTestSkipped('not today');
    }
}
