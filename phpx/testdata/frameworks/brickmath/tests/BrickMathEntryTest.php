<?php
use PHPUnit\Framework\TestCase;
use Brick\Math\BigInteger;
use Brick\Math\BigDecimal;
use Brick\Math\RoundingMode;

final class BrickMathEntryTest extends TestCase
{
    public function testBigInteger(): void
    {
        $a = BigInteger::of('123456789012345678901234567890');
        $this->assertSame('246913578024691357802469135780', (string) $a->multipliedBy(2));
        $this->assertSame('123456789012345678901234567891', (string) $a->plus(1));
        $this->assertSame('1', (string) BigInteger::of(10)->power(0));
        $this->assertSame('1267650600228229401496703205376', (string) BigInteger::of(2)->power(100));
        $this->assertSame('6', (string) BigInteger::of(18)->gcd(12));
    }
    public function testDecimal(): void
    {
        $d = BigDecimal::of('1.10')->plus('2.205');
        $this->assertSame('3.305', (string) $d);
        $this->assertSame('3.31', (string) $d->toScale(2, RoundingMode::HALF_UP));
        $this->assertSame('0.3333', (string) BigDecimal::of(1)->dividedBy(3, 4, RoundingMode::DOWN));
    }
    public function testCompare(): void
    {
        $this->assertTrue(BigInteger::of('99999999999999999999')->isGreaterThan('9999999999999999999'));
        $this->assertSame(-1, BigInteger::of(1)->compareTo(2));
    }
}
